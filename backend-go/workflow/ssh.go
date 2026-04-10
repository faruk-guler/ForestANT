package workflow

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

type SSHConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Key      string `json:"key"`
}

func DeployViaSSH(config SSHConfig, certPath, keyPath, remoteDest, reloadCmd string) error {
	var auth []ssh.AuthMethod

	if config.Key != "" {
		signer, err := ssh.ParsePrivateKey([]byte(config.Key))
		if err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}
		auth = append(auth, ssh.PublicKeys(signer))
	} else if config.Password != "" {
		auth = append(auth, ssh.Password(config.Password))
	}

	clientConfig := &ssh.ClientConfig{
		User:            config.Username,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // For demo, should use better host key check
	}

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	client, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	defer client.Close()

	// 1. Upload Certificate
	if err := uploadFile(client, certPath, filepath.Join(remoteDest, "fullchain.pem")); err != nil {
		return fmt.Errorf("cert upload failed: %w", err)
	}

	// 2. Upload Key
	if err := uploadFile(client, keyPath, filepath.Join(remoteDest, "privkey.pem")); err != nil {
		return fmt.Errorf("key upload failed: %w", err)
	}

	// 3. Run Reload Command
	if reloadCmd != "" {
		session, err := client.NewSession()
		if err != nil {
			return err
		}
		defer session.Close()

		if err := session.Run(reloadCmd); err != nil {
			return fmt.Errorf("reload cmd failed: %w", err)
		}
	}

	return nil
}

func uploadFile(client *ssh.Client, localPath, remotePath string) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	f, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer f.Close()

	stat, _ := f.Stat()

	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		fmt.Fprintf(w, "C%04o %d %s\n", 0644, stat.Size(), filepath.Base(remotePath))
		io.Copy(w, f)
		fmt.Fprint(w, "\x00")
	}()

	if err := session.Run(fmt.Sprintf("/usr/bin/scp -t %s", filepath.Dir(remotePath))); err != nil {
		return err
	}

	return nil
}
