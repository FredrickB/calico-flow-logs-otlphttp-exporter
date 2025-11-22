package goldmane

import (
	"crypto/tls"
	"crypto/x509"
	"os"

	"google.golang.org/grpc/credentials"
)

func newTLSConfig(caCertFilePath, publicCertPath, privateKeyPath string) (credentials.TransportCredentials, error) {
	clientCert, err := tls.LoadX509KeyPair(
		publicCertPath,
		privateKeyPath,
	)

	if err != nil {
		return nil, err
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
