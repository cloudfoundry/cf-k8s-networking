package resourcebuilders

import (
	"fmt"
	"strings"

	appsv1alpha1 "code.cloudfoundry.org/cf-k8s-networking/routecontroller/apis/apps/v1alpha1"
	networkingv1alpha3 "code.cloudfoundry.org/cf-k8s-networking/routecontroller/apis/networking/v1alpha3"
	istiov1alpha3 "istio.io/api/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("VirtualServiceBuilder", func() {
	It("returns a VirtualService resource for each route destination", func() {
		routes := appsv1alpha1.RouteList{Items: []appsv1alpha1.Route{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "route-guid-0",
					Namespace: "workload-namespace",
					Labels: map[string]string{
						"cloudfoundry.org/space_guid": "space-guid-0",
						"cloudfoundry.org/org_guid":   "org-guid-0",
					},
				},
				Spec: appsv1alpha1.RouteSpec{
					Host: "test0",
					Path: "/path0",
					Url:  "test0.domain0.example.com/path0",
					Domain: appsv1alpha1.RouteDomain{
						Name:     "domain0.example.com",
						Internal: false,
					},
					Destinations: []appsv1alpha1.RouteDestination{
						appsv1alpha1.RouteDestination{
							Guid:   "route-0-destination-guid-0",
							Port:   intPtr(9000),
							Weight: intPtr(91),
							App: appsv1alpha1.DestinationApp{
								Guid:    "app-guid-0",
								Process: appsv1alpha1.AppProcess{Type: "process-type-1"},
							},
							Selector: appsv1alpha1.DestinationSelector{
								MatchLabels: map[string]string{
									"cloudfoundry.org/app_guid":     "app-guid-0",
									"cloudfoundry.org/process_type": "process-type-1",
								},
							},
						},
						appsv1alpha1.RouteDestination{
							Guid:   "route-0-destination-guid-1",
							Port:   intPtr(9001),
							Weight: intPtr(9),
							App: appsv1alpha1.DestinationApp{
								Guid:    "app-guid-1",
								Process: appsv1alpha1.AppProcess{Type: "process-type-1"},
							},
							Selector: appsv1alpha1.DestinationSelector{
								MatchLabels: map[string]string{
									"cloudfoundry.org/app_guid":     "app-guid-1",
									"cloudfoundry.org/process_type": "process-type-1",
								},
							},
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "route-guid-1",
					Namespace: "workload-namespace",
					Labels: map[string]string{
						"cloudfoundry.org/space_guid": "space-guid-1",
						"cloudfoundry.org/org_guid":   "org-guid-1",
					},
				},
				Spec: appsv1alpha1.RouteSpec{
					Host: "test1",
					Path: "",
					Url:  "test1.domain1.example.com",
					Domain: appsv1alpha1.RouteDomain{
						Name:     "domain1.example.com",
						Internal: false,
					},
					Destinations: []appsv1alpha1.RouteDestination{
						appsv1alpha1.RouteDestination{
							Guid:   "route-1-destination-guid-0",
							Port:   intPtr(8080),
							Weight: intPtr(100),
							App: appsv1alpha1.DestinationApp{
								Guid:    "app-guid-1",
								Process: appsv1alpha1.AppProcess{Type: "process-type-1"},
							},
							Selector: appsv1alpha1.DestinationSelector{
								MatchLabels: map[string]string{
									"cloudfoundry.org/app_guid":     "app-guid-1",
									"cloudfoundry.org/process_type": "process-type-1",
								},
							},
						},
					},
				},
			},
		},
		}

		expectedVirtualServices := []networkingv1alpha3.VirtualService{
			networkingv1alpha3.VirtualService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      VirtualServiceName("test0.domain0.example.com"),
					Namespace: "workload-namespace",
					Labels:    map[string]string{},
					Annotations: map[string]string{
						"cloudfoundry.org/fqdn": "test0.domain0.example.com",
					},
				},
				Spec: networkingv1alpha3.VirtualServiceSpec{
					VirtualService: istiov1alpha3.VirtualService{
						Hosts:    []string{"test0.domain0.example.com"},
						Gateways: []string{"some-gateway0", "some-gateway1"},
						Http: []*istiov1alpha3.HTTPRoute{
							{
								Match: []*istiov1alpha3.HTTPMatchRequest{
									{
										Uri: &istiov1alpha3.StringMatch{
											MatchType: &istiov1alpha3.StringMatch_Prefix{
												Prefix: "/path0",
											},
										},
									},
								},
								Route: []*istiov1alpha3.HTTPRouteDestination{
									{
										Destination: &istiov1alpha3.Destination{Host: "s-route-0-destination-guid-0"},
										Headers: &istiov1alpha3.Headers{
											Request: &istiov1alpha3.Headers_HeaderOperations{
												Set: map[string]string{
													"CF-App-Id":           "app-guid-0",
													"CF-App-Process-Type": "process-type-1",
													"CF-Space-Id":         "space-guid-0",
													"CF-Organization-Id":  "org-guid-0",
												},
											},
										},
										Weight: 91,
									},
									{
										Destination: &istiov1alpha3.Destination{Host: "s-route-0-destination-guid-1"},
										Headers: &istiov1alpha3.Headers{
											Request: &istiov1alpha3.Headers_HeaderOperations{
												Set: map[string]string{
													"CF-App-Id":           "app-guid-1",
													"CF-App-Process-Type": "process-type-1",
													"CF-Space-Id":         "space-guid-0",
													"CF-Organization-Id":  "org-guid-0",
												},
											},
										},
										Weight: 9,
									},
								},
							},
						},
					},
				},
			},
			networkingv1alpha3.VirtualService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      VirtualServiceName("test1.domain1.example.com"),
					Namespace: "workload-namespace",
					Labels:    map[string]string{},
					Annotations: map[string]string{
						"cloudfoundry.org/fqdn": "test1.domain1.example.com",
					},
				},
				Spec: networkingv1alpha3.VirtualServiceSpec{
					VirtualService: istiov1alpha3.VirtualService{
						Hosts:    []string{"test1.domain1.example.com"},
						Gateways: []string{"some-gateway0", "some-gateway1"},
						Http: []*istiov1alpha3.HTTPRoute{
							{
								Route: []*istiov1alpha3.HTTPRouteDestination{
									{
										Destination: &istiov1alpha3.Destination{Host: "s-route-1-destination-guid-0"},
										Headers: &istiov1alpha3.Headers{
											Request: &istiov1alpha3.Headers_HeaderOperations{
												Set: map[string]string{
													"CF-App-Id":           "app-guid-1",
													"CF-App-Process-Type": "process-type-1",
													"CF-Space-Id":         "space-guid-1",
													"CF-Organization-Id":  "org-guid-1",
												},
											},
										},
										Weight: 100,
									},
								},
							},
						},
					},
				},
			},
		}

		builder := VirtualServiceBuilder{
			IstioGateways: []string{"some-gateway0", "some-gateway1"},
		}
		Expect(builder.Build(&routes)).To(Equal(expectedVirtualServices))
	})

	Describe("inferring weights", func() {
		var routes appsv1alpha1.RouteList

		BeforeEach(func() {
			routes = appsv1alpha1.RouteList{Items: []appsv1alpha1.Route{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "route-guid-0",
						Namespace: "workload-namespace",
						Labels: map[string]string{
							"cloudfoundry.org/space_guid": "space-guid-1",
							"cloudfoundry.org/org_guid":   "org-guid-1",
						},
					},
					Spec: appsv1alpha1.RouteSpec{
						Host: "test0",
						Path: "/path0",
						Url:  "test0.domain0.example.com/path0",
						Domain: appsv1alpha1.RouteDomain{
							Name:     "domain0.example.com",
							Internal: false,
						},
						Destinations: []appsv1alpha1.RouteDestination{
							appsv1alpha1.RouteDestination{
								Guid: "route-0-destination-guid-0",
								App: appsv1alpha1.DestinationApp{
									Guid:    "app-guid-0",
									Process: appsv1alpha1.AppProcess{Type: "process-type-1"},
								},
								Port:   intPtr(9000),
								Weight: nil,
							},
							appsv1alpha1.RouteDestination{
								Guid: "route-0-destination-guid-1",
								App: appsv1alpha1.DestinationApp{
									Guid:    "app-guid-1",
									Process: appsv1alpha1.AppProcess{Type: "process-type-1"},
								},
								Port:   intPtr(8080),
								Weight: nil,
							},
							appsv1alpha1.RouteDestination{
								Guid: "route-0-destination-guid-2",
								App: appsv1alpha1.DestinationApp{
									Guid:    "app-guid-2",
									Process: appsv1alpha1.AppProcess{Type: "process-type-1"},
								},
								Port:   intPtr(8080),
								Weight: nil,
							},
						},
					},
				},
			},
			}
		})

		Context("when weights aren't present but a route has multiple destinations", func() {
			Context("when the destinations DO NOT evenly divide to 100", func() {
				It("ensures the weights add to 100 and adds any remainder to the first destination", func() {
					builder := VirtualServiceBuilder{
						IstioGateways: []string{"some-gateway0", "some-gateway1"},
					}
					virtualservice := builder.Build(&routes)[0]
					Expect(virtualservice.Spec.Http[0].Route[0].Weight).To(Equal(int32(34)))
					Expect(virtualservice.Spec.Http[0].Route[1].Weight).To(Equal(int32(33)))
					Expect(virtualservice.Spec.Http[0].Route[2].Weight).To(Equal(int32(33)))
				})
			})

			Context("when the destinations DO evenly divide to 100", func() {
				It("evenly distributes the weights", func() {
					routes.Items[0].Spec.Destinations = []appsv1alpha1.RouteDestination{
						appsv1alpha1.RouteDestination{
							Guid: "route-0-destination-guid-0",
							App: appsv1alpha1.DestinationApp{
								Guid:    "app-guid-0",
								Process: appsv1alpha1.AppProcess{Type: "process-type-1"},
							},
							Port:   intPtr(9000),
							Weight: nil,
						},
						appsv1alpha1.RouteDestination{
							Guid: "route-0-destination-guid-1",
							App: appsv1alpha1.DestinationApp{
								Guid:    "app-guid-1",
								Process: appsv1alpha1.AppProcess{Type: "process-type-1"},
							},
							Port:   intPtr(8080),
							Weight: nil,
						},
					}

					builder := VirtualServiceBuilder{
						IstioGateways: []string{"some-gateway0", "some-gateway1"},
					}
					virtualservice := builder.Build(&routes)[0]
					Expect(virtualservice.Spec.Http[0].Route[0].Weight).To(Equal(int32(50)))
					Expect(virtualservice.Spec.Http[0].Route[1].Weight).To(Equal(int32(50)))
				})
			})
		})

		Context("when weights are present", func() {
			Context("when the weights sum up to 100", func() {
				It("leaves the weights alone", func() {
					routes.Items[0].Spec.Destinations[0].Weight = intPtr(70)
					routes.Items[0].Spec.Destinations[1].Weight = intPtr(20)
					routes.Items[0].Spec.Destinations[2].Weight = intPtr(10)

					builder := VirtualServiceBuilder{
						IstioGateways: []string{"some-gateway0", "some-gateway1"},
					}

					virtualservices := builder.Build(&routes)[0]
					Expect(virtualservices.Spec.Http[0].Route[0].Weight).To(Equal(int32(70)))
					Expect(virtualservices.Spec.Http[0].Route[1].Weight).To(Equal(int32(20)))
					Expect(virtualservices.Spec.Http[0].Route[2].Weight).To(Equal(int32(10)))
				})
			})

			Context("when the weights do not sum up to 100", func() {
				It("omits the invalid networkingv1alpha3.VirtualService", func() {
					invalidRoute := appsv1alpha1.Route{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "route-guid-0",
							Namespace: "workload-namespace",
							Labels: map[string]string{
								"cloudfoundry.org/space_guid": "space-guid-1",
								"cloudfoundry.org/org_guid":   "org-guid-1",
							},
						},
						Spec: appsv1alpha1.RouteSpec{
							Host: "invalid-route",
							Path: "/path0",
							Url:  "invalid-route.domain0.example.com/path0",
							Domain: appsv1alpha1.RouteDomain{
								Name:     "domain0.example.com",
								Internal: false,
							},
							Destinations: []appsv1alpha1.RouteDestination{
								appsv1alpha1.RouteDestination{
									Guid: "route-0-destination-guid-0",
									App: appsv1alpha1.DestinationApp{
										Guid:    "app-guid-0",
										Process: appsv1alpha1.AppProcess{Type: "process-type-1"},
									},
									Port:   intPtr(9000),
									Weight: intPtr(80),
								},
								appsv1alpha1.RouteDestination{
									Guid: "route-0-destination-guid-1",
									App: appsv1alpha1.DestinationApp{
										Guid:    "app-guid-1",
										Process: appsv1alpha1.AppProcess{Type: "process-type-1"},
									},
									Port:   intPtr(8080),
									Weight: intPtr(80),
								},
							},
						},
					}
					routes.Items = append(routes.Items, invalidRoute)

					builder := VirtualServiceBuilder{
						IstioGateways: []string{"some-gateway0", "some-gateway1"},
					}

					k8sResources := builder.Build(&routes)
					Expect(len(k8sResources)).To(Equal(1))

					virtualservices := k8sResources[0]
					Expect(len(virtualservices.Spec.Http)).To(Equal(1))
				})
			})

			Context("when one destination for a given route has a weight but the rest do not", func() {
				BeforeEach(func() {
					invalidRoute := appsv1alpha1.Route{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "route-guid-0",
							Namespace: "workload-namespace",
							Labels: map[string]string{
								"cloudfoundry.org/space_guid": "space-guid-1",
								"cloudfoundry.org/org_guid":   "org-guid-1",
							},
						},
						Spec: appsv1alpha1.RouteSpec{
							Host: "invalid-route",
							Path: "/path0",
							Url:  "invalid-route.domain0.example.com/path0",
							Domain: appsv1alpha1.RouteDomain{
								Name:     "domain0.example.com",
								Internal: false,
							},
							Destinations: []appsv1alpha1.RouteDestination{
								appsv1alpha1.RouteDestination{
									Guid: "route-0-destination-guid-0",
									App: appsv1alpha1.DestinationApp{
										Guid:    "app-guid-0",
										Process: appsv1alpha1.AppProcess{Type: "process-type-1"},
									},
									Port:   intPtr(9000),
									Weight: intPtr(80),
								},
								appsv1alpha1.RouteDestination{
									Guid: "route-0-destination-guid-1",
									App: appsv1alpha1.DestinationApp{
										Guid:    "app-guid-1",
										Process: appsv1alpha1.AppProcess{Type: "process-type-1"},
									},
									Port:   intPtr(8080),
									Weight: nil,
								},
							},
						},
					}
					routes.Items = append(routes.Items, invalidRoute)
				})

				It("omits the invalid networkingv1alpha3.VirtualService", func() {
					builder := VirtualServiceBuilder{
						IstioGateways: []string{"some-gateway0", "some-gateway1"},
					}

					k8sResources := builder.Build(&routes)
					Expect(len(k8sResources)).To(Equal(1))

					virtualservices := k8sResources[0]
					Expect(len(virtualservices.Spec.Http)).To(Equal(1))
				})
			})
		})

		Context("when a route is for an internal domain", func() {
			BeforeEach(func() {
				routes = appsv1alpha1.RouteList{Items: []appsv1alpha1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "route-guid-0",
							Namespace: "workload-namespace",
							Labels: map[string]string{
								"cloudfoundry.org/space_guid": "space-guid-0",
								"cloudfoundry.org/org_guid":   "org-guid-0",
							},
						},
						Spec: appsv1alpha1.RouteSpec{
							Host: "test0",
							Path: "",
							Url:  "test0.domain0.apps.internal",
							Domain: appsv1alpha1.RouteDomain{
								Name:     "domain0.apps.internal",
								Internal: true,
							},
							Destinations: []appsv1alpha1.RouteDestination{
								{
									Guid: "route-0-destination-guid-0",
									App: appsv1alpha1.DestinationApp{
										Guid:    "app-guid-0",
										Process: appsv1alpha1.AppProcess{Type: "process-type-0"},
									},
									Port:   intPtr(8080),
									Weight: intPtr(100),
								},
							},
						},
					},
				},
				}
			})

			It("uses the internal mesh gateways", func() {
				builder := VirtualServiceBuilder{
					IstioGateways: []string{"some-gateway0", "some-gateway1"},
				}

				virtualservices := builder.Build(&routes)[0]
				Expect(len(virtualservices.Spec.Gateways)).To(Equal(1))
				Expect(virtualservices.Spec.Gateways[0]).To(Equal("mesh"))
			})
		})

		Context("when two routes have the same fqdn", func() {
			BeforeEach(func() {
				routes = appsv1alpha1.RouteList{Items: []appsv1alpha1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "route-guid-0",
							Namespace: "workload-namespace",
							Labels: map[string]string{
								"cloudfoundry.org/space_guid": "space-guid-0",
								"cloudfoundry.org/org_guid":   "org-guid-0",
							},
						},
						Spec: appsv1alpha1.RouteSpec{
							Host: "test0",
							Path: "/path0",
							Url:  "test0.domain0.example.com/path0",
							Domain: appsv1alpha1.RouteDomain{
								Name:     "domain0.example.com",
								Internal: false,
							},
							Destinations: []appsv1alpha1.RouteDestination{
								{
									Guid: "route-0-destination-guid-0",
									App: appsv1alpha1.DestinationApp{
										Guid:    "app-guid-0",
										Process: appsv1alpha1.AppProcess{Type: "process-type-1"},
									},
									Port:   intPtr(9000),
									Weight: intPtr(100),
								},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "route-guid-1",
							Namespace: "workload-namespace",
							Labels: map[string]string{
								"cloudfoundry.org/space_guid": "space-guid-0",
								"cloudfoundry.org/org_guid":   "org-guid-0",
							},
						},
						Spec: appsv1alpha1.RouteSpec{
							Host: "test0",
							Path: "/path0/deeper",
							Url:  "test0.domain0.example.com/path0/deeper",
							Domain: appsv1alpha1.RouteDomain{
								Name:     "domain0.example.com",
								Internal: false,
							},
							Destinations: []appsv1alpha1.RouteDestination{
								{
									Guid: "route-1-destination-guid-0",
									App: appsv1alpha1.DestinationApp{
										Guid:    "app-guid-1",
										Process: appsv1alpha1.AppProcess{Type: "process-type-1"},
									},
									Port:   intPtr(8080),
									Weight: intPtr(100),
								},
							},
						},
					},
				},
				}
			})

			It("orders the paths by longest matching prefix", func() {
				expectedVirtualServices := []networkingv1alpha3.VirtualService{
					networkingv1alpha3.VirtualService{
						ObjectMeta: metav1.ObjectMeta{
							Name:      VirtualServiceName("test0.domain0.example.com"),
							Namespace: "workload-namespace",
							Labels:    map[string]string{},
							Annotations: map[string]string{
								"cloudfoundry.org/fqdn": "test0.domain0.example.com",
							},
						},
						Spec: networkingv1alpha3.VirtualServiceSpec{
							VirtualService: istiov1alpha3.VirtualService{
								Hosts:    []string{"test0.domain0.example.com"},
								Gateways: []string{"some-gateway0", "some-gateway1"},
								Http: []*istiov1alpha3.HTTPRoute{
									{
										Match: []*istiov1alpha3.HTTPMatchRequest{
											{
												Uri: &istiov1alpha3.StringMatch{
													MatchType: &istiov1alpha3.StringMatch_Prefix{
														Prefix: "/path0/deeper",
													},
												},
											},
										},
										Route: []*istiov1alpha3.HTTPRouteDestination{
											{
												Destination: &istiov1alpha3.Destination{Host: "s-route-1-destination-guid-0"},
												Headers: &istiov1alpha3.Headers{
													Request: &istiov1alpha3.Headers_HeaderOperations{
														Set: map[string]string{
															"CF-App-Id":           "app-guid-1",
															"CF-App-Process-Type": "process-type-1",
															"CF-Space-Id":         "space-guid-0",
															"CF-Organization-Id":  "org-guid-0",
														},
													},
												},
												Weight: 100,
											},
										},
									},
									{
										Match: []*istiov1alpha3.HTTPMatchRequest{
											{
												Uri: &istiov1alpha3.StringMatch{
													MatchType: &istiov1alpha3.StringMatch_Prefix{
														Prefix: "/path0",
													},
												},
											},
										},
										Route: []*istiov1alpha3.HTTPRouteDestination{
											{
												Destination: &istiov1alpha3.Destination{Host: "s-route-0-destination-guid-0"},
												Headers: &istiov1alpha3.Headers{
													Request: &istiov1alpha3.Headers_HeaderOperations{
														Set: map[string]string{
															"CF-App-Id":           "app-guid-0",
															"CF-App-Process-Type": "process-type-1",
															"CF-Space-Id":         "space-guid-0",
															"CF-Organization-Id":  "org-guid-0",
														},
													},
												},
												Weight: 100,
											},
										},
									},
								},
							},
						},
					},
				}

				builder := VirtualServiceBuilder{
					IstioGateways: []string{"some-gateway0", "some-gateway1"},
				}
				Expect(builder.Build(&routes)).To(Equal(expectedVirtualServices))
			})

			Context("and one of the routes has no destinations", func() {
				It("ignores the route without destinations", func() {
					routes = appsv1alpha1.RouteList{Items: []appsv1alpha1.Route{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "route-guid-0",
								Namespace: "workload-namespace",
								Labels: map[string]string{
									"cloudfoundry.org/space_guid": "space-guid-0",
									"cloudfoundry.org/org_guid":   "org-guid-0",
								},
							},
							Spec: appsv1alpha1.RouteSpec{
								Host: "test0",
								Path: "/path0",
								Url:  "test0.domain0.example.com/path0",
								Domain: appsv1alpha1.RouteDomain{
									Name:     "domain0.example.com",
									Internal: false,
								},
								Destinations: []appsv1alpha1.RouteDestination{
									{
										Guid: "route-0-destination-guid-0",
										App: appsv1alpha1.DestinationApp{
											Guid:    "app-guid-0",
											Process: appsv1alpha1.AppProcess{Type: "process-type-1"},
										},
										Port:   intPtr(9000),
										Weight: intPtr(100),
									},
								},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "route-guid-1",
								Namespace: "workload-namespace",
								Labels: map[string]string{
									"cloudfoundry.org/space_guid": "space-guid-0",
									"cloudfoundry.org/org_guid":   "org-guid-0",
								},
							},
							Spec: appsv1alpha1.RouteSpec{
								Host: "test0",
								Path: "/path0/deeper",
								Url:  "test0.domain0.example.com/path0/deeper",
								Domain: appsv1alpha1.RouteDomain{
									Name:     "domain0.example.com",
									Internal: false,
								},
								Destinations: []appsv1alpha1.RouteDestination{},
							},
						},
					},
					}

					builder := VirtualServiceBuilder{
						IstioGateways: []string{"some-gateway0", "some-gateway1"},
					}
					k8sResources := builder.Build(&routes)
					Expect(len(k8sResources)).To(Equal(1))

					virtualservice := k8sResources[0]
					Expect(virtualservice.Spec.Hosts[0]).To(Equal("test0.domain0.example.com"))
					Expect(virtualservice.Spec.Http[0].Match[0].Uri.MatchType).To(BeEquivalentTo(&istiov1alpha3.StringMatch_Prefix{Prefix: "/path0"}))
				})
			})

			Context("and one route is internal and one is external", func() {
				It("does not create a VirtualService for the fqdn", func() {
					routes = appsv1alpha1.RouteList{Items: []appsv1alpha1.Route{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "route-guid-0",
								Namespace: "workload-namespace",
								Labels: map[string]string{
									"cloudfoundry.org/space_guid": "space-guid-0",
									"cloudfoundry.org/org_guid":   "org-guid-0",
								},
							},
							Spec: appsv1alpha1.RouteSpec{
								Host: "test0",
								Path: "/path0",
								Url:  "test0.domain0.example.com/path0",
								Domain: appsv1alpha1.RouteDomain{
									Name:     "domain0.example.com",
									Internal: false,
								},
								Destinations: []appsv1alpha1.RouteDestination{
									{
										Guid: "route-0-destination-guid-0",
										App: appsv1alpha1.DestinationApp{
											Guid:    "app-guid-0",
											Process: appsv1alpha1.AppProcess{Type: "process-type-1"},
										},
										Port:   intPtr(9000),
										Weight: intPtr(100),
									},
								},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "route-guid-0",
								Namespace: "workload-namespace",
								Labels: map[string]string{
									"cloudfoundry.org/space_guid": "space-guid-0",
									"cloudfoundry.org/org_guid":   "org-guid-0",
								},
							},
							Spec: appsv1alpha1.RouteSpec{
								Host: "test0",
								Path: "/path1",
								Url:  "test0.domain0.example.com/path1",
								Domain: appsv1alpha1.RouteDomain{
									Name:     "domain0.example.com",
									Internal: true,
								},
								Destinations: []appsv1alpha1.RouteDestination{
									{
										Guid: "route-0-destination-guid-0",
										App: appsv1alpha1.DestinationApp{
											Guid:    "app-guid-0",
											Process: appsv1alpha1.AppProcess{Type: "process-type-1"},
										},
										Port:   intPtr(9000),
										Weight: intPtr(100),
									},
								},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "route-guid-1",
								Namespace: "workload-namespace",
								Labels: map[string]string{
									"cloudfoundry.org/space_guid": "space-guid-0",
									"cloudfoundry.org/org_guid":   "org-guid-0",
								},
							},
							Spec: appsv1alpha1.RouteSpec{
								Host: "test1",
								Path: "",
								Url:  "test1.domain1.example.com",
								Domain: appsv1alpha1.RouteDomain{
									Name:     "domain1.example.com",
									Internal: false,
								},
								Destinations: []appsv1alpha1.RouteDestination{
									{
										Guid: "route-1-destination-guid-1",
										App: appsv1alpha1.DestinationApp{
											Guid:    "app-guid-1",
											Process: appsv1alpha1.AppProcess{Type: "process-type-1"},
										},
										Port:   intPtr(9000),
										Weight: intPtr(100),
									},
								},
							},
						},
					},
					}

					builder := VirtualServiceBuilder{
						IstioGateways: []string{"some-gateway0", "some-gateway1"},
					}
					k8sResources := builder.Build(&routes)
					Expect(len(k8sResources)).To(Equal(1))

					virtualservices := k8sResources[0]
					Expect(virtualservices.Spec.Hosts[0]).To(Equal("test1.domain1.example.com"))
				})
			})

			Context("and the routes have different namespaces", func() {
				It("does not create a VirtualService for the fqdn", func() {
					routes = appsv1alpha1.RouteList{Items: []appsv1alpha1.Route{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "route-guid-0",
								Namespace: "workload-namespace",
								Labels: map[string]string{
									"cloudfoundry.org/space_guid": "space-guid-0",
									"cloudfoundry.org/org_guid":   "org-guid-0",
								},
							},
							Spec: appsv1alpha1.RouteSpec{
								Host: "test0",
								Path: "/path0",
								Url:  "test0.domain0.example.com/path0",
								Domain: appsv1alpha1.RouteDomain{
									Name:     "domain0.example.com",
									Internal: false,
								},
								Destinations: []appsv1alpha1.RouteDestination{
									{
										Guid: "route-0-destination-guid-0",
										App: appsv1alpha1.DestinationApp{
											Guid:    "app-guid-0",
											Process: appsv1alpha1.AppProcess{Type: "process-type-1"},
										},
										Port:   intPtr(9000),
										Weight: intPtr(100),
									},
								},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "route-guid-0",
								Namespace: "some-different-namespace",
								Labels: map[string]string{
									"cloudfoundry.org/space_guid": "space-guid-0",
									"cloudfoundry.org/org_guid":   "org-guid-0",
								},
							},
							Spec: appsv1alpha1.RouteSpec{
								Host: "test0",
								Path: "/path1",
								Url:  "test0.domain0.example.com/path1",
								Domain: appsv1alpha1.RouteDomain{
									Name:     "domain0.example.com",
									Internal: false,
								},
								Destinations: []appsv1alpha1.RouteDestination{
									{
										Guid: "route-0-destination-guid-0",
										App: appsv1alpha1.DestinationApp{
											Guid:    "app-guid-0",
											Process: appsv1alpha1.AppProcess{Type: "process-type-1"},
										},
										Port:   intPtr(9000),
										Weight: intPtr(100),
									},
								},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "route-guid-1",
								Namespace: "workload-namespace",
								Labels: map[string]string{
									"cloudfoundry.org/space_guid": "space-guid-0",
									"cloudfoundry.org/org_guid":   "org-guid-0",
								},
							},
							Spec: appsv1alpha1.RouteSpec{
								Host: "test1",
								Path: "",
								Url:  "test1.domain1.example.com",
								Domain: appsv1alpha1.RouteDomain{
									Name:     "domain1.example.com",
									Internal: false,
								},
								Destinations: []appsv1alpha1.RouteDestination{
									{
										Guid: "route-1-destination-guid-1",
										App: appsv1alpha1.DestinationApp{
											Guid:    "app-guid-1",
											Process: appsv1alpha1.AppProcess{Type: "process-type-1"},
										},
										Port:   intPtr(9000),
										Weight: intPtr(100),
									},
								},
							},
						},
					},
					}

					builder := VirtualServiceBuilder{
						IstioGateways: []string{"some-gateway0", "some-gateway1"},
					}
					k8sResources := builder.Build(&routes)
					Expect(len(k8sResources)).To(Equal(1))

					virtualservices := k8sResources[0]
					Expect(virtualservices.Spec.Hosts[0]).To(Equal("test1.domain1.example.com"))
				})
			})

			Context("when a route has no destinations", func() {
				It("does not create a VirtualService", func() {
					routes = appsv1alpha1.RouteList{Items: []appsv1alpha1.Route{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "route-guid-0",
								Namespace: "workload-namespace",
								Labels: map[string]string{
									"cloudfoundry.org/space_guid": "space-guid-0",
									"cloudfoundry.org/org_guid":   "org-guid-0",
								},
							},
							Spec: appsv1alpha1.RouteSpec{
								Host: "test0",
								Path: "/path0",
								Url:  "test0.domain0.example.com/path0",
								Domain: appsv1alpha1.RouteDomain{
									Name:     "domain0.example.com",
									Internal: false,
								},
								Destinations: []appsv1alpha1.RouteDestination{},
							},
						},
					},
					}

					builder := VirtualServiceBuilder{
						IstioGateways: []string{"some-gateway0", "some-gateway1"},
					}

					Expect(builder.Build(&routes)).To(BeEmpty())
				})
			})

			Context("when a destination has no weight", func() {
				It("sets the weight to 100", func() {
					routes = appsv1alpha1.RouteList{Items: []appsv1alpha1.Route{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "route-guid-0",
								Namespace: "workload-namespace",
								Labels: map[string]string{
									"cloudfoundry.org/space_guid": "space-guid-0",
									"cloudfoundry.org/org_guid":   "org-guid-0",
								},
							},
							Spec: appsv1alpha1.RouteSpec{
								Host: "test0",
								Path: "/path0",
								Url:  "test0.domain0.example.com/path0",
								Domain: appsv1alpha1.RouteDomain{
									Name:     "domain0.example.com",
									Internal: false,
								},
								Destinations: []appsv1alpha1.RouteDestination{
									{
										Guid: "route-0-destination-guid-0",
										App: appsv1alpha1.DestinationApp{
											Guid:    "app-guid-0",
											Process: appsv1alpha1.AppProcess{Type: "process-type-1"},
										},
										Port: intPtr(9000),
									},
								},
							},
						},
					},
					}

					builder := VirtualServiceBuilder{
						IstioGateways: []string{"some-gateway0", "some-gateway1"},
					}

					k8sResources := builder.Build(&routes)
					Expect(len(k8sResources)).To(Equal(1))

					virtualservices := k8sResources[0]
					Expect(virtualservices.Spec.Http[0].Route[0].Weight).To(Equal(int32(100)))
				})
			})

		})
	})

})

