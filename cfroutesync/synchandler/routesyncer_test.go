package synchandler_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/models"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/synchandler"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/synchandler/fakes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Sync", func() {
	var (
		fakeRouteSnapshot *fakes.RouteSnapshot
		syncHandler       *synchandler.RouteSyncer
	)

	BeforeEach(func() {
		fakeRouteSnapshot = &fakes.RouteSnapshot{}
		syncHandler = &synchandler.RouteSyncer{
			RouteSnapshotRepo: fakeRouteSnapshot,
		}

		fakeRouteSnapshot.GetReturns(&models.RouteSnapshot{
			Routes: []*models.Route{
				&models.Route{
					Guid: "route-guid-1",
					Host: "test1.example.com",
					Path: "/path1",
					Destinations: []*models.Destination{
						&models.Destination{
							Guid: "destination-guid-1",
							App: models.DestinationApp{
								Guid:    "app-guid-1",
								Process: "process-type-1",
							},
							Port:   9000,
							Weight: 10,
						},
					},
				},
				&models.Route{
					Guid: "route-guid-2",
					Host: "test2.example.com",
					Path: "/path2",
					Destinations: []*models.Destination{
						&models.Destination{
							Guid: "destination-guid-2",
							App: models.DestinationApp{
								Guid:    "app-guid-2",
								Process: "process-type-2",
							},
							Port:   8080,
							Weight: 80,
						},
					},
				},
			},
		})
	})

	It("returns children for a given parent", func() {
		syncRequest := synchandler.SyncRequest{
			Parent: synchandler.BulkSync{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec: synchandler.BulkSyncSpec{
					Template: synchandler.ParentTemplate{
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

		syncResponse := syncHandler.Sync(syncRequest)
		Expect(syncResponse).NotTo(BeNil())
		expectedChildren := []*synchandler.RouteCRD{
			&synchandler.RouteCRD{
				ApiVersion: "apps.cloudfoundry.org/v1alpha1",
				Kind:       "Route",
				ObjectMeta: metav1.ObjectMeta{
					Name: "route-guid-1",
					Labels: map[string]string{
						"cloudfoundry.org/bulk-sync-route": "true",
						"label-for-routes":                 "cool-label",
					},
				},
				Spec: synchandler.RouteCRDSpec{
					Host: "test1.example.com",
					Path: "/path1",
					Destinations: []synchandler.RouteCRDDestination{
						synchandler.RouteCRDDestination{
							Guid:   "destination-guid-1",
							Port:   9000,
							Weight: 10,
							App: synchandler.RouteCRDDestinationApp{
								Guid:    "app-guid-1",
								Process: "process-type-1",
							},
						},
					},
				},
			},
			&synchandler.RouteCRD{
				ApiVersion: "apps.cloudfoundry.org/v1alpha1",
				Kind:       "Route",
				ObjectMeta: metav1.ObjectMeta{
					Name: "route-guid-2",
					Labels: map[string]string{
						"cloudfoundry.org/bulk-sync-route": "true",
						"label-for-routes":                 "cool-label",
					},
				},
				Spec: synchandler.RouteCRDSpec{
					Host: "test2.example.com",
					Path: "/path2",
					Destinations: []synchandler.RouteCRDDestination{
						synchandler.RouteCRDDestination{
							Guid:   "destination-guid-2",
							Port:   8080,
							Weight: 80,
							App: synchandler.RouteCRDDestinationApp{
								Guid:    "app-guid-2",
								Process: "process-type-2",
							},
						},
					},
				},
			},
		}
		Expect(syncResponse.Children).To(Equal(expectedChildren))
	})

	Context("when there are no routes", func() {
		BeforeEach(func() {
			fakeRouteSnapshot.GetReturns(&models.RouteSnapshot{})
		})
		It("returns an empty list of children in the response", func() {
			syncRequest := synchandler.SyncRequest{
				Parent: synchandler.BulkSync{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: synchandler.BulkSyncSpec{
						Template: synchandler.ParentTemplate{
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

			syncResponse := syncHandler.Sync(syncRequest)
			Expect(syncResponse).NotTo(BeNil())
			Expect(syncResponse.Children).To(Equal([]*synchandler.RouteCRD{}))
		})
	})
})
