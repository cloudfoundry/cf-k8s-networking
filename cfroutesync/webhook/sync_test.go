package webhook_test

import (
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/models"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/webhook"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/webhook/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Sync", func() {
	var (
		fakeSnapshotRepo        *fakes.SnapshotRepo
		syncHandler             *webhook.Lineage
		syncRequest             webhook.SyncRequest
		expectedServices        []webhook.K8sResource
		expectedVirtualServices []webhook.K8sResource
	)

	BeforeEach(func() {
		fakeSnapshotRepo = &fakes.SnapshotRepo{}
		syncHandler = &webhook.Lineage{
			RouteSnapshotRepo: fakeSnapshotRepo,
			IstioGateways:     []string{"some-gateway0", "some-gateway1"},
		}

		syncRequest = webhook.SyncRequest{
			Parent: webhook.BulkSync{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec: webhook.BulkSyncSpec{
					Template: webhook.Template{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"cloudfoundry.org/bulk-sync-route": "true",
								"label-for-routes":                 "cool-label",
							},
						},
					},
				},
			},
		}

		fullSnapshot := models.RouteSnapshot{
			Routes: []models.Route{
				models.Route{
					Guid: "route-guid-0",
					Host: "test0",
					Path: "/path0",
					Domain: models.Domain{
						Guid:     "domain-0-guid",
						Name:     "domain0.example.com",
						Internal: false,
					},
					Destinations: []models.Destination{
						models.Destination{
							Guid: "route-0-destination-guid-0",
							App: models.App{
								Guid:    "app-guid-0",
								Process: models.Process{Type: "process-type-1"},
							},
							Port:   9000,
							Weight: models.IntPtr(10),
						},
						models.Destination{
							Guid: "route-0-destination-guid-1",
							App: models.App{
								Guid:    "app-guid-1",
								Process: models.Process{Type: "process-type-1"},
							},
							Port:   9001,
							Weight: models.IntPtr(11),
						},
					},
				},
				models.Route{
					Guid: "route-guid-1",
					Host: "test1",
					Domain: models.Domain{
						Guid:     "domain-1-guid",
						Name:     "domain1.apps.internal",
						Internal: true,
					},
					Destinations: []models.Destination{
						models.Destination{
							Guid: "route-1-destination-guid-0",
							App: models.App{
								Guid:    "app-guid-2",
								Process: models.Process{Type: "process-type-2"},
							},
							Port:   8080,
							Weight: models.IntPtr(80),
						},
					},
				},
				models.Route{
					Guid: "route-guid-2",
					Host: "test0",
					Path: "/path0/deeper", // test that longst path matches first
					Domain: models.Domain{
						Guid:     "domain-0-guid",
						Name:     "domain0.example.com",
						Internal: false,
					},
					Destinations: []models.Destination{
						models.Destination{
							Guid: "route-2-destination-guid-0",
							App: models.App{
								Guid:    "app-guid-1",
								Process: models.Process{Type: "process-type-1"},
							},
							Port:   8080,
							Weight: nil, // test that weights are omitted
						},
					},
				},
			},
		}

		expectedServices = []webhook.K8sResource{
			webhook.Service{
				ApiVersion: "v1",
				Kind:       "Service",
				ObjectMeta: metav1.ObjectMeta{
					Name: "s-route-0-destination-guid-0",
					Labels: map[string]string{
						"cloudfoundry.org/bulk-sync-route": "true",
						"label-for-routes":                 "cool-label",
						"cloudfoundry.org/route":           "route-guid-0",
						"cloudfoundry.org/app":             "app-guid-0",
						"cloudfoundry.org/process":         "process-type-1",
						"cloudfoundry.org/route-fqdn":      "test0.domain0.example.com",
					},
				},
				Spec: webhook.ServiceSpec{
					Selector: map[string]string{
						"app_guid":     "app-guid-0",
						"process_type": "process-type-1",
					},

					Ports: []webhook.ServicePort{
						webhook.ServicePort{
							Port: 9000,
						},
					},
				},
			},
			webhook.Service{
				ApiVersion: "v1",
				Kind:       "Service",
				ObjectMeta: metav1.ObjectMeta{
					Name: "s-route-0-destination-guid-1",
					Labels: map[string]string{
						"cloudfoundry.org/bulk-sync-route": "true",
						"label-for-routes":                 "cool-label",
						"cloudfoundry.org/route":           "route-guid-0",
						"cloudfoundry.org/app":             "app-guid-1",
						"cloudfoundry.org/process":         "process-type-1",
						"cloudfoundry.org/route-fqdn":      "test0.domain0.example.com",
					},
				},
				Spec: webhook.ServiceSpec{
					Selector: map[string]string{
						"app_guid":     "app-guid-1",
						"process_type": "process-type-1",
					},

					Ports: []webhook.ServicePort{
						webhook.ServicePort{
							Port: 9001,
						},
					},
				},
			},
			webhook.Service{
				ApiVersion: "v1",
				Kind:       "Service",
				ObjectMeta: metav1.ObjectMeta{
					Name: "s-route-1-destination-guid-0",
					Labels: map[string]string{
						"cloudfoundry.org/bulk-sync-route": "true",
						"label-for-routes":                 "cool-label",
						"cloudfoundry.org/route":           "route-guid-1",
						"cloudfoundry.org/app":             "app-guid-2",
						"cloudfoundry.org/process":         "process-type-2",
						"cloudfoundry.org/route-fqdn":      "test1.domain1.apps.internal",
					},
				},
				Spec: webhook.ServiceSpec{
					Selector: map[string]string{
						"app_guid":     "app-guid-2",
						"process_type": "process-type-2",
					},

					Ports: []webhook.ServicePort{
						webhook.ServicePort{
							Port: 8080,
						},
					},
				},
			},
			webhook.Service{
				ApiVersion: "v1",
				Kind:       "Service",
				ObjectMeta: metav1.ObjectMeta{
					Name: "s-route-2-destination-guid-0",
					Labels: map[string]string{
						"cloudfoundry.org/bulk-sync-route": "true",
						"label-for-routes":                 "cool-label",
						"cloudfoundry.org/route":           "route-guid-2",
						"cloudfoundry.org/app":             "app-guid-1",
						"cloudfoundry.org/process":         "process-type-1",
						"cloudfoundry.org/route-fqdn":      "test0.domain0.example.com",
					},
				},
				Spec: webhook.ServiceSpec{
					Selector: map[string]string{
						"app_guid":     "app-guid-1",
						"process_type": "process-type-1",
					},

					Ports: []webhook.ServicePort{
						webhook.ServicePort{
							Port: 8080,
						},
					},
				},
			},
		}
		expectedVirtualServices = []webhook.K8sResource{
			webhook.VirtualService{
				ApiVersion: "networking.istio.io/v1alpha3",
				Kind:       "VirtualService",
				ObjectMeta: metav1.ObjectMeta{
					Name: "test0.domain0.example.com",
					Labels: map[string]string{
						"cloudfoundry.org/bulk-sync-route": "true",
						"label-for-routes":                 "cool-label",
					},
				},
				Spec: webhook.VirtualServiceSpec{
					Hosts:    []string{"test0.domain0.example.com"},
					Gateways: []string{"some-gateway0", "some-gateway1"},
					Http: []webhook.HTTPRoute{
						{
							Match: []webhook.HTTPMatchRequest{{Uri: webhook.HTTPPrefixMatch{Prefix: "/path0/deeper"}}},
							Route: []webhook.HTTPRouteDestination{
								{
									Destination: webhook.VirtualServiceDestination{Host: "s-route-2-destination-guid-0"},
									Weight:      nil,
								},
							},
						},
						{
							Match: []webhook.HTTPMatchRequest{{Uri: webhook.HTTPPrefixMatch{Prefix: "/path0"}}},
							Route: []webhook.HTTPRouteDestination{
								{
									Destination: webhook.VirtualServiceDestination{Host: "s-route-0-destination-guid-0"},
									Weight:      models.IntPtr(10),
								},
								{
									Destination: webhook.VirtualServiceDestination{Host: "s-route-0-destination-guid-1"},
									Weight:      models.IntPtr(11),
								},
							},
						},
					},
				},
			},
			webhook.VirtualService{
				ApiVersion: "networking.istio.io/v1alpha3",
				Kind:       "VirtualService",
				ObjectMeta: metav1.ObjectMeta{
					Name: "test1.domain1.apps.internal",
					Labels: map[string]string{
						"cloudfoundry.org/bulk-sync-route": "true",
						"label-for-routes":                 "cool-label",
					},
				},
				Spec: webhook.VirtualServiceSpec{
					Hosts:    []string{"test1.domain1.apps.internal"},
					Gateways: []string{"mesh"},
					Http: []webhook.HTTPRoute{
						{
							Route: []webhook.HTTPRouteDestination{
								{
									Destination: webhook.VirtualServiceDestination{Host: "s-route-1-destination-guid-0"},
									Weight:      models.IntPtr(80),
								},
							},
						},
					},
				},
			},
		}

		fakeSnapshotRepo.GetReturns(&fullSnapshot, true)
	})

	It("returns services and virtual services as a metacontroller response️", func() {
		syncResponse, err := syncHandler.Sync(syncRequest)
		Expect(err).ToNot(HaveOccurred())
		Expect(syncResponse).NotTo(BeNil())

		expectedChildren := make([]webhook.K8sResource, 0)
		expectedChildren = append(expectedChildren, expectedServices...)
		expectedChildren = append(expectedChildren, expectedVirtualServices...)
		for i, _ := range expectedChildren {
			Expect(syncResponse.Children[i]).To(Equal(expectedChildren[i]))
		}
	})

	Context("when there's snapshot but it does not contain any routes", func() {
		BeforeEach(func() {
			fakeSnapshotRepo.GetReturns(&models.RouteSnapshot{}, true)
		})

		It("returns an empty list of children in the response", func() {
			syncResponse, err := syncHandler.Sync(syncRequest)
			Expect(err).ToNot(HaveOccurred())
			Expect(syncResponse).NotTo(BeNil())
			Expect(syncResponse.Children).To(Equal([]webhook.K8sResource{}))
		})
	})

	Context("when the repo says no snapshot is available", func() {
		BeforeEach(func() {
			fakeSnapshotRepo.GetReturns(nil, false)
		})

		It("returns a meaningful error", func() {
			_, err := syncHandler.Sync(syncRequest)
			Expect(err).To(Equal(webhook.UninitializedError))
		})
	})

	Context("when two routes share an FQDN but one has no destinations", func() {
		BeforeEach(func() {
			fullSnapshot := models.RouteSnapshot{
				Routes: []models.Route{
					models.Route{
						Guid: "route-guid-0",
						Host: "test0",
						Path: "/path0",
						Domain: models.Domain{
							Guid:     "domain-0-guid",
							Name:     "domain0.example.com",
							Internal: false,
						},
						Destinations: []models.Destination{
							models.Destination{
								Guid: "route-0-destination-guid-0",
								App: models.App{
									Guid:    "app-guid-0",
									Process: models.Process{Type: "process-type-1"},
								},
								Port:   9000,
								Weight: models.IntPtr(10),
							},
						},
					},
					models.Route{
						Guid: "route-guid-1",
						Host: "test1",
						Path: "/i-dont-have-destinations",
						Domain: models.Domain{
							Guid:     "domain-0-guid",
							Name:     "domain0.example.com",
							Internal: false,
						},
						Destinations: []models.Destination{},
					},
				},
			}

			expectedServices = []webhook.K8sResource{
				webhook.Service{
					ApiVersion: "v1",
					Kind:       "Service",
					ObjectMeta: metav1.ObjectMeta{
						Name: "s-route-0-destination-guid-0",
						Labels: map[string]string{
							"cloudfoundry.org/bulk-sync-route": "true",
							"label-for-routes":                 "cool-label",
							"cloudfoundry.org/route":           "route-guid-0",
							"cloudfoundry.org/app":             "app-guid-0",
							"cloudfoundry.org/process":         "process-type-1",
							"cloudfoundry.org/route-fqdn":      "test0.domain0.example.com",
						},
					},
					Spec: webhook.ServiceSpec{
						Selector: map[string]string{
							"app_guid":     "app-guid-0",
							"process_type": "process-type-1",
						},

						Ports: []webhook.ServicePort{
							webhook.ServicePort{
								Port: 9000,
							},
						},
					},
				},
			}
			expectedVirtualServices = []webhook.K8sResource{
				webhook.VirtualService{
					ApiVersion: "networking.istio.io/v1alpha3",
					Kind:       "VirtualService",
					ObjectMeta: metav1.ObjectMeta{
						Name: "test0.domain0.example.com",
						Labels: map[string]string{
							"cloudfoundry.org/bulk-sync-route": "true",
							"label-for-routes":                 "cool-label",
						},
					},
					Spec: webhook.VirtualServiceSpec{
						Hosts:    []string{"test0.domain0.example.com"},
						Gateways: []string{"some-gateway0", "some-gateway1"},
						Http: []webhook.HTTPRoute{
							{
								Match: []webhook.HTTPMatchRequest{{Uri: webhook.HTTPPrefixMatch{Prefix: "/path0"}}},
								Route: []webhook.HTTPRouteDestination{
									{
										Destination: webhook.VirtualServiceDestination{Host: "s-route-0-destination-guid-0"},
										Weight:      models.IntPtr(10),
									},
								},
							},
						},
					},
				},
			}

			fakeSnapshotRepo.GetReturns(&fullSnapshot, true)
		})

		It("creates a virtual service that only lists a match for the route that has destinations", func() {
			syncResponse, err := syncHandler.Sync(syncRequest)
			Expect(err).ToNot(HaveOccurred())
			Expect(syncResponse).NotTo(BeNil())

			expectedChildren := make([]webhook.K8sResource, 0)
			expectedChildren = append(expectedChildren, expectedServices...)
			expectedChildren = append(expectedChildren, expectedVirtualServices...)
			for i, _ := range expectedChildren {
				Expect(syncResponse.Children[i]).To(Equal(expectedChildren[i]))
			}
		})
	})

	Context("when a route has no destinations", func() {
		BeforeEach(func() {
			fakeSnapshotRepo.GetReturns(&models.RouteSnapshot{Routes: []models.Route{
				models.Route{
					Guid: "route-guid-0",
					Host: "test0",
					Path: "/path0",
					Domain: models.Domain{
						Guid:     "domain-0-guid",
						Name:     "domain0.example.com",
						Internal: false,
					},
					Destinations: []models.Destination{},
				},
			}}, true)
		})

		It("does not create a Service", func() {
			syncResponse, err := syncHandler.Sync(syncRequest)
			Expect(err).ToNot(HaveOccurred())
			Expect(syncResponse).NotTo(BeNil())
			Expect(syncResponse.Children).NotTo(ContainElement(BeAssignableToTypeOf(webhook.Service{})))
		})

		It("does not create a VirtualService", func() {
			syncResponse, err := syncHandler.Sync(syncRequest)
			Expect(err).ToNot(HaveOccurred())
			Expect(syncResponse).NotTo(BeNil())
			Expect(syncResponse.Children).NotTo(ContainElement(BeAssignableToTypeOf(webhook.VirtualService{})))
		})
	})
})
