package certmanager

import (
	"log"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/registration"
)

// User implements acme.User
type User struct {
	Email        string
	Registration *registration.Resource
	key          []byte
}

func (u *User) GetEmail() string {
	return u.Email
}
func (u *User) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *User) GetPrivateKey() []byte {
	return u.key
}

// Config holds the configuration for certificate management
type Config struct {
	Domain   string
	APIKey   string
	APIEmail string
}

// ObtainCertificate gets a certificate from Let's Encrypt using DNS challenge
func ObtainCertificate(config Config) (*certificate.Resource, error) {
	// Configure Cloudflare DNS provider
	cfConfig := cloudflare.NewDefaultConfig()
	cfConfig.AuthToken = config.APIKey
	cfConfig.AuthEmail = config.APIEmail
	
	provider, err := cloudflare.NewDNSProviderConfig(cfConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloudflare provider: %v", err)
	}

	// Create user
	myUser := &User{
		Email: config.APIEmail,
		key:   []byte("dummy-key"), // In production, use proper key management
	}

	// Create config for Let's Encrypt
	leConfig := lego.NewConfig(myUser)
	leConfig.CADirURL = lego.LEDirectoryProduction

	// Create client
	client, err := lego.NewClient(leConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	// Set DNS provider
	client.Challenge.SetDNS01Provider(provider,
		dns01.AddRecursiveNameservers([]string{"1.1.1.1:53"}),
		dns01.DisableCompletePropagationRequirement())

	// Request certificate
	request := certificate.ObtainRequest{
		Domains: []string{"*." + config.Domain},
		Bundle:  true,
	}

	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain certificate: %v", err)
	}

	log.Printf("Successfully obtained certificate for %s", config.Domain)
	log.Printf("Certificate valid until: %s", certificates.NotAfter.Format(time.RFC3339))

	return certificates, nil
}
