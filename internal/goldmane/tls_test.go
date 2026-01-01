package goldmane

import (
	"errors"
	"os"
	"testing"

	"github.com/mdelapenya/tlscert"
)

func TestInvalidCertificateFilePathsReturnsErrInvalidX509KeyPairLoading(t *testing.T) {
	credential, err := NewTLSConfig("invalidPath", "InvalidPath", "InvalidKey")

	if !errors.Is(err, ErrInvalidX509KeyPairLoading) || credential != nil {
		t.Errorf("err should have been returned for invalid paths to certs and invalid private key")
	}
}

func TestCorrectCreationOfTLS(t *testing.T) {
	caCert := tlscert.SelfSignedFromRequest(tlscert.Request{
		Host:      "localhost",
		Name:      "ca-cert",
		ParentDir: os.TempDir(),
	})
	if caCert == nil {
		t.Errorf("Failed to generate CA certificate")
	}
	defer os.Remove(caCert.CertPath)
	defer os.Remove(caCert.KeyPath)

	cert := tlscert.SelfSignedFromRequest(tlscert.Request{
		Host:      "localhost",
		Name:      "client-cert",
		Parent:    caCert,
		ParentDir: os.TempDir(),
	})
	if cert == nil {
		t.Errorf("Failed to generate certificate")
	}
	defer os.Remove(cert.CertPath)
	defer os.Remove(cert.KeyPath)

	credential, err := NewTLSConfig(caCert.CertPath, cert.CertPath, cert.KeyPath)

	if err != nil {
		t.Errorf("expected error to be nil, but was %v", err)
	}
	if credential == nil {
		t.Error("credential should not be nil")
	}
}
