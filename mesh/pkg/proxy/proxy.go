package proxy

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/go-acme/lego/v4/certificate"
)

// IngressRule represents a routing rule
type IngressRule struct {
	Path string
	Port int
}

// Config holds the configuration for the proxy
type Config struct {
	Certificates *certificate.Resource
	Rules        []IngressRule
	MTLSRules    []IngressRule
}

// StartProxy initializes and starts the reverse proxy servers
func StartProxy(config Config) error {
	// Create TLS config
	cert, err := tls.X509KeyPair(config.Certificates.Certificate, config.Certificates.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate pair: %v", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:  tls.VersionTLS12,
	}

	// Create proxy handlers
	proxyHandler := createProxyHandler(config.Rules)
	mtlsHandler := createProxyHandler(config.MTLSRules)

	// Start regular HTTPS server
	go func() {
		server := &http.Server{
			Addr:      ":443",
			Handler:   proxyHandler,
			TLSConfig: tlsConfig,
		}
		log.Printf("Starting HTTPS server on :443")
		if err := server.ListenAndServeTLS("", ""); err != nil {
			log.Printf("HTTPS server failed: %v", err)
		}
	}()

	// Start mTLS HTTPS server
	mtlsConfig := tlsConfig.Clone()
	mtlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	server := &http.Server{
		Addr:      ":8443",
		Handler:   mtlsHandler,
		TLSConfig: mtlsConfig,
	}
	log.Printf("Starting mTLS HTTPS server on :8443")
	return server.ListenAndServeTLS("", "")
}

func createProxyHandler(rules []IngressRule) http.Handler {
	mux := http.NewServeMux()
	
	for _, rule := range rules {
		targetURL, err := url.Parse(fmt.Sprintf("http://localhost:%d", rule.Port))
		if err != nil {
			log.Printf("Failed to parse target URL for path %s: %v", rule.Path, err)
			continue
		}

		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		mux.HandleFunc(rule.Path, proxy.ServeHTTP)
		log.Printf("Set up proxy for path: %s -> %s", rule.Path, targetURL.String())
	}

	return mux
}
