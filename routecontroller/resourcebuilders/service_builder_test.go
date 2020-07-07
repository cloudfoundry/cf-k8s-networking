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

func constructService(params serviceParams) corev1.Service {
	result := corev1.Service{}

	result.ObjectMeta = metav1.ObjectMeta{
		Name:      fmt.Sprintf("s-%s", params.destGUID),
		Namespace: "workload-namespace",
	}

	if params.appGUID != "" && params.processType != "" {
		result.ObjectMeta.Labels = map[string]string{
			"cloudfoundry.org/route_guid":   params.routeGUID,
			"cloudfoundry.org/app_guid":     params.appGUID,
			"cloudfoundry.org/process_type": params.processType,
		}
		result.Spec.Selector = map[string]string{
			"cloudfoundry.org/app_guid":     params.appGUID,
			"cloudfoundry.org/process_type": params.processType,
		}
	}

	if params.serviceUID != "" {
		result.ObjectMeta.UID = types.UID(params.serviceUID)
	}

	if params.fqdn != "" {
		result.ObjectMeta.Annotations = map[string]string{
			"cloudfoundry.org/route-fqdn": params.fqdn,
		}
	}

	if params.routeGUID != "" && params.routeUID != "" {
		result.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
			{
				APIVersion: "networking.cloudfoundry.org/v1alpha1",
				Kind:       "Route",
				Name:       params.routeGUID,
				UID:        types.UID(params.routeUID),
			},
		}
	}

	if params.clusterIP != "" {
		result.Spec.ClusterIP = params.clusterIP
	}

	if params.port != 0 {
		result.Spec.Ports = []corev1.ServicePort{
			{
				Port: params.port,
				Name: "http",
			},
		}
	}

	return result
}

var _ = Describe("ServiceBuilder", func() {
	Describe("Build", func() {
		It("returns a Service resource for each route destination", func() {
			route := networkingv1alpha1.RouteList{
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
				}}

			expectedServices := []corev1.Service{}
			expectedService1Params := serviceParams{
				fqdn:        "test0.domain0.example.com",
				destGUID:    "route-0-destination-guid-0",
				routeGUID:   route.Items[0].ObjectMeta.Name,
				routeUID:    string(route.Items[0].ObjectMeta.UID),
				appGUID:     "app-guid-0",
				processType: "process-type-1",
				port:        9000,
			}
			expectedService2Params := serviceParams{
				fqdn:        "test0.domain0.example.com",
				destGUID:    "route-0-destination-guid-1",
				routeGUID:   route.Items[0].ObjectMeta.Name,
				routeUID:    string(route.Items[0].ObjectMeta.UID),
				appGUID:     "app-guid-1",
				processType: "process-type-1",
				port:        9001,
			}

			expectedServices = append(expectedServices, constructService(expectedService1Params))
			expectedServices = append(expectedServices, constructService(expectedService2Params))

			builder := ServiceBuilder{}

			Expect(builder.Build(&route.Items[0])).To(Equal(expectedServices))
		})

		Context("when a route has no destinations", func() {
			It("does not create a Service", func() {
				route := networkingv1alpha1.RouteList{
					Items: []networkingv1alpha1.Route{
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

				builder := ServiceBuilder{}
				Expect(builder.Build(&route.Items[0])).To(BeEmpty())
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
