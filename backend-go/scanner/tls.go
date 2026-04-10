package scanner

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"
)

// TLSScanResult hem SSL expiry hem de security bilgisini tek bağlantıda döndürür
// FIX: Eskiden 2 ayrı TLS bağlantısı açılıyordu (GetSSLExpiry + CalculateSecurityRating)
// Şimdi tek bağlantıda her ikisi de alınıyor
type TLSScanResult struct {
	Expiry   *time.Time
	Security SecurityReport
}

type SecurityReport struct {
	Grade    string   `json:"grade"`
	Score    int      `json:"score"`
	Reasons  []string `json:"reasons"`
	Protocol string   `json:"protocol"`
	Cipher   string   `json:"cipher"`
}

// ScanTLS tek TLS bağlantısıyla hem expiry hem security döndürür
func ScanTLS(hostname string) TLSScanResult {
	result := TLSScanResult{
		Security: SecurityReport{
			Grade:   "F",
			Score:   0,
			Reasons: []string{},
		},
	}

	conf := &tls.Config{
		InsecureSkipVerify: true, // Expiry tespiti için gerekli (expired cert'lere de bağlanmalıyız)
	}

	dialer := &net.Dialer{Timeout: 5 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", hostname+":443", conf)
	if err != nil {
		result.Security.Reasons = append(result.Security.Reasons, "Failed to connect to port 443")
		return result
	}
	defer conn.Close()

	state := conn.ConnectionState()
	certs := state.PeerCertificates
	if len(certs) == 0 {
		result.Security.Reasons = append(result.Security.Reasons, "No certificates found")
		return result
	}

	cert := certs[0]

	// SSL Expiry
	exp := cert.NotAfter
	result.Expiry = &exp

	// Security Score
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
		result.Security.Reasons = append(result.Security.Reasons, "Insecure protocol: TLS 1.1")
	case tls.VersionTLS10:
		version = "TLS 1.0"
		score -= 40
		result.Security.Reasons = append(result.Security.Reasons, "Insecure protocol: TLS 1.0")
	default:
		version = "Unknown/Old"
		score -= 50
	}
	result.Security.Protocol = version

	// 2. Cipher Suite
	cipher := tls.CipherSuiteName(state.CipherSuite)
	result.Security.Cipher = cipher
	weakCiphers := []string{"CBC", "RC4", "3DES", "DES", "MD5"}
	for _, weak := range weakCiphers {
		if strings.Contains(strings.ToUpper(cipher), weak) {
			score -= 20
			result.Security.Reasons = append(result.Security.Reasons, "Weak cipher suite detected: "+cipher)
			break
		}
	}

	// 3. Key Strength
	bitLen := 0
	switch k := cert.PublicKey.(type) {
	case *rsa.PublicKey:
		bitLen = k.N.BitLen()
	case *ecdsa.PublicKey:
		bitLen = k.Curve.Params().BitSize
	}
	if bitLen > 0 && bitLen < 2048 && bitLen != 256 && bitLen != 384 {
		score -= 30
		result.Security.Reasons = append(result.Security.Reasons, fmt.Sprintf("Weak key length: %d bits", bitLen))
	}

	// 4. Expiration
	daysLeft := int(time.Until(cert.NotAfter).Hours() / 24)
	if daysLeft < 0 {
		score -= 100
		result.Security.Reasons = append(result.Security.Reasons, "Certificate expired")
	} else if daysLeft < 30 {
		score -= 10
		result.Security.Reasons = append(result.Security.Reasons, "Expiring soon (< 30 days)")
	}

	// Grade
	switch {
	case score >= 90:
		result.Security.Grade = "A"
	case score >= 80:
		result.Security.Grade = "B"
	case score >= 70:
		result.Security.Grade = "C"
	case score >= 60:
		result.Security.Grade = "D"
	default:
		result.Security.Grade = "F"
	}
	result.Security.Score = score

	return result
}

// GetSSLExpiry geriye dönük uyumluluk için (tek başına kullanım)
func GetSSLExpiry(hostname string) *time.Time {
	return ScanTLS(hostname).Expiry
}

// CalculateSecurityRating geriye dönük uyumluluk için
func CalculateSecurityRating(hostname string) SecurityReport {
	return ScanTLS(hostname).Security
}
