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
		// For deployment, we need a certificate. 
		// We'll check if we have one in the 'certificates' table or if we need to issue.
		err = runSSHDeploy(domainID, accessConfig, wConfig)
	case "acme_http":
		err = runACMEIssue(domainID, domainHostname)
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

func runACMEIssue(domainID int, hostname string) error {
	// Simple email fetch from settings
	var email string
	database.DB.QueryRow("SELECT value FROM settings WHERE key = 'acme_email'").Scan(&email)
	if email == "" {
		email = "admin@example.com"
	}

	res, err := acme.IssueCertificate(email, hostname)
	if err != nil {
		return err
	}

	// Parse expiry from certificate
	expiry := "2099-01-01"
	block, _ := pem.Decode(res.Certificate)
	if block != nil {
		cert, err := x509.ParseCertificate(block.Bytes)
		if err == nil {
			expiry = cert.NotAfter.Format("2006-01-02 15:04:05")
		}
	}

	// Save to database
	_, err = database.DB.Exec(`
		INSERT INTO certificates (domain_id, fullchain, privkey, issuer, expiry)
		VALUES (?, ?, ?, ?, ?)`,
		domainID, string(res.Certificate), string(res.PrivateKey), res.IssuerCertificate, expiry)

	return err
}

func runSSHDeploy(domainID int, accessCfg, workflowCfg string) error {
	var sshCfg SSHConfig
	if err := json.Unmarshal([]byte(accessCfg), &sshCfg); err != nil {
		return err
	}

	var wCfg struct {
		RemotePath string `json:"remote_path"`
	}
	json.Unmarshal([]byte(workflowCfg), &wCfg)

	// Fetch certificate
	var fullchain, privkey string
	err := database.DB.QueryRow("SELECT fullchain, privkey FROM certificates WHERE domain_id = ? ORDER BY created_at DESC LIMIT 1", domainID).Scan(&fullchain, &privkey)
	if err != nil {
		return fmt.Errorf("no certificate found for domain. run ACME first")
	}

	// Create temp files to upload
	tempDir, _ := os.MkdirTemp("", "dominant-cert")
	defer os.RemoveAll(tempDir)

	certFile := filepath.Join(tempDir, "cert.pem")
	keyFile := filepath.Join(tempDir, "key.pem")
	os.WriteFile(certFile, []byte(fullchain), 0644)
	os.WriteFile(keyFile, []byte(privkey), 0644)

	return DeployViaSSH(sshCfg, certFile, keyFile, wCfg.RemotePath, "nginx -s reload")
}