var _ = Describe("VirtualServiceName", func() {
	It("creates consistent and distinct resource names based on FQDN", func() {
		Expect(VirtualServiceName("domain0.example.com")).To(
			Equal("vs-674da971dcc8ee9403167e2a3e77e7a95f609d2825b838fc29a50e48c8dfeaea"))
		Expect(VirtualServiceName("domain1.example.com")).To(
			Equal("vs-68ff4f202372d7fde0b8ef285fa884cf8d88a0b2af81bd0ac0a11d785e06be21"))
	})

	It("removes special characters from FQDNs to create valid k8s resource names", func() {
		Expect(VirtualServiceName("*.wildcard-host.example.com")).To(
			Equal("vs-216d6f90aff241b01b456c94351f77221d9c238057fd4e4394ca5deadc1aae24"))

		Expect(VirtualServiceName("🙂.unicode-host.example.com")).To(
			Equal("vs-3b0a745e60e76cc7f14e5e22d37fc7af2c2b529c5be43e99551d9fa892ca3573"))
	})

	It("condenses long FQDNs to be under 253 characters to create valid k8s resource names", func() {
		const DNSLabelMaxLength = 63
		var longDNSLabel = strings.Repeat("a", DNSLabelMaxLength)

		longFQDN := fmt.Sprintf("%s.%s.%s.%s.example.com", longDNSLabel, longDNSLabel, longDNSLabel, longDNSLabel)
		Expect(VirtualServiceName(longFQDN)).To(
			Equal("vs-b2b7f04662a35e5d54b33c988c8ee4ddfdbcd33c5fbd0eb11e5c011009641015"))
	})
})
