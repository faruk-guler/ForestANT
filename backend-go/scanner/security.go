package scanner

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"
	"crypto/rsa"
	"crypto/ecdsa"
)

type SecurityReport struct {
	Grade      string   `json:"grade"`
	Score      int      `json:"score"`
	Reasons    []string `json:"reasons"`
	Protocol   string   `json:"protocol"`
	Cipher     string   `json:"cipher"`
}

func CalculateSecurityRating(hostname string) SecurityReport {
	report := SecurityReport{
		Grade:   "F",
		Score:   0,
		Reasons: []string{},
	}

	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, err := net.DialTimeout("tcp", hostname+":443", 5*time.Second)
	if err != nil {
		report.Reasons = append(report.Reasons, "Failed to connect to port 443")
		return report
	}
	defer conn.Close()

	tlsConn := tls.Client(conn, conf)
	err = tlsConn.Handshake()
	if err != nil {
		report.Reasons = append(report.Reasons, "TLS Handshake failed")
		return report
	}

	state := tlsConn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		report.Reasons = append(report.Reasons, "No certificates found")
		return report
	}

	cert := state.PeerCertificates[0]
	score := 100

	// 1. Protocol Version
	version := ""
	switch state.Version {
	case tls.VersionTLS13:
		version = "TLS 1.3"
		score += 5
	case tls.VersionTLS12:
		version = "TLS 1.2"
	case tls.VersionTLS11:
		version = "TLS 1.1"
		score -= 20
		report.Reasons = append(report.Reasons, "Insecure protocol: TLS 1.1")
	case tls.VersionTLS10:
		version = "TLS 1.0"
		score -= 40
		report.Reasons = append(report.Reasons, "Insecure protocol: TLS 1.0")
	default:
		version = "Unknown/Old"
		score -= 50
	}
	report.Protocol = version

	cipher := tls.CipherSuiteName(state.CipherSuite)
	report.Cipher = cipher
	
	// Check for weak ciphers
	weakCiphers := []string{"CBC", "RC4", "3DES", "DES", "MD5"}
	for _, weak := range weakCiphers {
		if strings.Contains(strings.ToUpper(cipher), weak) {
			score -= 20
			report.Reasons = append(report.Reasons, "Weak cipher suite detected: "+cipher)
			break
		}
	}

	// 3. Key Strength
	pubKey := cert.PublicKey
	bitLen := 0
	switch k := pubKey.(type) {
	case *rsa.PublicKey:
		bitLen = k.N.BitLen()
	case *ecdsa.PublicKey:
		bitLen = k.Curve.Params().BitSize
	}

	if bitLen > 0 {
		if bitLen < 2048 && bitLen != 256 && bitLen != 384 { // 256/384 are fine for ECDSA
			score -= 30
			report.Reasons = append(report.Reasons, fmt.Sprintf("Weak key length: %d bits", bitLen))
		}
	}

	// 4. Expiration
	daysLeft := int(time.Until(cert.NotAfter).Hours() / 24)
	if daysLeft < 0 {
		score -= 100
		report.Reasons = append(report.Reasons, "Certificate expired")
	} else if daysLeft < 30 {
		score -= 10
		report.Reasons = append(report.Reasons, "Expiring soon (< 30 days)")
	}

	// Calculate Grade
	if score >= 90 {
		report.Grade = "A"
	} else if score >= 80 {
		report.Grade = "B"
	} else if score >= 70 {
		report.Grade = "C"
	} else if score >= 60 {
		report.Grade = "D"
	} else {
		report.Grade = "F"
	}

	report.Score = score
	return report
}
