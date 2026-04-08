package scanner

import (
	"crypto/tls"
	"net"
	"time"
)

func GetSSLExpiry(hostname string) *time.Time {
	config := &tls.Config{InsecureSkipVerify: true}
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	
	conn, err := tls.DialWithDialer(dialer, "tcp", hostname+":443", config)
	if err != nil {
		return nil
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) > 0 {
		exp := certs[0].NotAfter
		return &exp
	}

	return nil
}
