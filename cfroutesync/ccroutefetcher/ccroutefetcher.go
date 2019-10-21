package ccroutefetcher

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/ccclient"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/models"
)

//go:generate counterfeiter -o fakes/ccclient.go --fake-name CCClient . ccClient
type ccClient interface {
	ListRoutes(token string) ([]ccclient.Route, error)
	ListDestinationsForRoute(routeGUID, token string) ([]ccclient.Destination, error)
	ListDomains(token string) ([]ccclient.Domain, error)
}

//go:generate counterfeiter -o fakes/uaaclient.go --fake-name UAAClient . uaaClient
type uaaClient interface {
	GetToken() (string, error)
}

//go:generate counterfeiter -o fakes/snapshotrepo.go --fake-name SnapshotRepo . snapshotRepo
type snapshotRepo interface {
	Put(snapshot *models.RouteSnapshot)
}

type Fetcher struct {
	CCClient     ccClient
	UAAClient    uaaClient
	SnapshotRepo snapshotRepo
}

// FetchOnce gets all the routing data from CC, builds a snapshot and puts it into the repo
func (f *Fetcher) FetchOnce() error {
	token, err := f.UAAClient.GetToken()
	if err != nil {
		return fmt.Errorf("uaa get token: %w", err)
	}

	routes, err := f.CCClient.ListRoutes(token)
	if err != nil {
		return fmt.Errorf("cc list routes: %w", err)
	}

	domains, err := f.CCClient.ListDomains(token)
	if err != nil {
		return fmt.Errorf("cc list domains: %w", err)
	}
	domainsMap := make(map[string]ccclient.Domain)
	for _, domain := range domains {
		domainsMap[domain.Guid] = domain
	}

	var snapshotRoutes []models.Route
	for _, route := range routes {
		destList, err := f.CCClient.ListDestinationsForRoute(route.Guid, token)
		if err != nil {
			return fmt.Errorf("cc list destinations for %s: %w", route.Guid, err)
		}

		routeDomainGuid := route.Relationships.Domain.Data.Guid
		domain, ok := domainsMap[routeDomainGuid]
		if !ok {
			return fmt.Errorf("route %s refers to missing domain %s", route.Guid, routeDomainGuid)
		}

		snapshotRoutes = append(snapshotRoutes, buildRouteForSnapshot(route, destList, domain))
	}

	snapshot := &models.RouteSnapshot{Routes: snapshotRoutes}
	f.SnapshotRepo.Put(snapshot)
	log.WithFields(log.Fields{
		"snapshot": *snapshot,
	}).Debug("Fetched and put snapshot")

	return nil
}

func buildRouteForSnapshot(route ccclient.Route, destinations []ccclient.Destination, domain ccclient.Domain) models.Route {
	var snapshotRouteDestinations []models.Destination
	for _, ccDestination := range destinations {
		snapshotDestination := models.Destination{
			Guid: ccDestination.Guid,
			App: models.App{
				Guid:    ccDestination.App.Guid,
				Process: models.Process{Type: ccDestination.App.Process.Type},
			},
			Port:   ccDestination.Port,
			Weight: ccDestination.Weight,
		}
		snapshotRouteDestinations = append(snapshotRouteDestinations, snapshotDestination)
	}

	return models.Route{
		Guid:         route.Guid,
		Host:         route.Host,
		Path:         route.Path,
		Url:          route.Url,
		Destinations: snapshotRouteDestinations,
		Domain: models.Domain{
			Guid:     domain.Guid,
			Name:     domain.Name,
			Internal: domain.Internal,
		},
	}
}
