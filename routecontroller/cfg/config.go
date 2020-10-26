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

	ingress, exists = os.LookupEnv("INGRESS_SOLUTION")
	c.IngressProvider = IngressProvider(ingress)

	if !exists {
		return nil, errors.New("INGRESS_SOLUTION not configured")
	}

	switch c.IngressProvider {
	case Istio:
		err := loadIstioConfig(c)
		if err != nil {
			return nil, err
		}
	case Contour:
		loadContour(c)
	default:
		return nil, errors.New(fmt.Sprintf("INGRESS_SOLUTION=%s not supported", c.IngressProvider))
	}

	var err error
	resyncInterval, exists := os.LookupEnv("RESYNC_INTERVAL")

	if exists {
		c.ResyncInterval, err = time.ParseDuration(fmt.Sprintf("%ss", resyncInterval))
		if err != nil {
			return nil, errors.New("could not parse the RESYNC_INTERVAL duration")
		}
	} else {
		c.ResyncInterval = 30 * time.Second
	}

	return c, nil
}

func loadContour(c *Config) {}

func loadIstioConfig(c *Config) error {
	var exists bool

	c.Istio.Gateway, exists = os.LookupEnv("ISTIO_GATEWAY_NAME")
	if !exists {
		return errors.New("ISTIO_GATEWAY_NAME not configured")
	}

	return nil
}
