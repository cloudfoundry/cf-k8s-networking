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
		fakeSnapshotRepo *fakes.SnapshotRepo
		syncHandler      *webhook.Lineage
	)

	BeforeEach(func() {
		fakeSnapshotRepo = &fakes.SnapshotRepo{}
		syncHandler = &webhook.Lineage{
			RouteSnapshotRepo: fakeSnapshotRepo,
		}

		fakeSnapshotRepo.GetReturns(&models.RouteSnapshot{
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
					Path: "/path1",
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
			},
		}, true)
	})

	It("returns children for a given parent", func() {
		syncRequest := webhook.SyncRequest{
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

		syncResponse, err := syncHandler.Sync(syncRequest)
		Expect(err).ToNot(HaveOccurred())

		Expect(syncResponse).NotTo(BeNil())
		expectedChildren := []webhook.Route{
			webhook.Route{
				ApiVersion: "apps.cloudfoundry.org/v1alpha1",
				Kind:       "Route",
				ObjectMeta: metav1.ObjectMeta{
					Name: "route-guid-0",
					Labels: map[string]string{
						"cloudfoundry.org/bulk-sync-route": "true",
						"label-for-routes":                 "cool-label",
					},
				},
				Spec: webhook.RouteSpec{
					Selector: webhook.Selector{MatchLabels: map[string]string{
						"cloudfoundry.org/route": "route-guid-0",
					}},
					Host: "test0",
					Path: "/path0",
					Domain: webhook.Domain{
						Guid:     "domain-0-guid",
						Name:     "domain0.example.com",
						Internal: false,
					},
					Destinations: []webhook.Destination{
						webhook.Destination{
							Guid:   "route-0-destination-guid-0",
							Port:   9000,
							Weight: models.IntPtr(10),
							App: webhook.App{
								Guid:    "app-guid-0",
								Process: webhook.Process{Type: "process-type-1"},
							},
						},
						webhook.Destination{
							Guid:   "route-0-destination-guid-1",
							Port:   9001,
							Weight: models.IntPtr(11),
							App: webhook.App{
								Guid:    "app-guid-1",
								Process: webhook.Process{Type: "process-type-1"},
							},
						},
					},
				},
			},
			webhook.Route{
				ApiVersion: "apps.cloudfoundry.org/v1alpha1",
				Kind:       "Route",
				ObjectMeta: metav1.ObjectMeta{
					Name: "route-guid-1",
					Labels: map[string]string{
						"cloudfoundry.org/bulk-sync-route": "true",
						"label-for-routes":                 "cool-label",
					},
				},
				Spec: webhook.RouteSpec{
					Selector: webhook.Selector{MatchLabels: map[string]string{
						"cloudfoundry.org/route": "route-guid-1",
					}},

					Host: "test1",
					Path: "/path1",
					Domain: webhook.Domain{
						Guid:     "domain-1-guid",
						Name:     "domain1.apps.internal",
						Internal: true,
					},
					Destinations: []webhook.Destination{
						webhook.Destination{
							Guid:   "route-1-destination-guid-0",
							Port:   8080,
							Weight: models.IntPtr(80),
							App: webhook.App{
								Guid:    "app-guid-2",
								Process: webhook.Process{Type: "process-type-2"},
							},
						},
					},
				},
			},
		}
		Expect(syncResponse.Children).To(Equal(expectedChildren))
	})

	Context("when there's a valid snapshot but it does not contain any routes", func() {
		BeforeEach(func() {
			fakeSnapshotRepo.GetReturns(&models.RouteSnapshot{}, true)
		})
		It("returns an empty list of children in the response", func() {
			syncRequest := webhook.SyncRequest{
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

			syncResponse, err := syncHandler.Sync(syncRequest)
			Expect(err).ToNot(HaveOccurred())
			Expect(syncResponse).NotTo(BeNil())
			Expect(syncResponse.Children).To(Equal([]webhook.Route{}))
		})
	})

	Context("when the repo says no snapshot is available", func() {
		BeforeEach(func() {
			fakeSnapshotRepo.GetReturns(nil, false)
		})
		It("returns a meaningful error", func() {
			syncRequest := webhook.SyncRequest{
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
			_, err := syncHandler.Sync(syncRequest)
			Expect(err).To(Equal(webhook.UninitializedError))
		})
	})

})
