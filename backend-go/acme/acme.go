package acme

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"log"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
)

var ActiveChallenges = make(map[string]string)

type HTTP01Provider struct{}

func (p *HTTP01Provider) Present(domain, token, keyAuth string) error {
	ActiveChallenges[token] = keyAuth
	return nil
}

func (p *HTTP01Provider) CleanUp(domain, token, keyAuth string) error {
	delete(ActiveChallenges, token)
	return nil
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

	// Use Let's Encrypt Staging by default for safety in this demo
	// In production, this should be configurable
	config.CADirURL = "https://acme-staging-v02.api.letsencrypt.org/directory"

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

	// 3. Setup Challenges
	err = client.Challenge.SetHTTP01Provider(&HTTP01Provider{})
	if err != nil {
		return nil, err
	}
	
	log.Printf("[ACME] Requesting certificate for %s...", domain)
	
	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}
	
	// This will fail right now because no challenges are solved.
	// We need to implement the challenge satisfaction logic.
	return client.Certificate.Obtain(request)
}
