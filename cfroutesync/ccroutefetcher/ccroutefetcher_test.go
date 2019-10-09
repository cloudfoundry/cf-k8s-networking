package ccroutefetcher_test

import (
	"errors"
	"reflect"

	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/ccclient"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/ccroutefetcher"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/ccroutefetcher/fakes"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Fetching once", func() {
	var (
		fakeCCClient            *fakes.CCClient
		fakeUAAClient           *fakes.UAAClient
		fakeSnapshotRepo        *fakes.SnapshotRepo
		fetcher                 *ccroutefetcher.Fetcher
		fakeConstructedSnapshot *models.RouteSnapshot
	)

	BeforeEach(func() {
		fakeCCClient = &fakes.CCClient{}
		fakeRoute0 := ccclient.Route{Guid: "route-guid-0"}
		fakeRoute1 := ccclient.Route{Guid: "route-guid-1"}
		fakeRoute2 := ccclient.Route{Guid: "route-guid-2"}
		fakeCCClient.ListRoutesReturns([]ccclient.Route{
			fakeRoute0,
			fakeRoute1,
			fakeRoute2,
		}, nil)

		route0DestinationList := []ccclient.Destination{
			{
				Guid: "route-0-destination-0",
			},
			{
				Guid: "route-0-destination-1",
			},
		}

		route1DestinationList := []ccclient.Destination{
			{
				Guid: "route-1-destination-0",
			},
		}

		route2DestinationList := []ccclient.Destination{
			{
				Guid: "route-2-destination-0",
			},
		}

		fakeCCClient.ListDestinationsForRouteReturnsOnCall(0, route0DestinationList, nil)
		fakeCCClient.ListDestinationsForRouteReturnsOnCall(1, route1DestinationList, nil)
		fakeCCClient.ListDestinationsForRouteReturnsOnCall(2, route2DestinationList, nil)

		fakeUAAClient = &fakes.UAAClient{}
		fakeUAAClient.GetTokenReturns("fake-uaa-token", nil)

		fakeSnapshotRepo = &fakes.SnapshotRepo{}

		fakeConstructedSnapshot = &models.RouteSnapshot{Routes: []*models.Route{
			&models.Route{Guid: "some-constructed-data"},
		}}
		fakeSnapshotBuilder := func(ccRoutes []ccclient.Route, ccRouteDestMap map[string][]ccclient.Destination) *models.RouteSnapshot {
			expectedCCRouteDestMap := map[string][]ccclient.Destination{
				fakeRoute0.Guid: route0DestinationList,
				fakeRoute1.Guid: route1DestinationList,
				fakeRoute2.Guid: route2DestinationList,
			}
			Expect(ccRoutes).To(Equal([]ccclient.Route{
				fakeRoute0,
				fakeRoute1,
				fakeRoute2,
			}))

			equalsExpectedMapData := reflect.DeepEqual(ccRouteDestMap, expectedCCRouteDestMap)

			Expect(equalsExpectedMapData).To(BeTrue())
			return fakeConstructedSnapshot
		}

		fetcher = &ccroutefetcher.Fetcher{
			CCClient:        fakeCCClient,
			UAAClient:       fakeUAAClient,
			SnapshotRepo:    fakeSnapshotRepo,
			SnapshotBuilder: fakeSnapshotBuilder,
		}
	})

	It("calls cc client to get routes and destinations", func() {
		err := fetcher.FetchOnce()
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeCCClient.ListRoutesCallCount()).To(Equal(1))
		token := fakeCCClient.ListRoutesArgsForCall(0)
		Expect(token).To(Equal("fake-uaa-token"))

		Expect(fakeCCClient.ListDestinationsForRouteCallCount()).To(Equal(3))

		routeGuid, _ := fakeCCClient.ListDestinationsForRouteArgsForCall(0)
		Expect(routeGuid).To(Equal("route-guid-0"))

		routeGuid, _ = fakeCCClient.ListDestinationsForRouteArgsForCall(1)
		Expect(routeGuid).To(Equal("route-guid-1"))

		routeGuid, token = fakeCCClient.ListDestinationsForRouteArgsForCall(2)
		Expect(routeGuid).To(Equal("route-guid-2"))
		Expect(token).To(Equal("fake-uaa-token"))
	})

	It("converts cc types to a route snapshot and puts that into the repo", func() {
		err := fetcher.FetchOnce()
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeSnapshotRepo.PutCallCount()).To(Equal(1))
		Expect(fakeSnapshotRepo.PutArgsForCall(0)).To(BeIdenticalTo(fakeConstructedSnapshot))
	})

	Context("when there is an error getting the token from UAA", func() {
		It("returns the error", func() {
			fakeUAAClient.GetTokenReturns("", errors.New("UAA broken real bad"))
			err := fetcher.FetchOnce()
			Expect(err).To(MatchError("UAA broken real bad"))
		})
	})

	Context("when there is an error getting Routes from Cloud Controller", func() {
		It("returns the error", func() {
			fakeCCClient.ListRoutesReturns(nil, errors.New("/v3/routes broken!"))
			err := fetcher.FetchOnce()
			Expect(err).To(MatchError("/v3/routes broken!"))
		})
	})

	Context("when there is an error getting Destinations from Cloud Controller", func() {
		It("returns the error", func() {
			fakeCCClient.ListDestinationsForRouteReturnsOnCall(0, nil, errors.New("/v3/routes/:guid/destinations broken!"))
			err := fetcher.FetchOnce()
			Expect(err).To(MatchError("/v3/routes/:guid/destinations broken!"))
		})
	})
})

