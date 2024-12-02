package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/registration"
	"go.mozilla.org/sops/v3/decrypt"
	"gopkg.in/yaml.v3"
)

// Manifest represents the structure of our YAML manifest
type Manifest struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Spec       struct {
		DNS struct {
			Zone     string `yaml:"zone"`
			Provider struct {
				Name string `yaml:"name"`
				Env  []struct {
					Name  string `yaml:"name"`
					Value string `yaml:"value"`
				} `yaml:"env"`
			} `yaml:"provider"`
		} `yaml:"dns"`
	} `yaml:"spec"`
}

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

func main() {
	// Path to the encrypted manifest
	manifestPath := "/etc/aapp-toolkit/aapp-manifest.yaml"

	// Read and decrypt the manifest
	decryptedData, err := decrypt.File(manifestPath, "yaml")
	if err != nil {
		log.Fatalf("Failed to decrypt manifest: %v", err)
	}

	// Parse the manifest
	var manifest Manifest
	if err := yaml.Unmarshal(decryptedData, &manifest); err != nil {
		log.Fatalf("Failed to parse manifest: %v", err)
	}

	// Extract domain from zone (remove wildcard if present)
	domain := strings.TrimPrefix(manifest.Spec.DNS.Zone, "*.")
	
	// Configure Cloudflare DNS provider
	var cfAPIKey, cfEmail string
	for _, env := range manifest.Spec.DNS.Provider.Env {
		switch env.Name {
		case "CF_API_KEY":
			cfAPIKey = env.Value
		case "CF_API_EMAIL":
			cfEmail = env.Value
		}
	}

	config := cloudflare.NewDefaultConfig()
	config.AuthToken = cfAPIKey
	config.AuthEmail = cfEmail
	
	provider, err := cloudflare.NewDNSProviderConfig(config)
	if err != nil {
		log.Fatalf("Failed to create Cloudflare provider: %v", err)
	}

	// Create user
	myUser := &User{
		Email: cfEmail,
		key:   []byte("dummy-key"), // In production, use proper key management
	}

	// Create config for Let's Encrypt
	config := lego.NewConfig(myUser)
	config.CADirURL = lego.LEDirectoryProduction // or lego.LEDirectoryStaging for testing

	// Create client
	client, err := lego.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Set DNS provider
	client.Challenge.SetDNS01Provider(provider,
		dns01.AddRecursiveNameservers([]string{"1.1.1.1:53"}),
		dns01.DisableCompletePropagationRequirement())

	// Request certificate
	request := certificate.ObtainRequest{
		Domains: []string{"*." + domain},
		Bundle:  true,
	}

	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		log.Fatalf("Failed to obtain certificate: %v", err)
	}

	log.Printf("Successfully obtained certificate for %s", domain)
	log.Printf("Certificate valid until: %s", certificates.NotAfter.Format(time.RFC3339))

	// Keep the service running
	select {}
}
