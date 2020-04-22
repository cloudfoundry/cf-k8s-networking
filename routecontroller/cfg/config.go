package cfg

import (
	"errors"
	"os"
)

type Config struct {
	Istio struct {
		// The Istio Gateway the route controller applies to
		Gateway string
	}
}

func Load() (*Config, error) {
	c := &Config{}
	var exists bool
	c.Istio.Gateway, exists = os.LookupEnv("ISTIO_GATEWAY_NAME")

	if !exists {
		return nil, errors.New("ISTIO_GATEWAY_NAME not configured")
	}

	return c, nil
}
