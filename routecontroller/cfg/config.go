package cfg

import (
	"errors"
	"fmt"
	"os"
	"time"
)

type Config struct {
	ResyncInterval time.Duration
	Istio          struct {
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

	var err error
	resync_interval, exists := os.LookupEnv("RESYNC_INTERVAL")

	if exists {
		c.ResyncInterval, err = time.ParseDuration(fmt.Sprintf("%ss", resync_interval))
		if err != nil {
			return nil, errors.New("could not parse the RESYNC_INTERVAL duration")
		}
	} else {
		c.ResyncInterval = 30 * time.Second
	}

	return c, nil
}
