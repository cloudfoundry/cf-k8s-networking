package resourcebuilders

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	networkingv1alpha1 "code.cloudfoundry.org/cf-k8s-networking/routecontroller/apis/networking/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("ServiceBuilder", func() {
	Describe("Build", func() {
		It("returns a Service resource for each route destination", func() {
			route := networkingv1alpha1.Route{
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
								Process: networkingv1alpha1.AppProcess{Type: "process-type-1"},
							},
							Selector: networkingv1alpha1.DestinationSelector{
								MatchLabels: map[string]string{
									"cloudfoundry.org/app_guid":     "app-guid-0",
									"cloudfoundry.org/process_type": "process-type-1",
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

			expectedServices := []corev1.Service{
				corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						OwnerReferences: []metav1.OwnerReference{
							metav1.OwnerReference{
								APIVersion: "networking.cloudfoundry.org/v1alpha1",
								Kind:       "Route",
								Name:       route.ObjectMeta.Name,
								UID:        route.ObjectMeta.UID,
							},
						},
						Name:      "s-route-0-destination-guid-0",
						Namespace: "workload-namespace",
						Labels: map[string]string{
							"cloudfoundry.org/route_guid":   "route-guid-0",
							"cloudfoundry.org/app_guid":     "app-guid-0",
							"cloudfoundry.org/process_type": "process-type-1",
						},
						Annotations: map[string]string{
							"cloudfoundry.org/route-fqdn": "test0.domain0.example.com",
						},
					},
					Spec: corev1.ServiceSpec{
						Selector: map[string]string{
							"cloudfoundry.org/app_guid":     "app-guid-0",
							"cloudfoundry.org/process_type": "process-type-1",
						},

						Ports: []corev1.ServicePort{
							{
								Port: 9000,
								Name: "http",
							},
						},
					},
				},
				corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name: "s-route-0-destination-guid-1",
						OwnerReferences: []metav1.OwnerReference{
							metav1.OwnerReference{
								APIVersion: "networking.cloudfoundry.org/v1alpha1",
								Kind:       "Route",
								Name:       route.ObjectMeta.Name,
								UID:        route.ObjectMeta.UID,
							},
						},
						Namespace: "workload-namespace",
						Labels: map[string]string{
							"cloudfoundry.org/route_guid":   "route-guid-0",
							"cloudfoundry.org/app_guid":     "app-guid-1",
							"cloudfoundry.org/process_type": "process-type-1",
						},
						Annotations: map[string]string{
							"cloudfoundry.org/route-fqdn": "test0.domain0.example.com",
						},
					},
					Spec: corev1.ServiceSpec{
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
					},
				},
			}

			builder := ServiceBuilder{}

			Expect(builder.Build(&route)).To(Equal(expectedServices))
		})

		Context("when a route has no destinations", func() {
			It("does not create a Service", func() {
				route := networkingv1alpha1.Route{
					TypeMeta: metav1.TypeMeta{Kind: "Route"},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "route-guid-0",
						Namespace: "workload-namespace",
						Labels: map[string]string{
							"cloudfoundry.org/space_guid": "space-guid-0",
							"cloudfoundry.org/org_guid":   "org-guid-0",
						},
						UID: "route-guid-0-k8s-uid",
					},
					Spec: networkingv1alpha1.RouteSpec{
						Host: "test0",
						Path: "/path0",
						Url:  "test0.domain0.example.com/path0",
						Domain: networkingv1alpha1.RouteDomain{
							Name:     "domain0.example.com",
							Internal: false,
						},
						Destinations: []networkingv1alpha1.RouteDestination{},
					},
				}

				builder := ServiceBuilder{}
				Expect(builder.Build(&route)).To(BeEmpty())
			})
		})
	})

	Describe("BuildMutateFunction", func() {
		It("builds a mutate function that copies desired state to actual resource", func() {
			actualService := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "s-route-0-destination-guid-1",
					Namespace:       "workload-namespace",
					UID:             "some-uid",
					OwnerReferences: []metav1.OwnerReference{},
				},
				Spec: corev1.ServiceSpec{
					ClusterIP: "1.2.3.4",
				},
			}

			desiredService := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "s-route-0-destination-guid-1",
					Namespace: "workload-namespace",
					Labels: map[string]string{
						"cloudfoundry.org/route_guid":   "route-guid-0",
						"cloudfoundry.org/app_guid":     "app-guid-1",
						"cloudfoundry.org/process_type": "process-type-1",
					},
					Annotations: map[string]string{
						"cloudfoundry.org/route-fqdn": "test0.domain0.example.com",
					},
					OwnerReferences: []metav1.OwnerReference{
						metav1.OwnerReference{
							APIVersion: "networking.cloudfoundry.org/v1alpha1",
							Kind:       "Route",
							Name:       "routey-boi",
							UID:        "asdfa-adfsdf-fdsfdsf",
						},
					},
				},
				Spec: corev1.ServiceSpec{
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
				},
			}

			Expect(len(actualService.ObjectMeta.OwnerReferences)).To(BeZero())
			builder := ServiceBuilder{}
			mutateFn := builder.BuildMutateFunction(actualService, desiredService)
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
