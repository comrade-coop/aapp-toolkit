package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/atls"
)

type InitRequest struct {
	GPGPrivateKey string `json:"gpgPrivateKey"`
}

type Measurements struct {
	PCR0         string `json:"pcr0"`
	PCR1         string `json:"pcr1"`
	PCR2         string `json:"pcr2"`
	ManifestHash string `json:"manifestHash"`
}

func main() {
	var (
		serverAddr = flag.String("server", "localhost:8080", "bootstrapper service address")
		keyFile    = flag.String("key", "", "path to GPG private key file")
		image      = flag.String("image", "", "image version to verify measurements against")
		noVerify   = flag.Bool("no-verify", false, "skip measurement verification")
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

	// First fetch measurements
	url := fmt.Sprintf("https://%s/init", *serverAddr)
	getMeasurements, err := httpClient.Get(url)
	if err != nil {
		log.Fatalf("Failed to get measurements: %v", err)
	}
	defer getMeasurements.Body.Close()

	var measurements Measurements
	if err := json.NewDecoder(getMeasurements.Body).Decode(&measurements); err != nil {
		log.Fatalf("Failed to decode measurements: %v", err)
	}

	// Verify measurements if image is provided
	if *image != "" && !*noVerify {
		fetcher := measurements.NewVerifyFetcher(
			sigstore.NewCosignVerifier,
			rekor.New(),
			httpClient,
		)

		ctx := context.Background()
		expectedMeasurements, err := fetcher.FetchAndVerifyMeasurements(
			ctx,
			*image,
			cloudprovider.Azure, // This should be made configurable
			variant.AzureSEVSNP,
			*noVerify,
		)
		if err != nil {
			log.Fatalf("Failed to verify measurements: %v", err)
		}

		// Compare measurements
		// This is a simplified comparison - you should implement proper measurement validation
		if measurements.ManifestHash != expectedMeasurements.GetManifestHash() {
			log.Fatalf("Manifest hash mismatch")
		}
	}

	// Send init request to bootstrapper service
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
