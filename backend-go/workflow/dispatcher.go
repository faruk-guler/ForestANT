package workflow

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"backend-go/acme"
	"backend-go/database"
)

func RunWorkflow(id int) {
	log.Printf("[Workflow] Starting workflow %d...", id)

	// 1. Fetch Workflow
	var wType, wConfig, domainHostname, accessConfig, accessProvider string
	var domainID, accessID int
	err := database.DB.QueryRow(`
		SELECT w.type, w.config, d.id, d.hostname, a.id, a.config, a.provider
		FROM workflows w
		JOIN domains d ON w.domain_id = d.id
		JOIN access a ON w.access_id = a.id
		WHERE w.id = ?`, id).Scan(&wType, &wConfig, &domainID, &domainHostname, &accessID, &accessConfig, &accessProvider)

	if err != nil {
		log.Printf("[Workflow] Error fetching workflow data: %v", err)
		return
	}

	database.LogScan(domainHostname, "workflow", "running", fmt.Sprintf("Started workflow: %s", wType))

	// 2. Execute Tasks
	switch wType {
	case "deploy_ssh":
		// FIX: hostname kullan, domainID değil (certificates tablosu domain TEXT tutuyor)
		err = runSSHDeploy(domainHostname, accessConfig, wConfig)
	case "acme_http":
		err = runACMEIssue(domainHostname)
	default:
		err = fmt.Errorf("unknown workflow type: %s", wType)
	}

	// 3. Update Status
	status := "success"
	msg := "Workflow completed successfully"
	if err != nil {
		status = "failed"
		msg = fmt.Sprintf("Workflow failed: %v", err)
	}

	database.DB.Exec("UPDATE workflows SET status = ?, last_run = CURRENT_TIMESTAMP WHERE id = ?", status, id)
	database.LogScan(domainHostname, "workflow", status, msg)
}

func runACMEIssue(hostname string) error {
	var email string
	database.DB.QueryRow("SELECT value FROM settings WHERE key = 'acme_email'").Scan(&email)
	// FIX: sessiz default yerine açık hata döndür
	if email == "" {
		return fmt.Errorf("acme_email is not configured in settings — please set it first")
	}

	res, err := acme.IssueCertificate(email, hostname)
	if err != nil {
		return err
	}

	// Parse expiry from certificate
	expiry := "2099-01-01"
	block, _ := pem.Decode(res.Certificate)
	if block != nil {
		cert, parseErr := x509.ParseCertificate(block.Bytes)
		if parseErr == nil {
			expiry = cert.NotAfter.Format("2006-01-02 15:04:05")
		}
	}

	// FIX: Doğru sütun adları: certificate, private_key, domain (TEXT)
	_, err = database.DB.Exec(`
		INSERT INTO certificates (domain, certificate, private_key, issuer, expiry)
		VALUES (?, ?, ?, ?, ?)`,
		hostname, string(res.Certificate), string(res.PrivateKey), string(res.IssuerCertificate), expiry)

	return err
}

func runSSHDeploy(hostname string, accessCfg, workflowCfg string) error {
	var sshCfg SSHConfig
	if err := json.Unmarshal([]byte(accessCfg), &sshCfg); err != nil {
		return fmt.Errorf("invalid access config JSON: %w", err)
	}

	var wCfg struct {
		RemotePath    string `json:"remote_path"`
		ReloadCommand string `json:"reload_command"` // FIX: artık config'den okunuyor
	}
	// Güvenli default'lar
	wCfg.RemotePath = "/etc/ssl/certs"
	wCfg.ReloadCommand = ""
	json.Unmarshal([]byte(workflowCfg), &wCfg)

	// FIX: Doğru sütun adları ve hostname ile sorgu (domain_id değil, domain TEXT)
	var fullchain, privkey string
	err := database.DB.QueryRow(
		"SELECT certificate, private_key FROM certificates WHERE domain = ? ORDER BY created_at DESC LIMIT 1",
		hostname,
	).Scan(&fullchain, &privkey)
	if err != nil {
		return fmt.Errorf("no certificate found for domain '%s' — run ACME first", hostname)
	}

	// Temp dosyalar oluştur
	tempDir, err := os.MkdirTemp("", "forestant-cert-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	certFile := filepath.Join(tempDir, "cert.pem")
	keyFile := filepath.Join(tempDir, "key.pem")

	if err := os.WriteFile(certFile, []byte(fullchain), 0644); err != nil {
		return fmt.Errorf("failed to write cert file: %w", err)
	}
	// FIX: Private key dosyası 0600 ile (sadece owner okuyabilir)
	if err := os.WriteFile(keyFile, []byte(privkey), 0600); err != nil {
		return fmt.Errorf("failed to write key file: %w", err)
	}

	return DeployViaSSH(sshCfg, certFile, keyFile, wCfg.RemotePath, wCfg.ReloadCommand)
}
