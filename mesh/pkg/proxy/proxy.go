package proxy

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/go-acme/lego/v4/certificate"
)

// IngressRule represents a routing rule
type IngressRule struct {
	Path     string
	Port     int
	IsMTLS   bool
	Location string // The backend location after stripping prefix
}

// Config holds the configuration for the proxy
type Config struct {
	Certificates *certificate.Resource
	Rules        []IngressRule
	MTLSRules    []IngressRule
}

// StartProxy initializes and starts the reverse proxy server
func StartProxy(config Config) error {
	// Create TLS config
	cert, err := tls.X509KeyPair(config.Certificates.Certificate, config.Certificates.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate pair: %v", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:  tls.VersionTLS12,
		GetConfigForClient: func(hello *tls.ClientHelloInfo) (*tls.Config, error) {
			// Check if the request path starts with /mtls/
			if strings.HasPrefix(hello.ServerName, "mtls.") {
				mtlsConfig := tlsConfig.Clone()
				mtlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
				return mtlsConfig, nil
			}
			return tlsConfig, nil
		},
	}

	// Combine all rules
	allRules := append([]IngressRule{}, config.Rules...)
	for _, rule := range config.MTLSRules {
		rule.IsMTLS = true
		allRules = append(allRules, rule)
	}

	// Create handler
	handler := createProxyHandler(allRules)

	// Wrap with mTLS check middleware
	combinedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Find matching rule
		var matchingRule *IngressRule
		for _, rule := range allRules {
			if strings.HasPrefix(r.URL.Path, rule.Path) {
				matchingRule = &rule
				break
			}
		}

		if matchingRule == nil {
			http.Error(w, "Path not found", http.StatusNotFound)
			return
		}

		// Check mTLS requirements
		if matchingRule.IsMTLS {
			if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
				http.Error(w, "Client certificate required", http.StatusUnauthorized)
				return
			}
		}

		handler.ServeHTTP(w, r)
	})

	// Start HTTPS server
	server := &http.Server{
		Addr:      ":443",
		Handler:   combinedHandler,
		TLSConfig: tlsConfig,
	}
	
	log.Printf("Starting HTTPS server on :443 (TLS and mTLS enabled)")
	return server.ListenAndServeTLS("", "")
}

func createProxyHandler(rules []IngressRule) http.Handler {
	mux := http.NewServeMux()
	
	for _, rule := range rules {
		rule := rule // Create new variable for closure
		targetURL, err := url.Parse(fmt.Sprintf("http://localhost:%d", rule.Port))
		if err != nil {
			log.Printf("Failed to parse target URL for path %s: %v", rule.Path, err)
			continue
		}

		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		
		// Modify the director to handle path rewriting
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			// Strip the prefix path and append any remaining path
			req.URL.Path = strings.TrimPrefix(req.URL.Path, rule.Path)
			if req.URL.Path == "" {
				req.URL.Path = "/"
			}
		}

		mux.HandleFunc(rule.Path, proxy.ServeHTTP)
		log.Printf("Set up proxy for path: %s -> http://localhost:%d (mTLS: %v)", 
			rule.Path, rule.Port, rule.IsMTLS)
	}

	return mux
}
