package main

import (
	"context"
	"log"
	"os"

	"github.com/FredrickB/calico-flow-logs-loki-exporter/v2/internal/goldmane"
)

const (
	CA_CERT_PATH_ENV     string = "CA_CERT_PATH"
	PRIVATE_KEY_PATH_ENV string = "PRIVATE_KEY_PATH"
	PUBLIC_CERT_PATH_ENV string = "PUBLIC_CERT_PATH"
	GOLDMANE_HOST_ENV    string = "GOLDMANE_HOST"
)

func main() {
	// Initialize logger
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Get path to certs required for Goldmane communication
	caCertFilePath, caCertSet := os.LookupEnv(CA_CERT_PATH_ENV)
	publicCertPath, publicCertSet := os.LookupEnv(PUBLIC_CERT_PATH_ENV)
	privateKeyPath, privateKeySet := os.LookupEnv(PRIVATE_KEY_PATH_ENV)
	goldmaneHost, goldmaneHostSet := os.LookupEnv(GOLDMANE_HOST_ENV)

	if !caCertSet || !privateKeySet || !publicCertSet || !goldmaneHostSet {
		log.Fatalf("One of the following environment variables is not set: %s, %s, %s, %s. All of these need to be set",
			CA_CERT_PATH_ENV, PRIVATE_KEY_PATH_ENV, PUBLIC_CERT_PATH_ENV, GOLDMANE_HOST_ENV)
	}

	log.Printf("%s set to %s", CA_CERT_PATH_ENV, caCertFilePath)
	log.Printf("%s set to %s", PRIVATE_KEY_PATH_ENV, privateKeyPath)
	log.Printf("%s set to %s", PUBLIC_CERT_PATH_ENV, publicCertPath)
	log.Printf("%s set to %s", GOLDMANE_HOST_ENV, goldmaneHost)

	client, err := goldmane.NewClient(
		goldmaneHost,
		caCertFilePath,
		publicCertPath,
		privateKeyPath,
	)
	if err != nil {
		log.Fatalf("Error while creating Goldmane client: %w", err)
	}

	// TODO: propagate to other systems, but for now, just print
	// the flows to the console.
	log.Println(client.GetFlow(context.Background()))

	defer client.Close()
}
