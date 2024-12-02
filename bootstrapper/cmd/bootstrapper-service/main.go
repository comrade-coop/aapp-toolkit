package main

import (
    "context"
    "crypto/tls"
    "encoding/json"
    "fmt"
    "io"
    "io/ioutil"
    "log"
    "net/http"
    "os/exec"
    "crypto/sha256"
    "encoding/hex"
    
    "github.com/edgelesssys/constellation/v2/image/measured-boot/measure"
    "github.com/edgelesssys/constellation/v2/internal/atls"
    "github.com/gorilla/mux"
)

func main() {
    // Get measurements file path from environment
    measurementsPath := os.Getenv("MEASUREMENTS_FILE")
    if measurementsPath == "" {
        measurementsPath = "/pcrs.json" // default path
    }

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
            // Read measurements from file
            measurementsData, err := ioutil.ReadFile(measurementsPath)
            if err != nil {
                log.Printf("Failed to read measurements file: %v", err)
                http.Error(w, "Failed to read measurements", http.StatusInternalServerError)
                return
            }

            var measurements Measurements
            if err := json.Unmarshal(measurementsData, &measurements); err != nil {
                log.Printf("Failed to parse measurements: %v", err)
                http.Error(w, "Failed to parse measurements", http.StatusInternalServerError)
                return
            }

            // Read and hash the AAP manifest
            manifestData, err := ioutil.ReadFile("/etc/aapp-toolkit/aap-manifest.yaml")
            if err != nil {
                log.Printf("Failed to read AAP manifest: %v", err)
                http.Error(w, "Failed to read AAP manifest", http.StatusInternalServerError)
                return
            }

            // Calculate SHA-256 hash of manifest
            hasher := sha256.New()
            hasher.Write(manifestData)
            measurements.ManifestHash = hex.EncodeToString(hasher.Sum(nil))

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

        // Validate the imported key can decrypt SOPS-encrypted manifest
        cmd = exec.Command("sops", "--decrypt", "/etc/aapp-toolkit/aap-manifest.yaml")
        output, err := cmd.CombinedOutput()
        if err != nil {
            log.Printf("SOPS validation failed: %v\nOutput: %s", err, output)
            http.Error(w, "Invalid GPG key - cannot decrypt manifest", http.StatusBadRequest)
            return
        }

        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, "GPG key imported and validated successfully")
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
