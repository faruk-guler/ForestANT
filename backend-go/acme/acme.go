package acme

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"log"
	"os"
	"sync"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
)

// FIX: Race condition önlemek için mutex ile koruma
var (
	challengeMu     sync.RWMutex
	ActiveChallenges = make(map[string]string)
)

type HTTP01Provider struct{}

func (p *HTTP01Provider) Present(domain, token, keyAuth string) error {
	challengeMu.Lock()
	defer challengeMu.Unlock()
	ActiveChallenges[token] = keyAuth
	return nil
}

func (p *HTTP01Provider) CleanUp(domain, token, keyAuth string) error {
	challengeMu.Lock()
	defer challengeMu.Unlock()
	delete(ActiveChallenges, token)
	return nil
}

// GetChallenge thread-safe challenge okuma (handlers.go için)
func GetChallenge(token string) (string, bool) {
	challengeMu.RLock()
	defer challengeMu.RUnlock()
	v, ok := ActiveChallenges[token]
	return v, ok
}

// User represents a lego user
type User struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *User) GetEmail() string {
	return u.Email
}
func (u *User) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *User) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func IssueCertificate(email, domain string) (*certificate.Resource, error) {
	// 1. Create a user
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	myUser := User{
		Email: email,
		key:   privateKey,
	}

	config := lego.NewConfig(&myUser)

	// FIX: ACME_MODE env'den okunuyor
	acmeMode := os.Getenv("ACME_MODE")
	if acmeMode == "production" {
		config.CADirURL = "https://acme-v02.api.letsencrypt.org/directory"
		log.Printf("[ACME] Using PRODUCTION Let's Encrypt for %s", domain)
	} else {
		config.CADirURL = "https://acme-staging-v02.api.letsencrypt.org/directory"
		log.Printf("[ACME] Using STAGING Let's Encrypt for %s (set ACME_MODE=production for real certs)", domain)
	}

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, err
	}

	// 2. Register
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, err
	}
	myUser.Registration = reg

	// 3. Setup HTTP-01 Challenge
	err = client.Challenge.SetHTTP01Provider(&HTTP01Provider{})
	if err != nil {
		return nil, err
	}

	log.Printf("[ACME] Requesting certificate for %s...", domain)

	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}

	return client.Certificate.Obtain(request)
}
