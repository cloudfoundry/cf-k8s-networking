package resourcebuilders

import (
	"fmt"
	"strings"

	networkingv1alpha1 "code.cloudfoundry.org/cf-k8s-networking/routecontroller/apis/networking/v1alpha1"
	hpv1 "github.com/projectcontour/contour/apis/projectcontour/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func constructHTTPProxy(params ingressResourceParams) hpv1.HTTPProxy {
	hp := hpv1.HTTPProxy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      HTTPProxyName(params.fqdn),
			Namespace: "workload-namespace",
			Labels:    map[string]string{},
			Annotations: map[string]string{
				"cloudfoundry.org/fqdn": params.fqdn,
			},
		},
		Spec: hpv1.HTTPProxySpec{
			VirtualHost: &hpv1.VirtualHost{
				Fqdn: params.fqdn,
			},
		},
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
	hp.ObjectMeta.OwnerReferences = owners

	if params.https != nil {
		hpRoutes := []hpv1.Route{}
		for _, http := range params.https {
			hpRoute := hpv1.Route{}
			if http.matchPrefix != "" {
				hpRoute.Conditions = []hpv1.MatchCondition{
					{
						Prefix: http.matchPrefix,
					},
				}
			}

			hpServices := []hpv1.Service{}
			for _, dest := range http.destinations {
				hpService := hpv1.Service{
					Name:   dest.host,
					Port:   8080,
					Weight: int64(dest.weight),
				}

				if !dest.noHeadersSet {
					hpService.RequestHeadersPolicy = &hpv1.HeadersPolicy{
						Set: []hpv1.HeaderValue{{
							Name:  "CF-App-Id",
							Value: dest.appGUID,
						}, {
							Name:  "CF-Space-Id",
							Value: dest.spaceGUID,
						}, {
							Name:  "CF-Organization-Id",
							Value: dest.orgGUID,
						}, {
							Name:  "CF-App-Process-Type",
							Value: "process-type-1",
						}},
					}

				}

				hpServices = append(hpServices, hpService)
			}

			hpRoute.Services = hpServices
			hpRoutes = append(hpRoutes, hpRoute)
		}

		hp.Spec.Routes = hpRoutes
	}

	return hp
}

