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

	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/models"
	"code.cloudfoundry.org/cf-networking-helpers/marshal"

	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/synchandler"

	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/ccclient"

	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/jsonclient"

	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/uaaclient"
	"code.cloudfoundry.org/tlsconfig"
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

	repo := &models.SnapshotRepo{}
	syncer := &synchandler.RouteSyncer{
		RouteSnapshotRepo: repo,
	}
	handler := &synchandler.SyncHandler{
		Marshaler:   marshal.MarshalFunc(json.Marshal),
		Unmarshaler: marshal.UnmarshalFunc(json.Unmarshal),
		Syncer:      syncer,
	}

	http.HandleFunc("/sync", handler.ServeHTTP)
	http.ListenAndServe(":8080", nil)
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
