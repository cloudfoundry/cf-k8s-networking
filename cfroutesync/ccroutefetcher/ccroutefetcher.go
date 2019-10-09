package ccroutefetcher

import (
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
	CCClient        ccClient
	UAAClient       uaaClient
	SnapshotRepo    snapshotRepo
	SnapshotBuilder func([]ccclient.Route, map[string][]ccclient.Destination) *models.RouteSnapshot
}

func (f *Fetcher) FetchOnce() error {
	token, err := f.UAAClient.GetToken()
	if err != nil {
		return err
	}

	routes, err := f.CCClient.ListRoutes(token)
	if err != nil {
		return err
	}

	routeGuidDestinationMap := make(map[string][]ccclient.Destination)
	for _, route := range routes {
		destList, err := f.CCClient.ListDestinationsForRoute(route.Guid, token)
		if err != nil {
			return err
		}

		routeGuidDestinationMap[route.Guid] = destList
	}

	snapshot := f.SnapshotBuilder(routes, routeGuidDestinationMap)

	f.SnapshotRepo.Put(snapshot)

	return nil
}

func SnapshotBuilder(routes []ccclient.Route, routeDestinationMap map[string][]ccclient.Destination) *models.RouteSnapshot {
	routeMap := make(map[string]ccclient.Route)
	for _, route := range routes {
		routeMap[route.Guid] = route
	}

	var snapshotRoutes []*models.Route
	for routeGuid, ccDestinationList := range routeDestinationMap {
		ccRoute := routeMap[routeGuid]

		var snapshotRouteDestinations []*models.Destination
		for _, ccDestination := range ccDestinationList {
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
		snapshotRoute := &models.Route{
			Guid:         ccRoute.Guid,
			Host:         ccRoute.Host,
			Path:         ccRoute.Path,
			Destinations: snapshotRouteDestinations,
		}

		snapshotRoutes = append(snapshotRoutes, snapshotRoute)
	}

	return &models.RouteSnapshot{Routes: snapshotRoutes}
}
