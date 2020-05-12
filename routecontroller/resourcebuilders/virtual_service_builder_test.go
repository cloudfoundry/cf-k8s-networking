package resourcebuilders

import (
	"fmt"
	"strings"

	istionetworkingv1alpha3 "code.cloudfoundry.org/cf-k8s-networking/routecontroller/apis/istio/networking/v1alpha3"
	networkingv1alpha1 "code.cloudfoundry.org/cf-k8s-networking/routecontroller/apis/networking/v1alpha1"
	istiov1alpha3 "istio.io/api/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type virtualServiceParams struct {
	fqdn     string
	internal bool
	https    []httpParams
	owners   []ownerParams
}

type httpParams struct {
	matchPrefix  string
	destinations []destParams
}

type ownerParams struct {
	routeName string
	routeUID  types.UID
}

type destParams struct {
	host      string
	appGUID   string
	spaceGUID string
	orgGUID   string
	weight    int32
}

type routeParams struct {
	name         string
	namespace    string
	host         string
	path         string
	domain       string
	internal     bool
	destinations []routeDestParams
}

type routeDestParams struct {
	destGUID string
	port     int
	weight   *int
	appGUID  string
}

func constructVirtualService(params virtualServiceParams) istionetworkingv1alpha3.VirtualService {
	vs := istionetworkingv1alpha3.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      VirtualServiceName(params.fqdn),
			Namespace: "workload-namespace",
			Labels:    map[string]string{},
			Annotations: map[string]string{
				"cloudfoundry.org/fqdn": params.fqdn,
			},
		},
		Spec: istionetworkingv1alpha3.VirtualServiceSpec{
			VirtualService: istiov1alpha3.VirtualService{
				Hosts:    []string{params.fqdn},
				Gateways: []string{"some-gateway0", "some-gateway1"},
			},
		},
	}

	if params.internal {
		vs.Spec.Gateways = []string{"mesh"}
	}

	owners := []metav1.OwnerReference{}
	for _, owner := range params.owners {
		owners = append(owners, metav1.OwnerReference{
			APIVersion: "networking.cloudfoundry.org/v1alpha1",
			Kind:       "Route",
			Name:       owner.routeName,
			UID:        types.UID(owner.routeUID),
		})
	}
	vs.ObjectMeta.OwnerReferences = owners

	https := []*istiov1alpha3.HTTPRoute{}
	for _, http := range params.https {
		httpRoute := istiov1alpha3.HTTPRoute{}
		if http.matchPrefix != "" {
			httpRoute.Match = []*istiov1alpha3.HTTPMatchRequest{
				{
					Uri: &istiov1alpha3.StringMatch{
						MatchType: &istiov1alpha3.StringMatch_Prefix{
							Prefix: http.matchPrefix,
						},
					},
				},
			}
		}

		routes := []*istiov1alpha3.HTTPRouteDestination{}
		for _, dest := range http.destinations {
			routes = append(routes, &istiov1alpha3.HTTPRouteDestination{
				Destination: &istiov1alpha3.Destination{Host: dest.host},
				Headers: &istiov1alpha3.Headers{
					Request: &istiov1alpha3.Headers_HeaderOperations{
						Set: map[string]string{
							"CF-App-Id":           dest.appGUID,
							"CF-App-Process-Type": "process-type-1",
							"CF-Space-Id":         dest.spaceGUID,
							"CF-Organization-Id":  dest.orgGUID,
						},
					},
				},
				Weight: dest.weight,
			})
		}

		httpRoute.Route = routes

		https = append(https, &httpRoute)
	}

	vs.Spec.VirtualService.Http = https
	return vs
}

