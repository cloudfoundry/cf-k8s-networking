package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/ccclient"

	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/jsonclient"

	"code.cloudfoundry.org/tlsconfig"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/uaaclient"
)

func main() {
	if err := mainWithError(); err != nil {
		log.Fatalf("%s", err)
	}
}

func mainWithError() error {
	var configDir string
	flag.StringVar(&configDir, "c", "", "directory with uaa config")
	flag.Parse()

	if configDir == "" {
		return fmt.Errorf("missing required flag for uaa config dir")
	}

	uaaClient, ccClient, err := buildClients(&Config{configDir})
	if err != nil {
		return fmt.Errorf("loading UAA config: %w", err)
	}

	for {
		token, err := uaaClient.GetToken()
		if err != nil {
			return fmt.Errorf("fetching token from UAA: %w", err)
		}

		routes, err := ccClient.ListRoutes(token)
		if err != nil {
			return fmt.Errorf("listing routes with Cloud Controller: %w", err)
		}
		routeJson, err := json.MarshalIndent(routes, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling route json: %w", err)
		}
		fmt.Printf("%s\n", string(routeJson))

		for _, r := range routes {
			destinations, err := ccClient.ListDestinationsForRoute(r.Guid, token)
			if err != nil {
				return fmt.Errorf("listing destinations with Cloud Controller: %w", err)
			}
			destinationJson, err := json.MarshalIndent(destinations, "", "  ")
			if err != nil {
				return fmt.Errorf("marshaling destination json: %w", err)
			}
			fmt.Printf("%s\n", string(destinationJson))
		}

		time.Sleep(10 * time.Second)
	}
	return nil
}

type Config struct {
	configDir string
}

func (c *Config) configFilePath(configFilename string) string {
	return filepath.Join(c.configDir, configFilename)
}

func (c *Config) getConfigString(configFilename string) (string, error) {
	bytes, err := ioutil.ReadFile(c.configFilePath(configFilename))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(bytes)), nil
}
func buildClients(c *Config) (*uaaclient.Client, *ccclient.Client, error) {
	uaaBaseUrl, err := c.getConfigString("uaaBaseUrl")
	if err != nil {
		return nil, nil, err
	}
	clientName, err := c.getConfigString("clientName")
	if err != nil {
		return nil, nil, err
	}
	clientSecret, err := c.getConfigString("clientSecret")
	if err != nil {
		return nil, nil, err
	}

	ccBaseUrl, err := c.getConfigString("ccBaseUrl")
	if err != nil {
		return nil, nil, err
	}

	tlsConfig, err := tlsconfig.
		Build(tlsconfig.WithInternalServiceDefaults()).
		Client(tlsconfig.WithAuthorityFromFile(c.configFilePath("ca")))
	if err != nil {
		return nil, nil, err
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	jsonClient := &jsonclient.JSONClient{
		HTTPClient: httpClient,
	}

	return &uaaclient.Client{
			BaseURL:    uaaBaseUrl,
			Name:       clientName,
			Secret:     clientSecret,
			JSONClient: jsonClient,
		}, &ccclient.Client{
			BaseURL:    ccBaseUrl,
			JSONClient: jsonClient,
		}, nil
}
