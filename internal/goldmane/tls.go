package goldmane

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"

	"google.golang.org/grpc/credentials"
)

var (
	ErrInvalidX509KeyPairLoading = errors.New("Could not load X509KeyPair")
)

func NewTLSConfig(caCertFilePath, publicCertPath, privateKeyPath string) (credentials.TransportCredentials, error) {
	clientCert, err := tls.LoadX509KeyPair(
		publicCertPath,
		privateKeyPath,
	)

	if err != nil {
		return nil, errors.Join(ErrInvalidX509KeyPairLoading, err)
	}

	caCert, err := os.ReadFile(caCertFilePath)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
		return nil, err
	}

	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      caCertPool,
	}), nil
}
