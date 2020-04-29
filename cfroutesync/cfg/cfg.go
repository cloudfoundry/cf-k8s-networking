package cfg

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	UAA struct {
		// Base URL for UAA, e.g. uaa.sys.example.com or uaa.cf.system.internal
		BaseURL string

		// UAA client name to use when acquiring a token for accessing Cloud Controller
		ClientName string

		// Client secret matching the client name
		ClientSecret string

		// Certificate authority that signed the UAA server cert
		CA *x509.CertPool
	}

	CC struct {
		// Base URL for Cloud Controller, e.g. api.sys.example.com or api.cf.system.internal
		BaseURL string

		// Certificate authority that signed the Cloud Controller server cert
		CA *x509.CertPool
	}

	Istio struct {
		// List of Istio Gateway names to use for workload ingress
		Gateways []string
	}
}

const (
	FileUAABaseURL      = "uaaBaseURL"
	FileUAAClientName   = "clientName"
	FileUAAClientSecret = "clientSecret"
	FileUAACA           = "uaaCA"
	FileCCBaseURL       = "ccBaseURL"
	FileCCCA            = "ccCA"
)

// Load loads a Config from environment variables or files within a directory on disk
// When running inside a K8s Cluster, this directory should probably be a volume mount of a K8s Secret
func Load(configDir string) (*Config, error) {
	ccBaseUrl, err := loadValue(configDir, FileCCBaseURL)
	if err != nil {
		return nil, err
	}
	uaaBaseURL, err := loadValue(configDir, FileUAABaseURL)
	if err != nil {
		return nil, err
	}
	clientName, err := loadValue(configDir, FileUAAClientName)
	if err != nil {
		return nil, err
	}
	clientSecret, err := loadValue(configDir, FileUAAClientSecret)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	uaaCA, err := loadCert(configDir, FileUAACA)
	if err != nil {
		return nil, err
	}

	ccCA, err := loadCert(configDir, FileCCCA)
	if err != nil {
		return nil, err
	}

	c := &Config{}
	c.UAA.BaseURL = uaaBaseURL
	c.UAA.ClientName = clientName
	c.UAA.ClientSecret = clientSecret
	c.UAA.CA = uaaCA
	c.CC.BaseURL = ccBaseUrl
	c.CC.CA = ccCA
	c.Istio.Gateways = []string{"istio-ingress"}
	return c, nil
}

func loadCert(configDir string, key string) (*x509.CertPool, error) {
	caContent, err := loadValue(configDir, key)

	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM([]byte(caContent)); !ok {
		return nil, fmt.Errorf("unable to load CA certificate")
	}
	return caCertPool, nil
}

func loadValue(configDir string, key string) (string, error) {
	value, exists := os.LookupEnv(key)
	if exists {
		return value, nil
	}
	return readFile(configDir, key)
}

func readFile(configDir string, filename string) (string, error) {
	bytes, err := ioutil.ReadFile(getPath(configDir, filename))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(bytes)), nil
}

func getPath(configDir string, filename string) string {
	return filepath.Join(configDir, filename)
}
