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
		fakeSnapshotRepo          *fakes.SnapshotRepo
		fakeServiceBuilder        *fakes.K8sResourceBuilder
		fakeVirtualServiceBuilder *fakes.K8sResourceBuilder
		lineage                   *webhook.Lineage
		syncRequest               webhook.SyncRequest
		fullSnapshot              models.RouteSnapshot
	)

	BeforeEach(func() {
		fakeSnapshotRepo = &fakes.SnapshotRepo{}
		fakeServiceBuilder = &fakes.K8sResourceBuilder{}
		fakeVirtualServiceBuilder = &fakes.K8sResourceBuilder{}

		lineage = &webhook.Lineage{
			RouteSnapshotRepo: fakeSnapshotRepo,
			K8sResourceBuilders: []webhook.K8sResourceBuilder{
				fakeServiceBuilder,
				fakeVirtualServiceBuilder,
			},
		}

		syncRequest = webhook.SyncRequest{
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

		fullSnapshot = models.RouteSnapshot{
			Routes: []models.Route{
				models.Route{
					Guid: "route-guid-0",
					Host: "test0",
					Path: "/path0",
				},
			},
		}

		services := []webhook.K8sResource{
			webhook.Service{
				Kind: "Service1",
			},
			webhook.Service{
				Kind: "Service2",
			},
		}

		virtualServices := []webhook.K8sResource{
			webhook.VirtualService{
				Kind: "VirtualService1",
			},
			webhook.VirtualService{
				Kind: "VirtualService2",
			},
		}

		fakeSnapshotRepo.GetReturns(&fullSnapshot, true)
		fakeServiceBuilder.BuildReturns(services)
		fakeVirtualServiceBuilder.BuildReturns(virtualServices)
	})

	It("returns services and virtual services as a metacontroller response️", func() {
		syncResponse, err := lineage.Sync(syncRequest)
		Expect(err).ToNot(HaveOccurred())
		Expect(syncResponse).NotTo(BeNil())

		sbRoutes, sbTemplate := fakeServiceBuilder.BuildArgsForCall(0)
		Expect(sbRoutes).To(Equal(fullSnapshot.Routes))
		Expect(sbTemplate).To(Equal(syncRequest.Parent.Spec.Template))

		vsbRoutes, vsbTemplate := fakeVirtualServiceBuilder.BuildArgsForCall(0)
		Expect(vsbRoutes).To(Equal(fullSnapshot.Routes))
		Expect(vsbTemplate).To(Equal(syncRequest.Parent.Spec.Template))

		expectedChildren := []webhook.K8sResource{
			webhook.Service{
				Kind: "Service1",
			},
			webhook.Service{
				Kind: "Service2",
			},
			webhook.VirtualService{
				Kind: "VirtualService1",
			},
			webhook.VirtualService{
				Kind: "VirtualService2",
			},
		}

		Expect(syncResponse.Children).To(Equal(expectedChildren))
	})

	Context("when there's snapshot but it does not contain any routes", func() {
		BeforeEach(func() {
			fakeSnapshotRepo.GetReturns(&models.RouteSnapshot{}, true)
			fakeServiceBuilder.BuildReturns([]webhook.K8sResource{})
			fakeVirtualServiceBuilder.BuildReturns([]webhook.K8sResource{})
		})

		It("returns an empty list of children in the response", func() {
			syncResponse, err := lineage.Sync(syncRequest)
			Expect(err).ToNot(HaveOccurred())
			Expect(syncResponse).NotTo(BeNil())
			Expect(syncResponse.Children).To(Equal([]webhook.K8sResource{}))
		})
	})

	Context("when the repo says no snapshot is available", func() {
		BeforeEach(func() {
			fakeSnapshotRepo.GetReturns(nil, false)
		})

		It("returns a meaningful error", func() {
			_, err := lineage.Sync(syncRequest)
			Expect(err).To(Equal(webhook.UninitializedError))
		})
	})
})
