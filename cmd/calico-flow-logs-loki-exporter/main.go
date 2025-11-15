package main

import (
	"log"
	"os"

	"github.com/FredrickB/calico-flow-logs-loki-exporter/v2/internal/goldmane"
)

const (
	CA_CERT_PATH_ENV     string = "CA_CERT_PATH"
	PRIVATE_KEY_PATH_ENV string = "PRIVATE_KEY_PATH"
	PUBLIC_CERT_PATH_ENV string = "PUBLIC_CERT_PATH"
)

func main() {
	// Initialize logger
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Get path to certs required for Goldmane communication
	caCert, caCertSet := os.LookupEnv(CA_CERT_PATH_ENV)
	privateKey, privateKeySet := os.LookupEnv(PRIVATE_KEY_PATH_ENV)
	publicCert, publicCertSet := os.LookupEnv(PUBLIC_CERT_PATH_ENV)

	if !caCertSet || !privateKeySet || !publicCertSet {
		log.Fatalf("One of the following environment variables is not set: %s, %s, %s. All of these need to be set",
			CA_CERT_PATH_ENV, PRIVATE_KEY_PATH_ENV, PUBLIC_CERT_PATH_ENV)
	}

	log.Printf("%s set to %s", CA_CERT_PATH_ENV, caCert)
	log.Printf("%s set to %s", PRIVATE_KEY_PATH_ENV, privateKey)
	log.Printf("%s set to %s", PUBLIC_CERT_PATH_ENV, publicCert)

	goldmane.NewClient()
}
