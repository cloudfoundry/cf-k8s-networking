package synchandler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/models"

	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/synchandler"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/synchandler/fakes"
	hfakes "code.cloudfoundry.org/cf-networking-helpers/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("ServeHTTP", func() {
	var (
		handler     *synchandler.SyncHandler
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

		handler = &synchandler.SyncHandler{
			Marshaler:   marshaler,
			Unmarshaler: unmarshaler,
			Syncer:      fakeSyncer,
		}

		fakeSyncResponse := &synchandler.SyncResponse{
			Children: []*synchandler.RouteCRD{
				&synchandler.RouteCRD{
					ApiVersion: "apps.cloudfoundry.org/v1alpha1",
					Kind:       "Route",
					ObjectMeta: metav1.ObjectMeta{
						Name: "route-guid-1",
						Labels: map[string]string{
							"cloudfoundry.org/bulk-sync-route": "true",
						},
					},
					Spec: synchandler.RouteCRDSpec{
						Host: "test1.example.com",
						Path: "/path1",
						Destinations: []synchandler.RouteCRDDestination{
							synchandler.RouteCRDDestination{
								Guid:   "destination-guid-1",
								Port:   9000,
								Weight: models.IntPtr(10),
								App: synchandler.RouteCRDDestinationApp{
									Guid:    "app-guid-1",
									Process: "process-type-1",
								},
							},
						},
					},
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

			request, err = http.NewRequest("POST", "/route_crds", requestBody)
			Expect(err).NotTo(HaveOccurred())
		})

		It("lists route CRDs", func() {
			handler.ServeHTTP(resp, request)

			expectedSyncRequest := synchandler.SyncRequest{
				Parent: synchandler.BulkSync{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "apps.cloudfoundry.org/v1alpha1",
						Kind:       "RouteBulkSync",
					},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: synchandler.BulkSyncSpec{
						Template: synchandler.ParentTemplate{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"cloudfoundry.org/bulk-sync-route": "true",
								},
							},
						},
						Selector: synchandler.ParentSelector{
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
				"children":[
					{
					    "apiVersion": "apps.cloudfoundry.org/v1alpha1",
					    "kind": "Route",
					    "metadata": {
					        "labels": {
					            "cloudfoundry.org/bulk-sync-route": "true"
					        },
					        "name": "route-guid-1",
							"creationTimestamp": null
					    },
					    "spec": {
					        "host": "test1.example.com",
					        "path": "/path1",
							"destinations": [
								{
									"app": {
										"guid": "app-guid-1",
										"process": "process-type-1"
									},
									"guid": "destination-guid-1",
									"port": 9000,
									"weight": 10
								}
							]
					    }
					}
				]
			}`

			Expect(resp.Code).To(Equal(http.StatusOK))
			Expect(resp.Body).To(MatchJSON(expectedResponseBody))
		})

		Context("when there are no routes", func() {
			BeforeEach(func() {
				fakeSyncResponse := &synchandler.SyncResponse{
					Children: []*synchandler.RouteCRD{},
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
				fakeSyncer.SyncReturns(nil, synchandler.UninitializedError)
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