var _ = Describe("HTTPProxyBuilder", func() {
	Describe("Build", func() {
		It("returns a HTTPProxy resource for each route destination", func() {
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

			expectedHTTPProxy := []hpv1.HTTPProxy{
				constructHTTPProxy(ingressResourceParams{
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
				constructHTTPProxy(ingressResourceParams{
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

			builder := HTTPProxyBuilder{}
			httpProxy, err := builder.Build(&routes)
			Expect(err).NotTo(HaveOccurred())
			Expect(httpProxy).To(Equal(expectedHTTPProxy))
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

			Context("when weights are present", func() {
				It("leaves the weights alone", func() {
					routes.Items[0].Spec.Destinations[0].Weight = intPtr(7)
					routes.Items[0].Spec.Destinations[1].Weight = intPtr(2)
					routes.Items[0].Spec.Destinations[2].Weight = intPtr(1)

					builder := HTTPProxyBuilder{}

					httpProxies, err := builder.Build(&routes)
					Expect(err).NotTo(HaveOccurred())
					Expect(httpProxies[0].Spec.Routes[0].Services[0].Weight).To(Equal(int64(7)))
					Expect(httpProxies[0].Spec.Routes[0].Services[1].Weight).To(Equal(int64(2)))
					Expect(httpProxies[0].Spec.Routes[0].Services[2].Weight).To(Equal(int64(1)))
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
						builder := HTTPProxyBuilder{}

						_, err := builder.Build(&routes)
						Expect(err).To(MatchError("invalid destinations for route route-guid-0: weights must be set on all or none"))
					})
				})
			})
		})

		Context("when two routes have the same fqdn", func() {
			var routes networkingv1alpha1.RouteList
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
				expectedHTTPProxies := []hpv1.HTTPProxy{
					constructHTTPProxy(ingressResourceParams{
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

				builder := HTTPProxyBuilder{}
				httpProxies, err := builder.Build(&routes)
				Expect(err).NotTo(HaveOccurred())
				Expect(httpProxies).To(Equal(expectedHTTPProxies))
			})

			Context("and one of the routes has no destinations", func() {
				It("sets a placeholder destination to that route", func() {
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

					builder := HTTPProxyBuilder{}
					k8sResources, err := builder.Build(&routes)
					Expect(err).NotTo(HaveOccurred())
					Expect(len(k8sResources)).To(Equal(1))

					httpProxy := k8sResources[0]
					Expect(httpProxy.Spec.VirtualHost.Fqdn).To(Equal("test0.domain0.example.com"))
					Expect(httpProxy.Spec.Routes[0].Conditions[0].Prefix).To(Equal("/path0"))
				})
			})

			Context("and one route is internal and one is external", func() {
				It("does not create a HTTPProxy for the fqdn", func() {
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

					builder := HTTPProxyBuilder{}
					_, err := builder.Build(&routes)
					Expect(err).To(MatchError("route guid route-guid-0 and route guid route-guid-1 disagree on whether or not the domain is internal"))
				})
			})

			Context("and the routes have different namespaces", func() {
				It("does not create a HTTPProxy for the fqdn", func() {
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

					builder := HTTPProxyBuilder{}
					_, err := builder.Build(&routes)
					Expect(err).To(MatchError("route guid route-guid-0 and route guid route-guid-1 share the same FQDN but have different namespaces"))
				})
			})

			Context("when a route has no destinations", func() {
				It("creates a HTTPProxies with a placeholder destination", func() {
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

					builder := HTTPProxyBuilder{}

					expectedHTTPProxies := []hpv1.HTTPProxy{
						constructHTTPProxy(ingressResourceParams{
							fqdn: "test0.domain0.example.com",
							owners: []ownerParams{
								{
									routeName: routes.Items[0].ObjectMeta.Name,
									routeUID:  routes.Items[0].ObjectMeta.UID,
								},
							},
							https: []httpParams{
								{
									destinations: []destParams{
										{
											host:         "no-destinations",
											noHeadersSet: true,
										},
									},
									matchPrefix: "/path0",
								},
							},
						}),
					}

					httpProxies, err := builder.Build(&routes)
					Expect(err).NotTo(HaveOccurred())
					Expect(httpProxies).To(Equal(expectedHTTPProxies))
				})
			})

		})
	})

	Describe("BuildMutateFunction", func() {
		It("builds a mutate function that copies desired state to actual resource", func() {
			actualHTTPProxy := &hpv1.HTTPProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      HTTPProxyName("test0.domain0.example.com"),
					Namespace: "workload-namespace",
					UID:       "some-uid",
				},
			}

			desiredHTTPProxy := constructHTTPProxy(ingressResourceParams{
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

			Expect(len(actualHTTPProxy.ObjectMeta.OwnerReferences)).To(BeZero())

			builder := HTTPProxyBuilder{}
			mutateFn := builder.BuildMutateFunction(actualHTTPProxy, &desiredHTTPProxy)
			err := mutateFn()
			Expect(err).NotTo(HaveOccurred())

			Expect(actualHTTPProxy.ObjectMeta.Name).To(Equal(HTTPProxyName("test0.domain0.example.com")))
			Expect(actualHTTPProxy.ObjectMeta.Namespace).To(Equal("workload-namespace"))
			Expect(actualHTTPProxy.ObjectMeta.UID).To(Equal(types.UID("some-uid")))
			Expect(actualHTTPProxy.ObjectMeta.Labels).To(Equal(desiredHTTPProxy.ObjectMeta.Labels))
			Expect(actualHTTPProxy.ObjectMeta.Annotations).To(Equal(desiredHTTPProxy.ObjectMeta.Annotations))
			Expect(len(actualHTTPProxy.ObjectMeta.OwnerReferences)).NotTo(BeZero())
			Expect(actualHTTPProxy.ObjectMeta.OwnerReferences).To(Equal(desiredHTTPProxy.ObjectMeta.OwnerReferences))
			Expect(actualHTTPProxy.Spec).To(Equal(desiredHTTPProxy.Spec))
		})
	})
})

var _ = Describe("HTTPProxyName", func() {
	It("creates consistent and distinct resource names based on FQDN", func() {
		Expect(HTTPProxyName("domain0.example.com")).To(
			Equal("674da971dcc8ee9403167e2a3e77e7a95f609d2825b838fc29a50e48c8dfeaea"))
		Expect(HTTPProxyName("domain1.example.com")).To(
			Equal("68ff4f202372d7fde0b8ef285fa884cf8d88a0b2af81bd0ac0a11d785e06be21"))
	})

	It("removes special characters from FQDNs to create valid k8s resource names", func() {
		Expect(HTTPProxyName("*.wildcard-host.example.com")).To(
			Equal("216d6f90aff241b01b456c94351f77221d9c238057fd4e4394ca5deadc1aae24"))
		Expect(HTTPProxyName("🌝.unicode-host.example.com")).To(
			Equal("e314c272f97459e55cc20dbf4794b82d386bfdbc334745df5c9eab08161bc811"))
	})

	It("condenses long FQDNs to be under 253 characters to create valid k8s resource names", func() {
		const DNSLabelMaxLength = 63
		var longDNSLabel = strings.Repeat("a", DNSLabelMaxLength)

		longFQDN := fmt.Sprintf("%s.%s.%s.%s.example.com", longDNSLabel, longDNSLabel, longDNSLabel, longDNSLabel)
		Expect(HTTPProxyName(longFQDN)).To(
			Equal("b2b7f04662a35e5d54b33c988c8ee4ddfdbcd33c5fbd0eb11e5c011009641015"))
	})
})