func constructRoute(params routeParams) networkingv1alpha1.Route {
	if params.namespace == "" {
		params.namespace = "workload-namespace"
	}

	route := networkingv1alpha1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      params.name,
			Namespace: params.namespace,
			Labels: map[string]string{
				"cloudfoundry.org/space_guid": "space-guid-0",
				"cloudfoundry.org/org_guid":   "org-guid-0",
			},
			UID: types.UID(fmt.Sprintf("%s-k8s-uid", params.name)),
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Route",
		},
		Spec: networkingv1alpha1.RouteSpec{
			Host: params.host,
			Path: params.path,
			Url:  fmt.Sprintf("%s.%s%s", params.host, params.domain, params.path),
			Domain: networkingv1alpha1.RouteDomain{
				Name:     params.domain,
				Internal: params.internal,
			},
		},
	}

	destinations := []networkingv1alpha1.RouteDestination{}
	for _, destination := range params.destinations {
		destinations = append(destinations, networkingv1alpha1.RouteDestination{
			Guid:   destination.destGUID,
			Port:   intPtr(destination.port),
			Weight: destination.weight,
			App: networkingv1alpha1.DestinationApp{
				Guid:    destination.appGUID,
				Process: networkingv1alpha1.AppProcess{Type: "process-type-1"},
			},
			Selector: networkingv1alpha1.DestinationSelector{
				MatchLabels: map[string]string{
					"cloudfoundry.org/app_guid":     destination.appGUID,
					"cloudfoundry.org/process_type": "process-type-1",
				},
			},
		})
	}
	route.Spec.Destinations = destinations
	return route
}