var _ = Describe("SnapshotBuilder", func() {
	It("creates a route snapshot", func() {
		fakeRoute0 := ccclient.Route{
			Guid: "route-0-guid",
			Host: "route-0-host",
			Path: "route-0-path",
			Url:  "route-0-url",
		}

		route0Destination0Weight := 10
		fakeRoute0Destination0 := ccclient.Destination{
			Guid:   "route-0-dest-0-guid",
			Weight: &route0Destination0Weight,
			Port:   8000,
		}
		fakeRoute0Destination0.App.Guid = "route-0-dest-0-app-0-guid"
		fakeRoute0Destination0.App.Process.Type = "route-0-dest-0-app-0-process-type"

		route0Destination1Weight := 11
		fakeRoute0Destination1 := ccclient.Destination{
			Guid:   "route-0-dest-1-guid",
			Weight: &route0Destination1Weight,
			Port:   8001,
		}
		fakeRoute0Destination1.App.Guid = "route-0-dest-1-app-1-guid"
		fakeRoute0Destination1.App.Process.Type = "route-0-dest-1-app-1-process-type"

		fakeRoute1 := ccclient.Route{
			Guid: "route-1-guid",
			Host: "route-1-host",
			Path: "route-1-path",
			Url:  "route-1-url",
		}

		route1Destination0Weight := 12
		fakeRoute1Destination0 := ccclient.Destination{
			Guid:   "route-1-dest-0-guid",
			Weight: &route1Destination0Weight,
			Port:   9000,
		}
		fakeRoute1Destination0.App.Guid = "route-1-dest-0-app-0-guid"
		fakeRoute1Destination0.App.Process.Type = "route-1-dest-0-app-0-process-type"

		routes := []ccclient.Route{fakeRoute0, fakeRoute1}
		routeDestMap := make(map[string][]ccclient.Destination)
		routeDestMap[fakeRoute0.Guid] = []ccclient.Destination{
			fakeRoute0Destination0,
			fakeRoute0Destination1,
		}
		routeDestMap[fakeRoute1.Guid] = []ccclient.Destination{
			fakeRoute1Destination0,
		}

		snapshot := ccroutefetcher.SnapshotBuilder(routes, routeDestMap)

		snapshotRoute0 := findRouteInSnapshot(fakeRoute0.Guid, snapshot)
		Expect(*snapshotRoute0).To(Equal(models.Route{
			Guid: "route-0-guid",
			Host: "route-0-host",
			Path: "route-0-path",
			Destinations: []*models.Destination{
				&models.Destination{
					Guid: "route-0-dest-0-guid",
					App: models.DestinationApp{
						Guid:    "route-0-dest-0-app-0-guid",
						Process: "route-0-dest-0-app-0-process-type",
					},
					Port:   8000,
					Weight: 10,
				},
				&models.Destination{
					Guid: "route-0-dest-1-guid",
					App: models.DestinationApp{
						Guid:    "route-0-dest-1-app-1-guid",
						Process: "route-0-dest-1-app-1-process-type",
					},
					Port:   8001,
					Weight: 11,
				},
			},
		}))

		snapshotRoute1 := findRouteInSnapshot(fakeRoute1.Guid, snapshot)
		Expect(*snapshotRoute1).To(Equal(models.Route{
			Guid: "route-1-guid",
			Host: "route-1-host",
			Path: "route-1-path",
			Destinations: []*models.Destination{
				&models.Destination{
					Guid: "route-1-dest-0-guid",
					App: models.DestinationApp{
						Guid:    "route-1-dest-0-app-0-guid",
						Process: "route-1-dest-0-app-0-process-type",
					},
					Port:   9000,
					Weight: 12,
				},
			},
		}))
	})

	Context("when a destination has a nil weight", func() {
		It("defaults the destination weight to 1", func() {
			fakeRoute0 := ccclient.Route{
				Guid: "route-0-guid",
				Host: "route-0-host",
				Path: "route-0-path",
				Url:  "route-0-url",
			}

			fakeRoute0Destination0 := ccclient.Destination{
				Guid:   "route-0-dest-0-guid",
				Weight: nil,
				Port:   8000,
			}
			fakeRoute0Destination0.App.Guid = "route-0-dest-0-app-0-guid"
			fakeRoute0Destination0.App.Process.Type = "route-0-dest-0-app-0-process-type"

			routes := []ccclient.Route{fakeRoute0}
			routeDestMap := make(map[string][]ccclient.Destination)
			routeDestMap[fakeRoute0.Guid] = []ccclient.Destination{
				fakeRoute0Destination0,
			}

			snapshot := ccroutefetcher.SnapshotBuilder(routes, routeDestMap)

			snapshotRoute0 := findRouteInSnapshot(fakeRoute0.Guid, snapshot)
			Expect(*snapshotRoute0).To(Equal(models.Route{
				Guid: "route-0-guid",
				Host: "route-0-host",
				Path: "route-0-path",
				Destinations: []*models.Destination{
					&models.Destination{
						Guid: "route-0-dest-0-guid",
						App: models.DestinationApp{
							Guid:    "route-0-dest-0-app-0-guid",
							Process: "route-0-dest-0-app-0-process-type",
						},
						Port:   8000,
						Weight: 1,
					},
				},
			}))
		})
	})
})

func findRouteInSnapshot(routeGuid string, snapshot *models.RouteSnapshot) *models.Route {
	for _, route := range snapshot.Routes {
		if route.Guid == routeGuid {
			return route
		}
	}

	return nil
}
