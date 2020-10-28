package cfg

import (
	"errors"
	"fmt"
	"os"
	"time"
)

type IngressProvider string

const (
	Istio   = IngressProvider("istio")
	Contour = IngressProvider("contour")
)

type Config struct {
	ResyncInterval  time.Duration
	IngressProvider IngressProvider
	Istio           struct {
		// The Istio Gateway the route controller applies to
		Gateway string
	}
}

func Load() (*Config, error) {
	c := &Config{}
	var exists bool
	var ingress string

	ingress, exists = os.LookupEnv("INGRESS_PROVIDER")
	c.IngressProvider = IngressProvider(ingress)

	if !exists {
		return nil, errors.New("INGRESS_PROVIDER not configured")
	}

	switch c.IngressProvider {
	case Istio:
		if err := loadIstioConfig(c); err != nil {
			return nil, err
		}
	case Contour:
		if err := loadContour(c); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("INGRESS_PROVIDER=%s not supported", c.IngressProvider)
	}

	var err error
	resyncInterval, exists := os.LookupEnv("RESYNC_INTERVAL")

	if exists {
		if c.ResyncInterval, err = time.ParseDuration(fmt.Sprintf("%ss", resyncInterval)); err != nil {
			return nil, errors.New("could not parse the RESYNC_INTERVAL duration")
		}
	} else {
		c.ResyncInterval = 30 * time.Second
	}

	return c, nil
}

func loadContour(c *Config) error {
	return nil
}

func loadIstioConfig(c *Config) error {
	var exists bool

	c.Istio.Gateway, exists = os.LookupEnv("ISTIO_GATEWAY_NAME")
	if !exists {
		return errors.New("ISTIO_GATEWAY_NAME not configured")
	}

	return nil
}