var _ = Describe("VirtualServiceBuilder", func() {
	Describe("Build", func() {
		It("returns a VirtualService resource for each route destination", func() {
			routes := networkingv1alpha1.RouteList{
				Items: []networkingv1alpha1.Route{
					constructRoute(routeParams{
						name:   "route-guid-0",
						host:   "test0",
						path:   "/path0",
						domain: "domain0.example.com",
						destinations: []routeDestParams{
							{
								destGUID: "route-0-destination-guid-0",
								port:     9000,
								weight:   intPtr(91),
								appGUID:  "app-guid-0",
							},
							{
								destGUID: "route-0-destination-guid-1",
								port:     9001,
								weight:   intPtr(9),
								appGUID:  "app-guid-1",
							},
						},
					}),
					constructRoute(routeParams{
						name:   "route-guid-1",
						host:   "test1",
						path:   "",
						domain: "domain1.example.com",
						destinations: []routeDestParams{
							{
								destGUID: "route-1-destination-guid-0",
								port:     8080,
								weight:   intPtr(100),
								appGUID:  "app-guid-1",
							},
						},
					}),
				}}

			expectedVirtualServices := []istionetworkingv1alpha3.VirtualService{
				constructVirtualService(virtualServiceParams{
					fqdn: "test0.domain0.example.com",
					owners: []ownerParams{
						{
							routeName: routes.Items[0].ObjectMeta.Name,
							routeUID:  routes.Items[0].ObjectMeta.UID,
						},
					},
					https: []httpParams{
						{
							matchPrefix: "/path0",
							destinations: []destParams{
								{
									host:      "s-route-0-destination-guid-0",
									appGUID:   "app-guid-0",
									spaceGUID: "space-guid-0",
									orgGUID:   "org-guid-0",
									weight:    91,
								},
								{
									host:      "s-route-0-destination-guid-1",
									appGUID:   "app-guid-1",
									spaceGUID: "space-guid-0",
									orgGUID:   "org-guid-0",
									weight:    9,
								},
							},
						},
					},
				}),
				constructVirtualService(virtualServiceParams{
					fqdn: "test1.domain1.example.com",
					owners: []ownerParams{
						{
							routeName: routes.Items[1].ObjectMeta.Name,
							routeUID:  routes.Items[1].ObjectMeta.UID,
						},
					},
					https: []httpParams{
						{
							destinations: []destParams{
								{
									host:      "s-route-1-destination-guid-0",
									appGUID:   "app-guid-1",
									spaceGUID: "space-guid-0",
									orgGUID:   "org-guid-0",
									weight:    100,
								},
							},
						},
					},
				}),
			}

			builder := VirtualServiceBuilder{
				IstioGateways: []string{"some-gateway0", "some-gateway1"},
			}
			virtualservice, err := builder.Build(&routes)
			Expect(err).NotTo(HaveOccurred())
			Expect(virtualservice).To(Equal(expectedVirtualServices))
		})

		Describe("inferring weights", func() {
			var routes networkingv1alpha1.RouteList

			BeforeEach(func() {
				routes = networkingv1alpha1.RouteList{
					Items: []networkingv1alpha1.Route{
						constructRoute(routeParams{
							name:   "route-guid-0",
							host:   "test0",
							path:   "/path0",
							domain: "domain0.example.com",
							destinations: []routeDestParams{
								{
									destGUID: "route-0-destination-guid-0",
									port:     9000,
									appGUID:  "app-guid-0",
								},
								{
									destGUID: "route-0-destination-guid-1",
									port:     8080,
									appGUID:  "app-guid-1",
								},
								{
									destGUID: "route-0-destination-guid-2",
									port:     8080,
									appGUID:  "app-guid-2",
								},
							},
						}),
					},
				}
			})

			Context("when weights aren't present but a route has multiple destinations", func() {
				Context("when the destinations DO NOT evenly divide to 100", func() {
					It("ensures the weights add to 100 and adds any remainder to the first destination", func() {
						builder := VirtualServiceBuilder{
							IstioGateways: []string{"some-gateway0", "some-gateway1"},
						}
						virtualservices, err := builder.Build(&routes)
						Expect(err).NotTo(HaveOccurred())
						Expect(virtualservices[0].Spec.Http[0].Route[0].Weight).To(Equal(int32(34)))
						Expect(virtualservices[0].Spec.Http[0].Route[1].Weight).To(Equal(int32(33)))
						Expect(virtualservices[0].Spec.Http[0].Route[2].Weight).To(Equal(int32(33)))
					})
				})

				Context("when the destinations DO evenly divide to 100", func() {
					It("evenly distributes the weights", func() {
						routes.Items[0].Spec.Destinations = []networkingv1alpha1.RouteDestination{
							{
								Guid: "route-0-destination-guid-0",
								App: networkingv1alpha1.DestinationApp{
									Guid:    "app-guid-0",
									Process: networkingv1alpha1.AppProcess{Type: "process-type-1"},
								},
								Port:   intPtr(9000),
								Weight: nil,
							},
							{
								Guid: "route-0-destination-guid-1",
								App: networkingv1alpha1.DestinationApp{
									Guid:    "app-guid-1",
									Process: networkingv1alpha1.AppProcess{Type: "process-type-1"},
								},
								Port:   intPtr(8080),
								Weight: nil,
							},
						}

						builder := VirtualServiceBuilder{
							IstioGateways: []string{"some-gateway0", "some-gateway1"},
						}
						virtualservices, err := builder.Build(&routes)
						Expect(err).NotTo(HaveOccurred())
						Expect(virtualservices[0].Spec.Http[0].Route[0].Weight).To(Equal(int32(50)))
						Expect(virtualservices[0].Spec.Http[0].Route[1].Weight).To(Equal(int32(50)))
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

						virtualservices, err := builder.Build(&routes)
						Expect(err).NotTo(HaveOccurred())
						Expect(virtualservices[0].Spec.Http[0].Route[0].Weight).To(Equal(int32(70)))
						Expect(virtualservices[0].Spec.Http[0].Route[1].Weight).To(Equal(int32(20)))
						Expect(virtualservices[0].Spec.Http[0].Route[2].Weight).To(Equal(int32(10)))
					})
				})

				Context("when the weights do not sum up to 100", func() {
					It("omits the invalid istionetworkingv1alpha3.VirtualService", func() {
						invalidRoutes := networkingv1alpha1.RouteList{
							Items: []networkingv1alpha1.Route{
								constructRoute(routeParams{
									name:   "route-guid-0",
									host:   "invalid-route",
									path:   "/path0",
									domain: "domain0.example.com",
									destinations: []routeDestParams{
										{
											destGUID: "route-0-destination-guid-0",
											port:     9000,
											weight:   intPtr(80),
											appGUID:  "app-guid-0",
										},
										{
											destGUID: "route-0-destination-guid-1",
											port:     8080,
											weight:   intPtr(80),
											appGUID:  "app-guid-1",
										},
									},
								}),
							},
						}

						routes.Items = append(routes.Items, invalidRoutes.Items[0])

						builder := VirtualServiceBuilder{
							IstioGateways: []string{"some-gateway0", "some-gateway1"},
						}

						_, err := builder.Build(&routes)
						Expect(err).To(MatchError("invalid destinations for route route-guid-0: weights must sum up to 100"))
					})
				})

				Context("when one destination for a given route has a weight but the rest do not", func() {
					BeforeEach(func() {
						invalidRoutes := networkingv1alpha1.RouteList{
							Items: []networkingv1alpha1.Route{
								constructRoute(routeParams{
									name:   "route-guid-0",
									host:   "invalid-route",
									path:   "/path0",
									domain: "invalid-route.domain0.example.com",
									destinations: []routeDestParams{
										{
											destGUID: "route-0-destination-guid-0",
											port:     9000,
											weight:   intPtr(80),
											appGUID:  "app-guid-0",
										},
										{
											destGUID: "route-0-destination-guid-1",
											port:     8080,
											appGUID:  "app-guid-1",
										},
									},
								}),
							},
						}

						routes.Items = append(routes.Items, invalidRoutes.Items[0])
					})

					It("returns an error", func() {
						builder := VirtualServiceBuilder{
							IstioGateways: []string{"some-gateway0", "some-gateway1"},
						}

						_, err := builder.Build(&routes)
						Expect(err).To(MatchError("invalid destinations for route route-guid-0: weights must be set on all or none"))
					})
				})
			})

			Context("when a route is for an internal domain", func() {
				BeforeEach(func() {
					routes = networkingv1alpha1.RouteList{
						Items: []networkingv1alpha1.Route{
							constructRoute(routeParams{
								name:     "route-guid-0",
								host:     "test0",
								path:     "",
								domain:   "domain0.apps.internal",
								internal: true,
								destinations: []routeDestParams{
									{
										weight:  intPtr(100),
										port:    8080,
										appGUID: "app-guid-0",
									},
								},
							}),
						},
					}
				})

				It("uses the internal mesh gateways", func() {
					builder := VirtualServiceBuilder{
						IstioGateways: []string{"some-gateway0", "some-gateway1"},
					}

					virtualservices, err := builder.Build(&routes)
					Expect(err).NotTo(HaveOccurred())
					Expect(len(virtualservices[0].Spec.Gateways)).To(Equal(1))
					Expect(virtualservices[0].Spec.Gateways[0]).To(Equal("mesh"))
				})
			})

			Context("when two routes have the same fqdn", func() {
				BeforeEach(func() {
					routes = networkingv1alpha1.RouteList{
						Items: []networkingv1alpha1.Route{
							constructRoute(routeParams{
								name:     "route-guid-0",
								host:     "test0",
								path:     "/path0",
								domain:   "domain0.example.com",
								internal: true,
								destinations: []routeDestParams{
									{
										destGUID: "route-0-destination-guid-0",
										port:     9000,
										weight:   intPtr(100),
										appGUID:  "app-guid-0",
									},
								},
							}),
							constructRoute(routeParams{
								name:     "route-guid-1",
								host:     "test0",
								path:     "/path0/deeper",
								domain:   "domain0.example.com",
								internal: true,
								destinations: []routeDestParams{
									{
										destGUID: "route-1-destination-guid-0",
										port:     8080,
										weight:   intPtr(100),
										appGUID:  "app-guid-1",
									},
								},
							}),
						},
					}
				})

				It("orders the paths by longest matching prefix", func() {
					expectedVirtualServices := []istionetworkingv1alpha3.VirtualService{
						constructVirtualService(virtualServiceParams{
							fqdn:     "test0.domain0.example.com",
							internal: true,
							owners: []ownerParams{
								{
									routeName: routes.Items[1].ObjectMeta.Name,
									routeUID:  routes.Items[1].ObjectMeta.UID,
								},
								{
									routeName: routes.Items[0].ObjectMeta.Name,
									routeUID:  routes.Items[0].ObjectMeta.UID,
								},
							},
							https: []httpParams{
								{
									matchPrefix: "/path0/deeper",
									destinations: []destParams{
										{
											host:      "s-route-1-destination-guid-0",
											appGUID:   "app-guid-1",
											spaceGUID: "space-guid-0",
											orgGUID:   "org-guid-0",
											weight:    100,
										},
									},
								},
								{
									matchPrefix: "/path0",
									destinations: []destParams{
										{
											host:      "s-route-0-destination-guid-0",
											appGUID:   "app-guid-0",
											spaceGUID: "space-guid-0",
											orgGUID:   "org-guid-0",
											weight:    100,
										},
									},
								},
							},
						}),
					}

					builder := VirtualServiceBuilder{
						IstioGateways: []string{"some-gateway0", "some-gateway1"},
					}
					virtualservice, err := builder.Build(&routes)
					Expect(err).NotTo(HaveOccurred())
					Expect(virtualservice).To(Equal(expectedVirtualServices))
				})

				Context("and one of the routes has no destinations", func() {
					It("ignores the route without destinations", func() {
						routes = networkingv1alpha1.RouteList{
							Items: []networkingv1alpha1.Route{
								constructRoute(routeParams{
									name:     "route-guid-0",
									host:     "test0",
									path:     "/path0",
									domain:   "domain0.example.com",
									internal: true,
									destinations: []routeDestParams{
										{
											destGUID: "route-0-destination-guid-0",
											port:     9000,
											weight:   intPtr(100),
											appGUID:  "app-guid-0",
										},
									},
								}),
								constructRoute(routeParams{
									name:         "route-guid-1",
									host:         "test0",
									path:         "/path0/deeper",
									domain:       "domain0.example.com",
									internal:     true,
									destinations: []routeDestParams{},
								}),
							},
						}

						builder := VirtualServiceBuilder{
							IstioGateways: []string{"some-gateway0", "some-gateway1"},
						}
						k8sResources, err := builder.Build(&routes)
						Expect(err).NotTo(HaveOccurred())
						Expect(len(k8sResources)).To(Equal(1))

						virtualservice := k8sResources[0]
						Expect(virtualservice.Spec.Hosts[0]).To(Equal("test0.domain0.example.com"))
						Expect(virtualservice.Spec.Http[0].Match[0].Uri.MatchType).To(BeEquivalentTo(&istiov1alpha3.StringMatch_Prefix{Prefix: "/path0"}))
					})
				})

				Context("and one route is internal and one is external", func() {
					It("does not create a VirtualService for the fqdn", func() {
						routes = networkingv1alpha1.RouteList{
							Items: []networkingv1alpha1.Route{
								constructRoute(routeParams{
									name:     "route-guid-0",
									host:     "test0",
									path:     "/path0",
									domain:   "domain0.example.com",
									internal: false,
									destinations: []routeDestParams{
										{
											destGUID: "route-0-destination-guid-0",
											port:     9000,
											weight:   intPtr(100),
											appGUID:  "app-guid-0",
										},
									},
								}),
								constructRoute(routeParams{
									name:     "route-guid-1",
									host:     "test0",
									path:     "/path1",
									domain:   "domain0.example.com",
									internal: true,
									destinations: []routeDestParams{
										{
											destGUID: "route-1-destination-guid-0",
											port:     9000,
											weight:   intPtr(100),
											appGUID:  "app-guid-0",
										},
									},
								}),
								constructRoute(routeParams{
									name:     "route-guid-2",
									host:     "test1",
									path:     "",
									domain:   "domain1.example.com",
									internal: false,
									destinations: []routeDestParams{
										{
											destGUID: "route-2-destination-guid-0",
											port:     9000,
											weight:   intPtr(100),
											appGUID:  "app-guid-1",
										},
									},
								}),
							},
						}

						builder := VirtualServiceBuilder{
							IstioGateways: []string{"some-gateway0", "some-gateway1"},
						}
						_, err := builder.Build(&routes)
						Expect(err).To(MatchError("route guid route-guid-0 and route guid route-guid-1 disagree on whether or not the domain is internal"))
					})
				})

				Context("and the routes have different namespaces", func() {
					It("does not create a VirtualService for the fqdn", func() {
						routes = networkingv1alpha1.RouteList{
							Items: []networkingv1alpha1.Route{
								constructRoute(routeParams{
									name:     "route-guid-0",
									host:     "test0",
									path:     "/path0",
									domain:   "domain0.example.com",
									internal: false,
									destinations: []routeDestParams{
										{
											destGUID: "route-0-destination-guid-0",
											port:     9000,
											weight:   intPtr(100),
											appGUID:  "app-guid-0",
										},
									},
								}),
								constructRoute(routeParams{
									name:      "route-guid-1",
									namespace: "some-different-namespace",
									host:      "test0",
									path:      "/path1",
									domain:    "domain0.example.com",
									internal:  false,
									destinations: []routeDestParams{
										{
											destGUID: "route-1-destination-guid-0",
											port:     9000,
											weight:   intPtr(100),
											appGUID:  "app-guid-0",
										},
									},
								}),
								constructRoute(routeParams{
									name:     "route-guid-2",
									host:     "test1",
									path:     "",
									domain:   "domain1.example.com",
									internal: false,
									destinations: []routeDestParams{
										{
											destGUID: "route-2-destination-guid-0",
											port:     9000,
											weight:   intPtr(100),
											appGUID:  "app-guid-1",
										},
									},
								}),
							},
						}

						builder := VirtualServiceBuilder{
							IstioGateways: []string{"some-gateway0", "some-gateway1"},
						}
						_, err := builder.Build(&routes)
						Expect(err).To(MatchError("route guid route-guid-0 and route guid route-guid-1 share the same FQDN but have different namespaces"))
					})
				})

				Context("when a route has no destinations", func() {
					It("does not create a VirtualService", func() {
						routes = networkingv1alpha1.RouteList{
							Items: []networkingv1alpha1.Route{
								constructRoute(routeParams{
									name:         "route-guid-0",
									host:         "test0",
									path:         "/path0",
									domain:       "domain0.example.com",
									internal:     false,
									destinations: []routeDestParams{},
								}),
							},
						}

						builder := VirtualServiceBuilder{
							IstioGateways: []string{"some-gateway0", "some-gateway1"},
						}

						virtualservice, err := builder.Build(&routes)
						Expect(err).NotTo(HaveOccurred())
						Expect(virtualservice).To(BeEmpty())
					})
				})

				Context("when a destination has no weight", func() {
					It("sets the weight to 100", func() {
						routes = networkingv1alpha1.RouteList{
							Items: []networkingv1alpha1.Route{
								constructRoute(routeParams{
									name:     "route-guid-0",
									host:     "test0",
									path:     "/path0",
									domain:   "domain0.example.com",
									internal: false,
									destinations: []routeDestParams{
										{
											destGUID: "route-0-destination-guid-0",
											port:     9000,
											appGUID:  "app-guid-1",
										},
									},
								}),
							},
						}

						builder := VirtualServiceBuilder{
							IstioGateways: []string{"some-gateway0", "some-gateway1"},
						}

						k8sResources, err := builder.Build(&routes)
						Expect(err).NotTo(HaveOccurred())
						Expect(len(k8sResources)).To(Equal(1))

						virtualservices := k8sResources[0]
						Expect(virtualservices.Spec.Http[0].Route[0].Weight).To(Equal(int32(100)))
					})
				})
			})
		})
	})

	Describe("BuildMutateFunction", func() {
		It("builds a mutate function that copies desired state to actual resource", func() {
			actualVirtualService := &istionetworkingv1alpha3.VirtualService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      VirtualServiceName("test0.domain0.example.com"),
					Namespace: "workload-namespace",
					UID:       "some-uid",
				},
			}

			desiredVirtualService := constructVirtualService(virtualServiceParams{
				fqdn: "test0.domain0.example.com",
				owners: []ownerParams{
					{
						routeName: "banana",
						routeUID:  "ham-ding-er",
					},
				},
				https: []httpParams{
					{
						matchPrefix: "/path0",
						destinations: []destParams{
							{
								host:      "s-route-0-destination-guid-0",
								appGUID:   "app-guid-0",
								spaceGUID: "space-guid-0",
								orgGUID:   "org-guid-0",
								weight:    100,
							},
						},
					},
				},
			})

			Expect(len(actualVirtualService.ObjectMeta.OwnerReferences)).To(BeZero())

			builder := VirtualServiceBuilder{}
			mutateFn := builder.BuildMutateFunction(actualVirtualService, &desiredVirtualService)
			err := mutateFn()
			Expect(err).NotTo(HaveOccurred())

			Expect(actualVirtualService.ObjectMeta.Name).To(Equal(VirtualServiceName("test0.domain0.example.com")))
			Expect(actualVirtualService.ObjectMeta.Namespace).To(Equal("workload-namespace"))
			Expect(actualVirtualService.ObjectMeta.UID).To(Equal(types.UID("some-uid")))
			Expect(actualVirtualService.ObjectMeta.Labels).To(Equal(desiredVirtualService.ObjectMeta.Labels))
			Expect(actualVirtualService.ObjectMeta.Annotations).To(Equal(desiredVirtualService.ObjectMeta.Annotations))
			Expect(len(actualVirtualService.ObjectMeta.OwnerReferences)).NotTo(BeZero())
			Expect(actualVirtualService.ObjectMeta.OwnerReferences).To(Equal(desiredVirtualService.ObjectMeta.OwnerReferences))
			Expect(actualVirtualService.Spec).To(Equal(desiredVirtualService.Spec))
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
