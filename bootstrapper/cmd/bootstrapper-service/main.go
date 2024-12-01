package main

import (
    "context"
    "crypto/sha256"
    "crypto/tls"
    "encoding/json"
    "fmt"
    "io"
    "io/ioutil"
    "log"
    "net/http"
    "os/exec"
    
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
    
    type InitRequest struct {
        GPGPrivateKey string `json:"gpgPrivateKey"`
    }

    type Measurements struct {
        PCR0        string `json:"pcr0"`
        PCR1        string `json:"pcr1"`
        PCR2        string `json:"pcr2"`
        ManifestHash string `json:"manifestHash"`
    }

    r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "OK")
    })

    r.HandleFunc("/init", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
            // Read manifest file
            manifestData, err := ioutil.ReadFile("/etc/aapp-toolkit/aapp-manifest.yaml")
            if err != nil {
                log.Printf("Failed to read manifest file: %v", err)
                http.Error(w, "Failed to read manifest", http.StatusInternalServerError)
                return
            }

            // Calculate manifest hash
            manifestHash := fmt.Sprintf("%x", sha256.Sum256(manifestData))

            // In a real implementation, these would be read from the TPM
            // For now, using placeholder values
            measurements := Measurements{
                PCR0:        "0000000000000000000000000000000000000000000000000000000000000000",
                PCR1:        "0000000000000000000000000000000000000000000000000000000000000001",
                PCR2:        "0000000000000000000000000000000000000000000000000000000000000002",
                ManifestHash: manifestHash,
            }

            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(measurements)
            return

        case http.MethodPost:

        // Read request body
        var req InitRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Invalid request body", http.StatusBadRequest)
            return
        }

        // Import GPG key using gpg command
        cmd := exec.Command("gpg", "--import")
        stdin, err := cmd.StdinPipe()
        if err != nil {
            log.Printf("Failed to create stdin pipe: %v", err)
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }

        // Start the command before writing to stdin
        if err := cmd.Start(); err != nil {
            log.Printf("Failed to start gpg command: %v", err)
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }

        // Write the private key to gpg's stdin
        if _, err := io.WriteString(stdin, req.GPGPrivateKey); err != nil {
            log.Printf("Failed to write to stdin: %v", err)
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }
        stdin.Close()

        // Wait for the command to complete
        if err := cmd.Wait(); err != nil {
            log.Printf("GPG import failed: %v", err)
            http.Error(w, "Failed to import GPG key", http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, "GPG key imported successfully")
        }
    })
    
    server := &http.Server{
        Addr:      ":8080",
        Handler:   r,
        TLSConfig: tlsConfig,
    }
    
    log.Println("Starting ATLS server on :8080")
    log.Fatal(server.ListenAndServeTLS("", "")) // Empty strings since cert/key are handled by ATLS
}
