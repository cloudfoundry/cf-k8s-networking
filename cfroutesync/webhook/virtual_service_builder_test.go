package webhook_test

import (
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/models"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/webhook"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("VirtualServiceBuilder", func() {
	var (
		template webhook.Template
	)

	BeforeEach(func() {
		template = webhook.Template{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"cloudfoundry.org/bulk-sync-route": "true",
					"label-for-routes":                 "cool-label",
				},
			},
		}
	})

	It("returns a VirtualService resource for each route destination", func() {
		routes := []models.Route{
			models.Route{
				Guid: "route-guid-0",
				Host: "test0",
				Path: "/path0",
				Url:  "test0.domain0.example.com/path0",
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
				Path: "",
				Url:  "test1.domain1.example.com",
				Domain: models.Domain{
					Guid:     "domain-1-guid",
					Name:     "domain1.example.com",
					Internal: false,
				},
				Destinations: []models.Destination{
					models.Destination{
						Guid: "route-1-destination-guid-0",
						App: models.App{
							Guid:    "app-guid-1",
							Process: models.Process{Type: "process-type-1"},
						},
						Port:   8080,
						Weight: models.IntPtr(12),
					},
				},
			},
		}

		expectedVirtualServices := []webhook.K8sResource{
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
					Name: "test1.domain1.example.com",
					Labels: map[string]string{
						"cloudfoundry.org/bulk-sync-route": "true",
						"label-for-routes":                 "cool-label",
					},
				},
				Spec: webhook.VirtualServiceSpec{
					Hosts:    []string{"test1.domain1.example.com"},
					Gateways: []string{"some-gateway0", "some-gateway1"},
					Http: []webhook.HTTPRoute{
						{
							Route: []webhook.HTTPRouteDestination{
								{
									Destination: webhook.VirtualServiceDestination{Host: "s-route-1-destination-guid-0"},
									Weight:      models.IntPtr(12),
								},
							},
						},
					},
				},
			},
		}

		builder := webhook.VirtualServiceBuilder{
			IstioGateways: []string{"some-gateway0", "some-gateway1"},
		}
		Expect(builder.Build(routes, template)).To(Equal(expectedVirtualServices))
	})

	Describe("inferring weights", func() {
		var routes []models.Route
		BeforeEach(func() {
			routes = []models.Route{
				models.Route{
					Guid: "route-guid-0",
					Host: "test0",
					Path: "/path0",
					Url:  "test0.domain0.example.com/path0",
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
							Weight: nil,
						},
						models.Destination{
							Guid: "route-0-destination-guid-1",
							App: models.App{
								Guid:    "app-guid-1",
								Process: models.Process{Type: "process-type-1"},
							},
							Port:   8080,
							Weight: nil,
						},
						models.Destination{
							Guid: "route-0-destination-guid-2",
							App: models.App{
								Guid:    "app-guid-2",
								Process: models.Process{Type: "process-type-1"},
							},
							Port:   8080,
							Weight: nil,
						},
					},
				},
			}
		})

		Context("when weights aren't present but a route has multiple destinations", func() {
			It("evenly distributes weights that sum to 100", func() {
				expectedVirtualServices := []webhook.K8sResource{
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
											Weight:      models.IntPtr(34),
										},
										{
											Destination: webhook.VirtualServiceDestination{Host: "s-route-0-destination-guid-1"},
											Weight:      models.IntPtr(33),
										},
										{
											Destination: webhook.VirtualServiceDestination{Host: "s-route-0-destination-guid-2"},
											Weight:      models.IntPtr(33),
										},
									},
								},
							},
						},
					},
				}

				builder := webhook.VirtualServiceBuilder{
					IstioGateways: []string{"some-gateway0", "some-gateway1"},
				}
				Expect(builder.Build(routes, template)).To(Equal(expectedVirtualServices))
			})

			Context("when one destination has a weight but the rest do not", func() {
				It("panics, because this is invalid data from CAPI", func() {
					routes[0].Destinations[0].Weight = models.IntPtr(10)
					builder := webhook.VirtualServiceBuilder{
						IstioGateways: []string{"some-gateway0", "some-gateway1"},
					}
					Expect(func() { builder.Build(routes, template) }).To(Panic())
				})
			})
		})
	})

	Context("when a route is for an internal domain", func() {
		It("uses the internal mesh gateways", func() {
			routes := []models.Route{
				models.Route{
					Guid: "route-guid-0",
					Host: "test0",
					Path: "",
					Url:  "test0.domain0.apps.internal",
					Domain: models.Domain{
						Guid:     "domain-0-guid",
						Name:     "domain0.apps.internal",
						Internal: true,
					},
					Destinations: []models.Destination{
						models.Destination{
							Guid: "route-0-destination-guid-0",
							App: models.App{
								Guid:    "app-guid-0",
								Process: models.Process{Type: "process-type-0"},
							},
							Port:   8080,
							Weight: models.IntPtr(80),
						},
					},
				},
			}

			expectedVirtualServices := []webhook.K8sResource{
				webhook.VirtualService{
					ApiVersion: "networking.istio.io/v1alpha3",
					Kind:       "VirtualService",
					ObjectMeta: metav1.ObjectMeta{
						Name: "test0.domain0.apps.internal",
						Labels: map[string]string{
							"cloudfoundry.org/bulk-sync-route": "true",
							"label-for-routes":                 "cool-label",
						},
					},
					Spec: webhook.VirtualServiceSpec{
						Hosts:    []string{"test0.domain0.apps.internal"},
						Gateways: []string{"mesh"},
						Http: []webhook.HTTPRoute{
							{
								Route: []webhook.HTTPRouteDestination{
									{
										Destination: webhook.VirtualServiceDestination{Host: "s-route-0-destination-guid-0"},
										Weight:      models.IntPtr(80),
									},
								},
							},
						},
					},
				},
			}

			builder := webhook.VirtualServiceBuilder{
				IstioGateways: []string{"some-gateway0", "some-gateway1"},
			}
			Expect(builder.Build(routes, template)).To(Equal(expectedVirtualServices))
		})

	})

	Context("when two routes have the same fqdn", func() {
		It("orders the paths by longest matching prefix", func() {
			routes := []models.Route{
				models.Route{
					Guid: "route-guid-0",
					Host: "test0",
					Path: "/path0",
					Url:  "test0.domain0.example.com/path0",
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
					Host: "test0",
					Path: "/path0/deeper",
					Url:  "test0.domain0.example.com/path0/deeper",
					Domain: models.Domain{
						Guid:     "domain-0-guid",
						Name:     "domain0.example.com",
						Internal: false,
					},
					Destinations: []models.Destination{
						models.Destination{
							Guid: "route-1-destination-guid-0",
							App: models.App{
								Guid:    "app-guid-1",
								Process: models.Process{Type: "process-type-1"},
							},
							Port:   8080,
							Weight: nil,
						},
					},
				},
			}

			expectedVirtualServices := []webhook.K8sResource{
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
										Destination: webhook.VirtualServiceDestination{Host: "s-route-1-destination-guid-0"},
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
								},
							},
						},
					},
				},
			}

			builder := webhook.VirtualServiceBuilder{
				IstioGateways: []string{"some-gateway0", "some-gateway1"},
			}
			Expect(builder.Build(routes, template)).To(Equal(expectedVirtualServices))
		})

		Context("and one of the routes has no destinations", func() {
			It("ignores the route without destinations", func() {
				routes := []models.Route{
					models.Route{
						Guid: "route-guid-0",
						Host: "test0",
						Path: "/path0",
						Url:  "test0.domain0.example.com/path0",
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
						Host: "test0",
						Path: "/path0/deeper",
						Url:  "test0.domain0.example.com/path0/deeper",
						Domain: models.Domain{
							Guid:     "domain-0-guid",
							Name:     "domain0.example.com",
							Internal: false,
						},
						Destinations: []models.Destination{},
					},
				}

				expectedVirtualServices := []webhook.K8sResource{
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

				builder := webhook.VirtualServiceBuilder{
					IstioGateways: []string{"some-gateway0", "some-gateway1"},
				}
				Expect(builder.Build(routes, template)).To(Equal(expectedVirtualServices))

			})
		})

		Context("and one route is internal and one is external", func() {
			It("panics", func() {
				routes := []models.Route{
					models.Route{
						Guid: "route-guid-0",
						Host: "test0",
						Path: "",
						Url:  "test0.domain0.example.com",
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
						Host: "test0",
						Path: "",
						Url:  "test0.domain0.example.com",
						Domain: models.Domain{
							Guid:     "domain-0-guid",
							Name:     "domain0.example.com",
							Internal: true,
						},
						Destinations: []models.Destination{
							models.Destination{
								Guid: "route-1-destination-guid-1",
								App: models.App{
									Guid:    "app-guid-1",
									Process: models.Process{Type: "process-type-1"},
								},
								Port:   9000,
								Weight: models.IntPtr(10),
							},
						},
					},
				}

				builder := webhook.VirtualServiceBuilder{
					IstioGateways: []string{"some-gateway0", "some-gateway1"},
				}
				Expect(func() { builder.Build(routes, template) }).To(Panic())
			})
		})
	})

	Context("when a route has no destinations", func() {
		It("does not create a VirtualService", func() {
			routes := []models.Route{
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
			}

			builder := webhook.VirtualServiceBuilder{
				IstioGateways: []string{"some-gateway0", "some-gateway1"},
			}
			Expect(builder.Build(routes, template)).To(Equal([]webhook.K8sResource{}))
		})
	})

	Context("when a destination has no weight", func() {
		It("omits weight on the VirtualSevice", func() {
			routes := []models.Route{
				models.Route{
					Guid: "route-guid-0",
					Host: "test0",
					Path: "",
					Url:  "test0.domain0.example.com",
					Domain: models.Domain{
						Guid:     "domain-0-guid",
						Name:     "domain0.example.com",
						Internal: false,
					},
					Destinations: []models.Destination{
						models.Destination{
							Guid: "route-0-destination-guid-0",
							App: models.App{
								Guid:    "app-guid-1",
								Process: models.Process{Type: "process-type-1"},
							},
							Port:   8080,
							Weight: nil,
						},
					},
				},
			}

			expectedVirtualServices := []webhook.K8sResource{
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
								Route: []webhook.HTTPRouteDestination{
									{
										Destination: webhook.VirtualServiceDestination{Host: "s-route-0-destination-guid-0"},
										Weight:      nil,
									},
								},
							},
						},
					},
				},
			}

			builder := webhook.VirtualServiceBuilder{
				IstioGateways: []string{"some-gateway0", "some-gateway1"},
			}
			Expect(builder.Build(routes, template)).To(Equal(expectedVirtualServices))
		})
	})
})
