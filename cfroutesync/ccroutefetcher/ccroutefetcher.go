package ccroutefetcher

import (
	"fmt"

	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/ccclient"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/models"
)

//go:generate counterfeiter -o fakes/ccclient.go --fake-name CCClient . ccClient
type ccClient interface {
	ListRoutes(token string) ([]ccclient.Route, error)
	ListDestinationsForRoute(routeGUID, token string) ([]ccclient.Destination, error)
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

func (f *Fetcher) FetchOnce() error {
	token, err := f.UAAClient.GetToken()
	if err != nil {
		return fmt.Errorf("uaa get token: %w", err)
	}

	routes, err := f.CCClient.ListRoutes(token)
	if err != nil {
		return fmt.Errorf("cc list routes: %w", err)
	}

	var snapshotRoutes []*models.Route
	for _, route := range routes {
		destList, err := f.CCClient.ListDestinationsForRoute(route.Guid, token)
		if err != nil {
			return fmt.Errorf("cc list destinations for %s: %w", route.Guid, err)
		}

		snapshotRoutes = append(snapshotRoutes, buildSnapshot(route, destList))
	}

	f.SnapshotRepo.Put(&models.RouteSnapshot{Routes: snapshotRoutes})

	return nil
}

func buildSnapshot(route ccclient.Route, destinations []ccclient.Destination) *models.Route {
	var snapshotRouteDestinations []*models.Destination
	for _, ccDestination := range destinations {
		snapshotDestination := &models.Destination{
			Guid: ccDestination.Guid,
			App: models.DestinationApp{
				Guid:    ccDestination.App.Guid,
				Process: ccDestination.App.Process.Type,
			},
			Port:   ccDestination.Port,
			Weight: ccDestination.Weight,
		}
		snapshotRouteDestinations = append(snapshotRouteDestinations, snapshotDestination)
	}

	return &models.Route{
		Guid:         route.Guid,
		Host:         route.Host,
		Path:         route.Path,
		Destinations: snapshotRouteDestinations,
	}
}
