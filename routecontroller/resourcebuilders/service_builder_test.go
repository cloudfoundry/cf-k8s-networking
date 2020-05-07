package resourcebuilders

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	networkingv1alpha1 "code.cloudfoundry.org/cf-k8s-networking/routecontroller/apis/networking/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type serviceParams struct {
	fqdn        string
	destGUID    string
	routeGUID   string
	routeUID    string
	appGUID     string
	processType string
	port        int32
	clusterIP   string
	serviceUID  string
}

var _ = Describe("ServiceBuilder", func() {
	Describe("Build", func() {
		It("returns a Service resource for each route destination", func() {
			route := routeWithMultipleDestinations()

			expectedServices := []corev1.Service{}
			expectedService1Params := serviceParams{
				fqdn:        "test0.domain0.example.com",
				destGUID:    "route-0-destination-guid-0",
				routeGUID:   route.ObjectMeta.Name,
				routeUID:    string(route.ObjectMeta.UID),
				appGUID:     "app-guid-0",
				processType: "process-type-0",
				port:        9000,
			}
			expectedService2Params := serviceParams{
				fqdn:        "test0.domain0.example.com",
				destGUID:    "route-0-destination-guid-1",
				routeGUID:   route.ObjectMeta.Name,
				routeUID:    string(route.ObjectMeta.UID),
				appGUID:     "app-guid-1",
				processType: "process-type-1",
				port:        9001,
			}

			expectedServices = append(expectedServices, constructService(expectedService1Params))
			expectedServices = append(expectedServices, constructService(expectedService2Params))

			builder := ServiceBuilder{}

			Expect(builder.Build(&route)).To(Equal(expectedServices))
		})

		Context("when a route has no destinations", func() {
			It("does not create a Service", func() {
				route := routeWithNoDestinations()

				builder := ServiceBuilder{}
				Expect(builder.Build(&route)).To(BeEmpty())
			})
		})
	})

	Describe("BuildMutateFunction", func() {
		It("builds a mutate function that copies desired state to actual resource", func() {
			actualServiceParams := serviceParams{
				destGUID:   "route-0-destination-guid-1",
				serviceUID: "some-uid",
				clusterIP:  "1.2.3.4",
			}

			desiredServiceParams := serviceParams{
				fqdn:        "test0.domain0.example.com",
				destGUID:    "route-0-destination-guid-1",
				routeGUID:   "routey-boi",
				routeUID:    "asdfa-adfsdf-fdsfdsf",
				appGUID:     "app-guid-1",
				processType: "process-type-1",
				port:        9001,
			}
			actualService := constructService(actualServiceParams)
			desiredService := constructService(desiredServiceParams)

			Expect(len(actualService.ObjectMeta.OwnerReferences)).To(BeZero())
			builder := ServiceBuilder{}
			mutateFn := builder.BuildMutateFunction(&actualService, &desiredService)
			err := mutateFn()
			Expect(err).NotTo(HaveOccurred())

			Expect(actualService.ObjectMeta.Name).To(Equal("s-route-0-destination-guid-1"))
			Expect(actualService.ObjectMeta.Namespace).To(Equal("workload-namespace"))
			Expect(actualService.ObjectMeta.UID).To(Equal(types.UID("some-uid")))
			Expect(actualService.ObjectMeta.Labels).To(Equal(desiredService.ObjectMeta.Labels))
			Expect(actualService.ObjectMeta.Annotations).To(Equal(desiredService.ObjectMeta.Annotations))
			Expect(len(actualService.ObjectMeta.OwnerReferences)).NotTo(BeZero())
			Expect(actualService.ObjectMeta.OwnerReferences).To(Equal(desiredService.ObjectMeta.OwnerReferences))
			Expect(actualService.Spec).To(Equal(corev1.ServiceSpec{
				ClusterIP: "1.2.3.4",
				Selector: map[string]string{
					"cloudfoundry.org/app_guid":     "app-guid-1",
					"cloudfoundry.org/process_type": "process-type-1",
				},
				Ports: []corev1.ServicePort{
					{
						Port: 9001,
						Name: "http",
					},
				},
			}))
		})
	})
})

