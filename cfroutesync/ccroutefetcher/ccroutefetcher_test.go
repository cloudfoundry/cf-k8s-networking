package ccroutefetcher_test

import (
	"errors"

	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/ccclient"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/ccroutefetcher"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/ccroutefetcher/fakes"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Fetching once", func() {
	var (
		fakeCCClient     *fakes.CCClient
		fakeUAAClient    *fakes.UAAClient
		fakeSnapshotRepo *fakes.SnapshotRepo
		expectedSnapshot *models.RouteSnapshot
		fetcher          *ccroutefetcher.Fetcher
	)

	BeforeEach(func() {
		fakeRoute0Destination0 := ccclient.Destination{
			Guid:   "route-0-dest-0-guid",
			Weight: models.IntPtr(10),
			Port:   8000,
		}
		fakeRoute0Destination0.App.Guid = "route-0-dest-0-app-0-guid"
		fakeRoute0Destination0.App.Process.Type = "route-0-dest-0-app-0-process-type"

		fakeRoute0Destination1 := ccclient.Destination{
			Guid:   "route-0-dest-1-guid",
			Weight: models.IntPtr(11),
			Port:   8001,
		}
		fakeRoute0Destination1.App.Guid = "route-0-dest-1-app-1-guid"
		fakeRoute0Destination1.App.Process.Type = "route-0-dest-1-app-1-process-type"

		fakeRoute1Destination0 := ccclient.Destination{
			Guid:   "route-1-dest-0-guid",
			Weight: models.IntPtr(12),
			Port:   9000,
		}
		fakeRoute1Destination0.App.Guid = "route-1-dest-0-app-0-guid"
		fakeRoute1Destination0.App.Process.Type = "route-1-dest-0-app-0-process-type"

		fakeCCClient = &fakes.CCClient{}

		routesList := []ccclient.Route{
			ccclient.Route{
				Guid: "route-0-guid",
				Host: "route-0-host",
				Path: "route-0-path",
				Url:  "route-0-url",
			},
			ccclient.Route{
				Guid: "route-1-guid",
				Host: "route-1-host",
				Path: "route-1-path",
				Url:  "route-1-url",
			},
			ccclient.Route{
				Guid: "route-2-guid",
				Host: "route-2-host",
				Path: "route-2-path",
				Url:  "route-2-url",
			},
		}
		routesList[0].Relationships.Domain.Data.Guid = "domain-0"
		routesList[1].Relationships.Domain.Data.Guid = "domain-1"
		routesList[2].Relationships.Domain.Data.Guid = "domain-1"
		fakeCCClient.ListRoutesReturns(routesList, nil)

		fakeCCClient.ListDestinationsForRouteReturnsOnCall(0, []ccclient.Destination{
			fakeRoute0Destination0,
			fakeRoute0Destination1,
		}, nil)
		fakeCCClient.ListDestinationsForRouteReturnsOnCall(1, []ccclient.Destination{
			fakeRoute1Destination0,
		}, nil)
		fakeCCClient.ListDestinationsForRouteReturnsOnCall(2, []ccclient.Destination{}, nil)

		fakeCCClient.ListDomainsReturns([]ccclient.Domain{
			{
				Guid:     "domain-0",
				Name:     "domain0.example.com",
				Internal: false,
			},
			{
				Guid:     "domain-1",
				Name:     "domain1.apps.internal",
				Internal: true,
			},
		}, nil)

		fakeUAAClient = &fakes.UAAClient{}
		fakeUAAClient.GetTokenReturns("fake-uaa-token", nil)

		fakeSnapshotRepo = &fakes.SnapshotRepo{}

		expectedSnapshot = &models.RouteSnapshot{
			Routes: []*models.Route{
				&models.Route{
					Guid: "route-0-guid",
					Host: "route-0-host",
					Path: "route-0-path",
					Domain: &models.Domain{
						Guid:     "domain-0",
						Name:     "domain0.example.com",
						Internal: false,
					},
					Destinations: []*models.Destination{
						&models.Destination{
							Guid: "route-0-dest-0-guid",
							App: models.App{
								Guid:    "route-0-dest-0-app-0-guid",
								Process: "route-0-dest-0-app-0-process-type",
							},
							Port:   8000,
							Weight: models.IntPtr(10),
						},
						&models.Destination{
							Guid: "route-0-dest-1-guid",
							App: models.App{
								Guid:    "route-0-dest-1-app-1-guid",
								Process: "route-0-dest-1-app-1-process-type",
							},
							Port:   8001,
							Weight: models.IntPtr(11),
						},
					},
				},
				&models.Route{
					Guid: "route-1-guid",
					Host: "route-1-host",
					Path: "route-1-path",
					Domain: &models.Domain{
						Guid:     "domain-1",
						Name:     "domain1.apps.internal",
						Internal: true,
					},
					Destinations: []*models.Destination{
						&models.Destination{
							Guid: "route-1-dest-0-guid",
							App: models.App{
								Guid:    "route-1-dest-0-app-0-guid",
								Process: "route-1-dest-0-app-0-process-type",
							},
							Port:   9000,
							Weight: models.IntPtr(12),
						},
					},
				},
				&models.Route{
					Guid: "route-2-guid",
					Host: "route-2-host",
					Path: "route-2-path",
					Domain: &models.Domain{
						Guid:     "domain-1",
						Name:     "domain1.apps.internal",
						Internal: true,
					},
					Destinations: nil,
				},
			},
		}

		fetcher = &ccroutefetcher.Fetcher{
			CCClient:     fakeCCClient,
			UAAClient:    fakeUAAClient,
			SnapshotRepo: fakeSnapshotRepo,
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
		Expect(routeGuid).To(Equal("route-0-guid"))

		routeGuid, _ = fakeCCClient.ListDestinationsForRouteArgsForCall(1)
		Expect(routeGuid).To(Equal("route-1-guid"))

		routeGuid, token = fakeCCClient.ListDestinationsForRouteArgsForCall(2)
		Expect(routeGuid).To(Equal("route-2-guid"))
		Expect(token).To(Equal("fake-uaa-token"))
	})

	It("converts cc types to a route snapshot and puts that into the repo", func() {
		err := fetcher.FetchOnce()
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeSnapshotRepo.PutCallCount()).To(Equal(1))
		Expect(fakeSnapshotRepo.PutArgsForCall(0)).To(Equal(expectedSnapshot))
	})

	Context("when there is an error getting the token from UAA", func() {
		It("returns the error", func() {
			fakeUAAClient.GetTokenReturns("", errors.New("banana"))
			err := fetcher.FetchOnce()
			Expect(err).To(MatchError("uaa get token: banana"))
		})
	})

	Context("when there is an error getting Routes from Cloud Controller", func() {
		It("returns the error", func() {
			fakeCCClient.ListRoutesReturns(nil, errors.New("potato!"))
			err := fetcher.FetchOnce()
			Expect(err).To(MatchError("cc list routes: potato!"))
		})
	})

	Context("when there is an error getting Destinations from Cloud Controller", func() {
		It("returns the error", func() {
			fakeCCClient.ListDestinationsForRouteReturnsOnCall(0, nil, errors.New("bam!"))
			err := fetcher.FetchOnce()
			Expect(err).To(MatchError("cc list destinations for route-0-guid: bam!"))
		})
	})

	Context("when there is an error getting Domains from Cloud Controller", func() {
		It("returns the error", func() {
			fakeCCClient.ListDomainsReturns(nil, errors.New("ohno!"))
			err := fetcher.FetchOnce()
			Expect(err).To(MatchError("cc list domains: ohno!"))
		})
	})

	Context("when a route refers to a domain that was not found", func() {
		It("returns an error", func() {
			fakeCCClient.ListDomainsReturns([]ccclient.Domain{
				{
					Guid:     "domain-1",
					Name:     "domain1.apps.internal",
					Internal: true,
				},
			}, nil)
			err := fetcher.FetchOnce()
			Expect(err).To(MatchError("route route-0-guid refers to missing domain domain-0"))
		})
	})
})
