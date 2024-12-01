package main

import (
    "context"
    "crypto/tls"
    "fmt"
    "log"
    "net/http"
    
    "github.com/edgelesssys/constellation/v2/internal/atls"
    "github.com/gorilla/mux"
)

func main() {
    // Initialize ATLS issuer
    issuer, err := atls.NewIssuer()
    if err != nil {
        log.Fatalf("Failed to create ATLS issuer: %v", err)
    }

    // Create TLS config with ATLS
    tlsConfig := &tls.Config{
        GetConfigForClient: issuer.GetConfigForClient,
        ClientAuth:         tls.RequireAnyClientCert,
    }

    r := mux.NewRouter()
    
    r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "OK")
    })
    
    server := &http.Server{
        Addr:      ":8080",
        Handler:   r,
        TLSConfig: tlsConfig,
    }
    
    log.Println("Starting ATLS server on :8080")
    log.Fatal(server.ListenAndServeTLS("", "")) // Empty strings since cert/key are handled by ATLS
}
