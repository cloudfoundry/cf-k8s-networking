package cfg

import (
	"io/ioutil"
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

		// PEM file path for the certificate authority that signed the UAA server cert
		CAFile string
	}

	CC struct {
		// Base URL for Cloud Controller, e.g. api.sys.example.com or api.cf.system.internal
		BaseURL string

		// PEM file path for the certificate authority that signed the CC server cert
		CAFile string
	}

	Istio struct {
		// List of Istio Gateway names to use for workload ingress
		Gateways []string
	}
}

const (
	FileUAABaseURL      = "uaaBaseUrl"
	FileUAAClientName   = "clientName"
	FileUAAClientSecret = "clientSecret"
	FileUAACA           = "ca"
	FileCCBaseURL       = "ccBaseUrl"
	FileCCCA            = "ca" // currently we expect 1 CA for both UAA and CC
)

// FromDir loads a Config from files within a directory on disk
// When running inside a K8s Cluster, this directory should probably be a volume mount of a K8s Secret
func FromDir(configDir string) (*Config, error) {
	getPath := func(filename string) string { return filepath.Join(configDir, filename) }
	readFile := func(filename string) (string, error) {
		bytes, err := ioutil.ReadFile(getPath(filename))
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(bytes)), nil
	}
	uaaBaseUrl, err := readFile(FileUAABaseURL)
	if err != nil {
		return nil, err
	}
	clientName, err := readFile(FileUAAClientName)
	if err != nil {
		return nil, err
	}
	clientSecret, err := readFile(FileUAAClientSecret)
	if err != nil {
		return nil, err
	}

	ccBaseUrl, err := readFile(FileCCBaseURL)
	if err != nil {
		return nil, err
	}

	c := &Config{}
	c.UAA.BaseURL = uaaBaseUrl
	c.UAA.ClientName = clientName
	c.UAA.ClientSecret = clientSecret
	c.UAA.CAFile = getPath(FileUAACA)
	c.CC.BaseURL = ccBaseUrl
	c.CC.CAFile = getPath(FileCCCA)
	c.Istio.Gateways = []string{"istio-ingress"}
	return c, nil
}
