package goldmane

import (
	"errors"
	"testing"
)

func TestInvalidCertificateFilePathsReturnsErrInvalidX509KeyPairLoading(t *testing.T) {
	credential, err := NewTLSConfig("invalidPath", "InvalidPath", "InvalidKey")

	if !errors.Is(err, ErrInvalidX509KeyPairLoading) || credential != nil {
		t.Errorf("err should have been returned for invalid paths to certs and invalid private key")
	}
}