func constructService(p serviceParams) corev1.Service {
	result := corev1.Service{}

	result.ObjectMeta = metav1.ObjectMeta{
		Name:      fmt.Sprintf("s-%s", p.destGUID),
		Namespace: "workload-namespace",
	}

	if p.appGUID != "" && p.processType != "" {
		result.ObjectMeta.Labels = map[string]string{
			"cloudfoundry.org/route_guid":   p.routeGUID,
			"cloudfoundry.org/app_guid":     p.appGUID,
			"cloudfoundry.org/process_type": p.processType,
		}
		result.Spec.Selector = map[string]string{
			"cloudfoundry.org/app_guid":     p.appGUID,
			"cloudfoundry.org/process_type": p.processType,
		}
	}

	if p.serviceUID != "" {
		result.ObjectMeta.UID = types.UID(p.serviceUID)
	}

	if p.fqdn != "" {
		result.ObjectMeta.Annotations = map[string]string{
			"cloudfoundry.org/route-fqdn": p.fqdn,
		}
	}

	if p.routeGUID != "" && p.routeUID != "" {
		result.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
			metav1.OwnerReference{
				APIVersion: "networking.cloudfoundry.org/v1alpha1",
				Kind:       "Route",
				Name:       p.routeGUID,
				UID:        types.UID(p.routeUID),
			},
		}
	}

	if p.clusterIP != "" {
		result.Spec.ClusterIP = p.clusterIP
	}

	if p.port != 0 {
		result.Spec.Ports = []corev1.ServicePort{
			{
				Port: p.port,
				Name: "http",
			},
		}
	}

	return result
}

func routeWithMultipleDestinations() networkingv1alpha1.Route {
	return networkingv1alpha1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "route-guid-0",
			Namespace: "workload-namespace",
			Labels: map[string]string{
				"cloudfoundry.org/space_guid": "space-guid-0",
				"cloudfoundry.org/org_guid":   "org-guid-0",
			},
			UID: "route-guid-0-k8s-uid",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Route",
		},
		Spec: networkingv1alpha1.RouteSpec{
			Host: "test0",
			Path: "/path0",
			Url:  "test0.domain0.example.com/path0",
			Domain: networkingv1alpha1.RouteDomain{
				Name:     "domain0.example.com",
				Internal: false,
			},
			Destinations: []networkingv1alpha1.RouteDestination{
				networkingv1alpha1.RouteDestination{
					Guid:   "route-0-destination-guid-0",
					Port:   intPtr(9000),
					Weight: intPtr(91),
					App: networkingv1alpha1.DestinationApp{
						Guid:    "app-guid-0",
						Process: networkingv1alpha1.AppProcess{Type: "process-type-0"},
					},
					Selector: networkingv1alpha1.DestinationSelector{
						MatchLabels: map[string]string{
							"cloudfoundry.org/app_guid":     "app-guid-0",
							"cloudfoundry.org/process_type": "process-type-0",
						},
					},
				},
				networkingv1alpha1.RouteDestination{
					Guid:   "route-0-destination-guid-1",
					Port:   intPtr(9001),
					Weight: intPtr(9),
					App: networkingv1alpha1.DestinationApp{
						Guid:    "app-guid-1",
						Process: networkingv1alpha1.AppProcess{Type: "process-type-1"},
					},
					Selector: networkingv1alpha1.DestinationSelector{
						MatchLabels: map[string]string{
							"cloudfoundry.org/app_guid":     "app-guid-1",
							"cloudfoundry.org/process_type": "process-type-1",
						},
					},
				},
			},
		},
	}
}

func routeWithNoDestinations() networkingv1alpha1.Route {
	route := routeWithMultipleDestinations()
	destinations := []networkingv1alpha1.RouteDestination{}
	route.Spec.Destinations = destinations
	return route
}
