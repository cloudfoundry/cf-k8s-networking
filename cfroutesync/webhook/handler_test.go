package webhook_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/webhook"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/webhook/fakes"
	hfakes "code.cloudfoundry.org/cf-networking-helpers/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("ServeHTTP", func() {
	var (
		handler     *webhook.SyncHandler
		resp        *httptest.ResponseRecorder
		marshaler   *hfakes.Marshaler
		unmarshaler *hfakes.Unmarshaler
		fakeSyncer  *fakes.Syncer
	)

	BeforeEach(func() {
		marshaler = &hfakes.Marshaler{}
		marshaler.MarshalStub = json.Marshal

		unmarshaler = &hfakes.Unmarshaler{}
		unmarshaler.UnmarshalStub = json.Unmarshal

		fakeSyncer = &fakes.Syncer{}

		handler = &webhook.SyncHandler{
			Marshaler:   marshaler,
			Unmarshaler: unmarshaler,
			Syncer:      fakeSyncer,
		}

		fakeSyncResponse := &webhook.SyncResponse{
			Children: []webhook.K8sResource{
				webhook.VirtualService{
					ApiVersion: "networking.istio.io/v1alpha3",
					Kind:       "VirtualService",
				},
				webhook.Service{
					ApiVersion: "v1",
					Kind:       "Service",
				},
			},
		}
		fakeSyncer.SyncReturns(fakeSyncResponse, nil)

		resp = httptest.NewRecorder()
	})
	Context("with a valid metacontroller request", func() {
		var (
			metacontrollerRequestBody string
			request                   *http.Request
			err                       error
		)

		BeforeEach(func() {
			metacontrollerRequestBody = `
			{
				"controller": {},
				"parent": {
                    "apiVersion": "apps.cloudfoundry.org/v1alpha1",
                    "kind": "RouteBulkSync",
                    "metadata": {},
                    "spec": {
				        "selector": {
				            "matchLabels": {
				                "cloudfoundry.org/bulk-sync-route": "true"
				            }
				        },
				        "template": {
				            "metadata": {
				                "labels": {
				                    "cloudfoundry.org/bulk-sync-route": "true"
				                }
				            }
				        }
				    },
                    "status": {}
                },
				"children": [],
				"finalizing": false
			}
		`

			requestBody := bytes.NewBuffer([]byte(metacontrollerRequestBody))

			request, err = http.NewRequest("POST", "/sync", requestBody)
			Expect(err).NotTo(HaveOccurred())
		})

		It("lists VirtualService and Service objects as children", func() {
			handler.ServeHTTP(resp, request)

			expectedSyncRequest := webhook.SyncRequest{
				Parent: webhook.BulkSync{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "apps.cloudfoundry.org/v1alpha1",
						Kind:       "RouteBulkSync",
					},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: webhook.BulkSyncSpec{
						Template: webhook.Template{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"cloudfoundry.org/bulk-sync-route": "true",
								},
							},
						},
						Selector: webhook.Selector{
							MatchLabels: map[string]string{
								"cloudfoundry.org/bulk-sync-route": "true",
							},
						},
					},
				},
			}

			Expect(fakeSyncer.SyncArgsForCall(0)).To(Equal(expectedSyncRequest))

			expectedResponseBody := `
{
	"children": [{
			"apiVersion": "networking.istio.io/v1alpha3",
			"kind": "VirtualService",
			"metadata": {
				"creationTimestamp": null
			},
			"spec": {
				"hosts": null,
				"gateways": null,
				"http": null
			}
		},
		{
			"apiVersion": "v1",
			"kind": "Service",
			"metadata": {
				"creationTimestamp": null
			},
			"spec": {
				"selector": null,
				"ports": null
			}
		}
	]
}
`

			Expect(resp.Code).To(Equal(http.StatusOK))
			Expect(resp.Body).To(MatchJSON(expectedResponseBody))
		})

		Context("when there are no routes", func() {
			BeforeEach(func() {
				fakeSyncResponse := &webhook.SyncResponse{
					Children: []webhook.K8sResource{},
				}
				fakeSyncer.SyncReturns(fakeSyncResponse, nil)
			})

			It("returns an empty list of children", func() {
				handler.ServeHTTP(resp, request)

				Expect(resp.Code).To(Equal(http.StatusOK))
				Expect(resp.Body).To(MatchJSON(`{"children":  []}`))
			})
		})

		Context("when json marshalling returns an error", func() {
			BeforeEach(func() {
				marshaler.MarshalStub = func(interface{}) ([]byte, error) {
					return nil, errors.New("yerba-mate-marshalling-err")
				}
			})

			It("returns an InternalServerError", func() {
				handler.ServeHTTP(resp, request)

				Expect(resp.Code).To(Equal(http.StatusInternalServerError))
				Expect(resp.Body).To(MatchJSON(`{"error": "failed to marshal response"}`))
			})
		})

		Context("when json unmarshalling returns an error", func() {
			BeforeEach(func() {
				unmarshaler.UnmarshalStub = func([]byte, interface{}) error {
					return errors.New("unmarshalling-err")
				}
			})

			It("returns a StatusBadRequest", func() {
				handler.ServeHTTP(resp, request)

				Expect(resp.Code).To(Equal(http.StatusBadRequest))
				Expect(resp.Body).To(MatchJSON(`{"error": "failed to unmarshal request"}`))
			})
		})

		Context("when the syncer isn't yet initialized", func() {
			BeforeEach(func() {
				fakeSyncer.SyncReturns(nil, webhook.UninitializedError)
			})

			It("returns the error in the response and sets a 500 code so that metacontroller won't attempt to modify state in the k8s api", func() {
				handler.ServeHTTP(resp, request)

				Expect(resp.Code).To(Equal(http.StatusInternalServerError))
				Expect(resp.Body).To(ContainSubstring("uninitialized"))
			})
		})

		Context("when the syncer returns any other error", func() {
			BeforeEach(func() {
				fakeSyncer.SyncReturns(nil, errors.New("unknown error!!!"))
			})

			It("returns an Internal Server Error", func() {
				handler.ServeHTTP(resp, request)

				Expect(resp.Code).To(Equal(http.StatusInternalServerError))
				Expect(resp.Body).To(ContainSubstring("Internal Server Error"))
			})
		})
	})
})
