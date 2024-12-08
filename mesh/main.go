package main

import (
	"log"
	"strings"

	"your-repo-path/mesh/pkg/certmanager"
	"your-repo-path/mesh/pkg/proxy"
	"go.mozilla.org/sops/v3/decrypt"
	"gopkg.in/yaml.v3"
)

// Manifest represents the structure of our YAML manifest
type Manifest struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Spec       struct {
		Container struct {
			Image string `yaml:"image"`
		} `yaml:"container"`
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
		Ingress struct {
			Rules []struct {
				HTTP struct {
					Paths []struct {
						Path    string `yaml:"path"`
						Backend struct {
							Service struct {
								Port struct {
									Number int `yaml:"number"`
								} `yaml:"port"`
							} `yaml:"service"`
						} `yaml:"backend"`
					} `yaml:"paths"`
				} `yaml:"http"`
			} `yaml:"rules"`
		} `yaml:"ingress"`
		MTLSIngress struct {
			Rules []struct {
				HTTP struct {
					Paths []struct {
						Path    string `yaml:"path"`
						Backend struct {
							Service struct {
								Port struct {
									Number int `yaml:"number"`
								} `yaml:"port"`
							} `yaml:"service"`
						} `yaml:"backend"`
					} `yaml:"paths"`
				} `yaml:"http"`
			} `yaml:"rules"`
		} `yaml:"mtlsIngress"`
	} `yaml:"spec"`
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
	
	// Get Cloudflare credentials
	var cfAPIKey, cfEmail string
	for _, env := range manifest.Spec.DNS.Provider.Env {
		switch env.Name {
		case "CF_API_KEY":
			cfAPIKey = env.Value
		case "CF_API_EMAIL":
			cfEmail = env.Value
		}
	}

	// Get certificate
	certConfig := certmanager.Config{
		Domain:   domain,
		APIKey:   cfAPIKey,
		APIEmail: cfEmail,
	}
	
	certificates, err := certmanager.ObtainCertificate(certConfig)
	if err != nil {
		log.Fatalf("Failed to obtain certificate: %v", err)
	}

	// Prepare proxy rules
	var rules, mtlsRules []proxy.IngressRule
	
	for _, rule := range manifest.Spec.Ingress.Rules {
		for _, path := range rule.HTTP.Paths {
			rules = append(rules, proxy.IngressRule{
				Path: path.Path,
				Port: path.Backend.Service.Port.Number,
				IsMTLS: false,
			})
		}
	}

	for _, rule := range manifest.Spec.MTLSIngress.Rules {
		for _, path := range rule.HTTP.Paths {
			mtlsRules = append(mtlsRules, proxy.IngressRule{
				Path: path.Path,
				Port: path.Backend.Service.Port.Number,
				IsMTLS: true,
			})
		}
	}

	// Start proxy
	proxyConfig := proxy.Config{
		Certificates: certificates,
		Rules:       rules,
		MTLSRules:   mtlsRules,
	}

	if err := proxy.StartProxy(proxyConfig); err != nil {
		log.Fatalf("Proxy server failed: %v", err)
	}
}
