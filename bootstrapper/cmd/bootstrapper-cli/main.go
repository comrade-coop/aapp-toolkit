package main

import (
    "fmt"
    "os"
)

func main() {
    fmt.Println("Bootstrapper CLI")
    if len(os.Args) < 2 {
        fmt.Println("Usage: bootstrapper-cli <command>")
        os.Exit(1)
    }
}
package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/edgelesssys/constellation/v2/internal/atls"
)

type InitRequest struct {
	GPGPrivateKey string `json:"gpgPrivateKey"`
}

func main() {
	var (
		serverAddr = flag.String("server", "localhost:8080", "bootstrapper service address")
		keyFile    = flag.String("key", "", "path to GPG private key file")
	)
	flag.Parse()

	if *keyFile == "" {
		log.Fatal("--key flag is required")
	}

	// Read GPG private key file
	keyData, err := ioutil.ReadFile(*keyFile)
	if err != nil {
		log.Fatalf("Failed to read key file: %v", err)
	}

	// Initialize ATLS client
	client, err := atls.NewClient()
	if err != nil {
		log.Fatalf("Failed to create ATLS client: %v", err)
	}

	// Create HTTP client with ATLS config
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				GetClientCertificate: client.GetClientCertificate,
				InsecureSkipVerify:   true, // Required for ATLS as cert is verified differently
			},
		},
	}

	// Prepare request payload
	reqBody := InitRequest{
		GPGPrivateKey: string(keyData),
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Fatalf("Failed to marshal request: %v", err)
	}

	// Send request to bootstrapper service
	url := fmt.Sprintf("https://%s/init", *serverAddr)
	resp, err := httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read and display response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Error: %s\n", body)
		os.Exit(1)
	}

	fmt.Println(string(body))
}
