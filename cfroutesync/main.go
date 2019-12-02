package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"code.cloudfoundry.org/cf-networking-helpers/marshal"
	"code.cloudfoundry.org/tlsconfig"
	log "github.com/sirupsen/logrus"

	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/ccclient"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/ccroutefetcher"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/cfg"
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
	log.SetOutput(os.Stdout)

	var (
		configDir  string
		listenAddr string
		verbosity  int
	)

	flag.StringVar(&configDir, "c", "", "config directory")
	flag.StringVar(&listenAddr, "l", ":8080", "listen address for serving webhook to metacontroller")
	flag.IntVar(&verbosity, "v", 4, "log verbosity")
	flag.Parse()

	log.SetLevel(log.Level(verbosity))
	if configDir == "" {
		return fmt.Errorf("missing required flag for config dir")
	}

	config, err := cfg.Load(configDir)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	log.WithFields(log.Fields{"dir": configDir}).Info("loaded config")

	uaaTLSConfig, err := tlsconfig.
		Build(tlsconfig.WithInternalServiceDefaults()).
		Client(tlsconfig.WithAuthority(config.UAA.CA))
	if err != nil {
		return fmt.Errorf("building UAA TLS config: %w", err)
	}

	ccTLSConfig, err := tlsconfig.
		Build(tlsconfig.WithInternalServiceDefaults()).
		Client(tlsconfig.WithAuthority(config.CC.CA))
	if err != nil {
		return fmt.Errorf("building CC TLS config: %w", err)
	}

	snapshotRepo := &models.SnapshotRepo{}

	fetcher := &ccroutefetcher.Fetcher{
		CCClient: &ccclient.Client{
			BaseURL: config.CC.BaseURL,
			JSONClient: &jsonclient.JSONClient{
				HTTPClient: &http.Client{
					Transport: &http.Transport{
						TLSClientConfig: ccTLSConfig,
					},
				},
			},
		},
		UAAClient: &uaaclient.Client{
			BaseURL: config.UAA.BaseURL,
			Name:    config.UAA.ClientName,
			Secret:  config.UAA.ClientSecret,
			JSONClient: &jsonclient.JSONClient{
				HTTPClient: &http.Client{
					Transport: &http.Transport{
						TLSClientConfig: uaaTLSConfig,
					},
				},
			},
		},
		SnapshotRepo: snapshotRepo,
	}

	webhookMux := http.NewServeMux()
	webhookMux.Handle("/sync", &webhook.SyncHandler{
		Marshaler:   marshal.MarshalFunc(json.Marshal),
		Unmarshaler: marshal.UnmarshalFunc(json.Unmarshal),
		Syncer: &webhook.Lineage{
			RouteSnapshotRepo: snapshotRepo,
			K8sResourceBuilders: []webhook.K8sResourceBuilder{
				&webhook.ServiceBuilder{
					PodLabelPrefix: config.Experimental.EiriniPodLabelPrefix,
				},
				&webhook.VirtualServiceBuilder{IstioGateways: config.Istio.Gateways},
			},
		},
	})

	log.Info("starting webhook server")
	go http.ListenAndServe(listenAddr, webhookMux)

	log.Info("starting cc fetch loop")
	for {
		err := fetcher.FetchOnce()
		if err != nil {
			log.WithError(err).Errorf("fetching")
		}

		time.Sleep(10 * time.Second)
	}
}
