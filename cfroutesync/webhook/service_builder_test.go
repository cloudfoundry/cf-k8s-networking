package webhook_test

import (
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/models"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/webhook"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("ServiceBuilder", func() {
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

	It("returns a Service resource for each route destination", func() {
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
				Url:  "test1.domain1.apps.internal",
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
				Path: "/path0/deeper", // test that longest path matches first
				Url:  "test0.domain0.example.com/path0/deeper",
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
		}

		expectedServices := []webhook.K8sResource{
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
					},
					Annotations: map[string]string{
						"cloudfoundry.org/route-fqdn": "test0.domain0.example.com",
					},
				},
				Spec: webhook.ServiceSpec{
					Selector: map[string]string{
						"custom-pod-label-prefix/app_guid":     "app-guid-0",
						"custom-pod-label-prefix/process_type": "process-type-1",
					},

					Ports: []webhook.ServicePort{
						webhook.ServicePort{
							Port: 9000,
							Name: "http",
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
					},
					Annotations: map[string]string{
						"cloudfoundry.org/route-fqdn": "test0.domain0.example.com",
					},
				},
				Spec: webhook.ServiceSpec{
					Selector: map[string]string{
						"custom-pod-label-prefix/app_guid":     "app-guid-1",
						"custom-pod-label-prefix/process_type": "process-type-1",
					},

					Ports: []webhook.ServicePort{
						webhook.ServicePort{
							Port: 9001,
							Name: "http",
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
					},
					Annotations: map[string]string{
						"cloudfoundry.org/route-fqdn": "test1.domain1.apps.internal",
					},
				},
				Spec: webhook.ServiceSpec{
					Selector: map[string]string{
						"custom-pod-label-prefix/app_guid":     "app-guid-2",
						"custom-pod-label-prefix/process_type": "process-type-2",
					},

					Ports: []webhook.ServicePort{
						webhook.ServicePort{
							Port: 8080,
							Name: "http",
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
					},
					Annotations: map[string]string{
						"cloudfoundry.org/route-fqdn": "test0.domain0.example.com",
					},
				},
				Spec: webhook.ServiceSpec{
					Selector: map[string]string{
						"custom-pod-label-prefix/app_guid":     "app-guid-1",
						"custom-pod-label-prefix/process_type": "process-type-1",
					},

					Ports: []webhook.ServicePort{
						webhook.ServicePort{
							Port: 8080,
							Name: "http",
						},
					},
				},
			},
		}

		builder := webhook.ServiceBuilder{
			PodLabelPrefix: "custom-pod-label-prefix/",
		}
		Expect(builder.Build(routes, template)).To(Equal(expectedServices))
	})

	Context("when a route has no destinations", func() {
		It("does not create a Service", func() {
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

			builder := webhook.ServiceBuilder{}
			Expect(builder.Build(routes, template)).To(Equal([]webhook.K8sResource{}))
		})
	})
})
