package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"code.cloudfoundry.org/cf-networking-helpers/marshal"
	"code.cloudfoundry.org/tlsconfig"

	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/ccclient"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/ccroutefetcher"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/jsonclient"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/models"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/uaaclient"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/webhook"
)

func main() {
	if err := mainWithError(); err != nil {
		log.Fatalf("%s", err)
	}
}

func mainWithError() error {
	log.SetFormatter(&log.JSONFormatter{})

	var configDir string
	flag.StringVar(&configDir, "c", "", "directory with uaa config")
	flag.Parse()

	if configDir == "" {
		return fmt.Errorf("missing required flag for uaa config dir")
	}

	uaaClient, ccClient, err := buildClients(&Config{configDir})
	if err != nil {
		return err
	}

	snapshotRepo := &models.SnapshotRepo{}

	fetcher := &ccroutefetcher.Fetcher{
		CCClient:     ccClient,
		UAAClient:    uaaClient,
		SnapshotRepo: snapshotRepo,
	}

	webhookMux := http.NewServeMux()
	webhookMux.Handle("/sync", &webhook.SyncHandler{
		Marshaler:   marshal.MarshalFunc(json.Marshal),
		Unmarshaler: marshal.UnmarshalFunc(json.Unmarshal),
		Syncer: &webhook.Lineage{
			RouteSnapshotRepo: snapshotRepo,
		},
	})
	go http.ListenAndServe(":8080", webhookMux)

	for {
		err := fetcher.FetchOnce()
		if err != nil {
			log.Printf("fetch error: %s", err)
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
