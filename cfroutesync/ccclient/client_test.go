package ccclient_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/ccclient"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/ccclient/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cloud Controller Client", func() {
	var (
		ccClient   *ccclient.Client
		jsonClient *fakes.JSONClient
		token      string
	)

	BeforeEach(func() {
		jsonClient = &fakes.JSONClient{}
		ccClient = &ccclient.Client{
			JSONClient: jsonClient,
			BaseURL:    "https://some.base.url",
		}
		token = "fake-token"
	})

	Describe("ListRoutes", func() {
		BeforeEach(func() {
			body := `
			{
				"pagination": {
					"total_results": 3,
					"total_pages": 1,
					"first": {
						"href": "https://api.example.org/v3/routes?page=1&per_page=2"
					},
					"last": {
						"href": "https://api.example.org/v3/routes?page=2&per_page=2"
					},
					"next": {
						"href": "https://api.example.org/v3/routes?page=2&per_page=2"
					},
					"previous": null
				},
				"resources": [{
						"guid": "fake-guid",
						"host": "fake-host",
						"path": "/fake_path",
						"url": "fake-host.fake-domain.com/fake_path",
						"metadata": {
							"labels": {},
							"annotations": {}
						},
						"relationships": {
							"domain": {
								"data": {
									"guid": "fake-domain-1-guid"
								}
							},
							"space": {
								"data": {
									"guid": "fake-space-1-guid"
								}
							}
						}
					},
					{
						"guid": "fake-guid2",
						"host": "fake-host2",
						"path": "/fake_path2",
						"url": "fake-host2.fake-domain.com/fake_path2",
						"relationships": {
							"domain": {
								"data": {
									"guid": "fake-domain-2-guid"
								}
							},
							"space": {
								"data": {
									"guid": "fake-space-2-guid"
								}
							}
						}
					}
				]
			}
			`
			jsonClient.MakeRequestStub = func(req *http.Request, responseStruct interface{}) error {
				return json.Unmarshal([]byte(body), responseStruct)
			}
		})

		It("returns a list of routes", func() {
			routeResults, err := ccClient.ListRoutes(token)
			Expect(err).To(Not(HaveOccurred()))
			route1 := ccclient.Route{
				Guid: "fake-guid",
				Host: "fake-host",
				Path: "/fake_path",
				Url:  "fake-host.fake-domain.com/fake_path",
			}
			route1.Relationships.Domain.Data.Guid = "fake-domain-1-guid"
			route1.Relationships.Space.Data.Guid = "fake-space-1-guid"

			route2 := ccclient.Route{
				Guid: "fake-guid2",
				Host: "fake-host2",
				Path: "/fake_path2",
				Url:  "fake-host2.fake-domain.com/fake_path2",
			}
			route2.Relationships.Domain.Data.Guid = "fake-domain-2-guid"
			route2.Relationships.Space.Data.Guid = "fake-space-2-guid"

			Expect(len(routeResults)).To(Equal(2))
			Expect(routeResults).To(ContainElement(route1))
			Expect(routeResults).To(ContainElement(route2))
		})

		It("forms the right request URL", func() {
			_, err := ccClient.ListRoutes(token)
			Expect(err).To(Not(HaveOccurred()))

			receivedRequest, _ := jsonClient.MakeRequestArgsForCall(0)
			Expect(receivedRequest.Method).To(Equal("GET"))
			Expect(receivedRequest.URL.Path).To(Equal("/v3/routes"))
		})

		It("sets the provided token as an Authorization header on the request", func() {
			_, err := ccClient.ListRoutes(token)
			Expect(err).To(Not(HaveOccurred()))

			receivedRequest, _ := jsonClient.MakeRequestArgsForCall(0)

			authHeader := receivedRequest.Header["Authorization"]
			Expect(authHeader).To(HaveLen(1))
			Expect(authHeader[0]).To(Equal("bearer fake-token"))
		})

		Context("this only supports 5000 routes", func() {
			It("requests 5000 results per page", func() {
				_, err := ccClient.ListRoutes(token)
				Expect(err).To(Not(HaveOccurred()))
				receivedRequest, _ := jsonClient.MakeRequestArgsForCall(0)
				Expect(receivedRequest.URL.Query()["per_page"]).To(Equal([]string{"5000"}))
			})
			It("errors if there is more than one page of results", func() {
				body := `{ "pagination": { "total_pages": 2 } }`
				jsonClient.MakeRequestStub = func(req *http.Request, responseStruct interface{}) error {
					return json.Unmarshal([]byte(body), responseStruct)
				}

				_, err := ccClient.ListRoutes(token)
				Expect(err).To(MatchError(ContainSubstring("too many results, paging not implemented")))
			})
		})

		Context("when the json client returns an error", func() {
			BeforeEach(func() {
				jsonClient.MakeRequestReturns(errors.New("potato"))
			})

			It("returns a helpful error", func() {
				_, err := ccClient.ListRoutes(token)
				Expect(err).To(MatchError(ContainSubstring("potato")))
			})
		})

		Context("when the url is malformed", func() {
			BeforeEach(func() {
				ccClient.BaseURL = "%%%%%%%"
			})

			It("returns a helpful error", func() {
				_, err := ccClient.ListRoutes(token)
				Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
			})
		})
	})

	Describe("ListDestinationsForRoute", func() {
		var (
			routeGUID string
		)
		BeforeEach(func() {
			routeGUID = "fake-route-guid"
			body := `
			{
				"destinations": [
				  {
					"guid": "fake-destination-guid-1",
					"app": {
					  "guid": "fake-app-guid-1",
					  "process": {
						"type": "web"
					  }
					},
					"weight": null,
					"port": 8080
				  },
				  {
					"guid": "fake-destination-guid-2",
					"app": {
					  "guid": "fake-app-guid-2",
					  "process": {
						"type": "worker"
					  }
					},
					"weight": 5,
					"port": 9000
				  }
				],
				"links": {
				  "self": {
					"href": "https://api.example.org/v3/routes/fake-route-guid/destinations"
				  },
				  "route": {
					"href": "https://api.example.org/v3/routes/fake-route-guid"
				  }
				}
			}
			`
			jsonClient.MakeRequestStub = func(req *http.Request, responseStruct interface{}) error {
				return json.Unmarshal([]byte(body), responseStruct)
			}
		})

		It("returns a list of destinations for the given route", func() {
			routeWeight := 5
			routeDestinationResults, err := ccClient.ListDestinationsForRoute(routeGUID, token)
			Expect(err).To(Not(HaveOccurred()))

			routeDestination1 := ccclient.Destination{
				Guid:   "fake-destination-guid-1",
				Weight: nil,
				Port:   8080,
			}
			routeDestination1.App.Guid = "fake-app-guid-1"
			routeDestination1.App.Process.Type = "web"

			routeDestination2 := ccclient.Destination{
				Guid:   "fake-destination-guid-2",
				Weight: &routeWeight,
				Port:   9000,
			}
			routeDestination2.App.Guid = "fake-app-guid-2"
			routeDestination2.App.Process.Type = "worker"

			Expect(len(routeDestinationResults)).To(Equal(2))
			Expect(routeDestinationResults).To(ContainElement(routeDestination1))
			Expect(routeDestinationResults).To(ContainElement(routeDestination2))
		})

		It("forms the right request URL", func() {
			_, err := ccClient.ListDestinationsForRoute(routeGUID, token)
			Expect(err).To(Not(HaveOccurred()))

			receivedRequest, _ := jsonClient.MakeRequestArgsForCall(0)
			Expect(receivedRequest.Method).To(Equal("GET"))
			Expect(receivedRequest.URL.Path).To(Equal(fmt.Sprintf("/v3/routes/%s/destinations", routeGUID)))
		})

		It("sets the provided token as an Authorization header on the request", func() {
			_, err := ccClient.ListDestinationsForRoute(routeGUID, token)
			Expect(err).To(Not(HaveOccurred()))

			receivedRequest, _ := jsonClient.MakeRequestArgsForCall(0)

			authHeader := receivedRequest.Header["Authorization"]
			Expect(authHeader).To(HaveLen(1))
			Expect(authHeader[0]).To(Equal("bearer fake-token"))
		})

		Context("when the url is malformed", func() {
			BeforeEach(func() {
				ccClient.BaseURL = "%%%%%%%"
			})

			It("returns a helpful error", func() {
				_, err := ccClient.ListDestinationsForRoute(routeGUID, token)
				Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
			})
		})

		Context("when the json client returns an error", func() {
			BeforeEach(func() {
				jsonClient.MakeRequestReturns(errors.New("potato"))
			})

			It("returns a helpful error", func() {
				_, err := ccClient.ListDestinationsForRoute(routeGUID, token)
				Expect(err).To(MatchError(ContainSubstring("potato")))
			})
		})

	})

	Describe("ListDomains", func() {
		BeforeEach(func() {
			body := `
			{
			  "pagination": {
				"total_results": 3,
				"total_pages": 1,
				"first": {
				  "href": "https://api.example.org/v3/domains?page=1&per_page=2"
				},
				"last": {
				  "href": "https://api.example.org/v3/domains?page=2&per_page=2"
				},
				"next": {
				  "href": "https://api.example.org/v3/domains?page=2&per_page=2"
				},
				"previous": null
			  },
			  "resources": [
				{
				  "guid": "fake-domain-1-guid",
				  "name": "fake-domain1.example.com",
                  "internal": false,
				  "metadata": {
					"labels": {},
					"annotations": {}
				  }
				},
				{
				  "guid": "fake-domain-2-guid",
				  "name": "fake-domain2.example.com",
                  "internal": true,
				  "metadata": {
					"labels": {},
					"annotations": {}
				  }
				}
			  ]
			}
			`
			jsonClient.MakeRequestStub = func(req *http.Request, responseStruct interface{}) error {
				return json.Unmarshal([]byte(body), responseStruct)
			}
		})

		It("returns a list of domains", func() {
			domainResults, err := ccClient.ListDomains(token)
			Expect(err).To(Not(HaveOccurred()))
			domain1 := ccclient.Domain{
				Guid:     "fake-domain-1-guid",
				Name:     "fake-domain1.example.com",
				Internal: false,
			}
			domain2 := ccclient.Domain{
				Guid:     "fake-domain-2-guid",
				Name:     "fake-domain2.example.com",
				Internal: true,
			}
			Expect(len(domainResults)).To(Equal(2))
			Expect(domainResults).To(ContainElement(domain1))
			Expect(domainResults).To(ContainElement(domain2))
		})

		It("forms the right request URL", func() {
			_, err := ccClient.ListDomains(token)
			Expect(err).To(Not(HaveOccurred()))

			receivedRequest, _ := jsonClient.MakeRequestArgsForCall(0)
			Expect(receivedRequest.Method).To(Equal("GET"))
			Expect(receivedRequest.URL.Path).To(Equal("/v3/domains"))
		})

		It("sets the provided token as an Authorization header on the request", func() {
			_, err := ccClient.ListDomains(token)
			Expect(err).To(Not(HaveOccurred()))

			receivedRequest, _ := jsonClient.MakeRequestArgsForCall(0)

			authHeader := receivedRequest.Header["Authorization"]
			Expect(authHeader).To(HaveLen(1))
			Expect(authHeader[0]).To(Equal("bearer fake-token"))
		})

		Context("this only supports 5000 domains", func() {
			It("requests 5000 results per page", func() {
				_, err := ccClient.ListDomains(token)
				Expect(err).To(Not(HaveOccurred()))
				receivedRequest, _ := jsonClient.MakeRequestArgsForCall(0)
				Expect(receivedRequest.URL.Query()["per_page"]).To(Equal([]string{"5000"}))
			})
			It("errors if there is more than one page of results", func() {
				body := `{ "pagination": { "total_pages": 2 } }`
				jsonClient.MakeRequestStub = func(req *http.Request, responseStruct interface{}) error {
					return json.Unmarshal([]byte(body), responseStruct)
				}

				_, err := ccClient.ListDomains(token)
				Expect(err).To(MatchError(ContainSubstring("too many results, paging not implemented")))
			})
		})

		Context("when the json client returns an error", func() {
			BeforeEach(func() {
				jsonClient.MakeRequestReturns(errors.New("potato"))
			})

			It("returns a helpful error", func() {
				_, err := ccClient.ListDomains(token)
				Expect(err).To(MatchError(ContainSubstring("potato")))
			})
		})

		Context("when the url is malformed", func() {
			BeforeEach(func() {
				ccClient.BaseURL = "%%%%%%%"
			})

			It("returns a helpful error", func() {
				_, err := ccClient.ListDomains(token)
				Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
			})
		})
	})

	Describe("ListSpaces", func() {
		BeforeEach(func() {
			body := `
			{
			  "pagination": {
				"total_results": 3,
				"total_pages": 1,
				"first": {
				  "href": "https://api.example.org/v3/spaces?page=1&per_page=2"
				},
				"last": {
				  "href": "https://api.example.org/v3/spaces?page=2&per_page=2"
				},
				"next": {
				  "href": "https://api.example.org/v3/spaces?page=2&per_page=2"
				},
				"previous": null
			  },
			  "resources": [
				{
				  "guid": "fake-space-1-guid",
				  "name": "fake-space1",
                  "relationships": {
					"organization": {
					  "data": {
						"guid": "fake-org-guid-1"
					  }
					}
				  },
				  "metadata": {
					"labels": {},
					"annotations": {}
				  }
				},
				{
				  "guid": "fake-space-2-guid",
				  "name": "fake-space2",
                  "relationships": {
					"organization": {
					  "data": {
						"guid": "fake-org-guid-2"
					  }
					}
				  },
				  "metadata": {
					"labels": {},
					"annotations": {}
				  }
				}
			  ]
			}
			`
			jsonClient.MakeRequestStub = func(req *http.Request, responseStruct interface{}) error {
				return json.Unmarshal([]byte(body), responseStruct)
			}
		})

		It("returns a list of spaces", func() {
			spaceResults, err := ccClient.ListSpaces(token)
			Expect(err).To(Not(HaveOccurred()))
			space1 := ccclient.Space{
				Guid: "fake-space-1-guid",
			}
			space1.Relationships.Organization.Data.Guid = "fake-org-guid-1"

			space2 := ccclient.Space{
				Guid: "fake-space-2-guid",
			}
			space2.Relationships.Organization.Data.Guid = "fake-org-guid-2"

			Expect(len(spaceResults)).To(Equal(2))
			Expect(spaceResults).To(ContainElement(space1))
			Expect(spaceResults).To(ContainElement(space2))
		})

		It("forms the right request URL", func() {
			_, err := ccClient.ListSpaces(token)
			Expect(err).To(Not(HaveOccurred()))

			receivedRequest, _ := jsonClient.MakeRequestArgsForCall(0)
			Expect(receivedRequest.Method).To(Equal("GET"))
			Expect(receivedRequest.URL.Path).To(Equal("/v3/spaces"))
		})

		It("sets the provided token as an Authorization header on the request", func() {
			_, err := ccClient.ListSpaces(token)
			Expect(err).To(Not(HaveOccurred()))

			receivedRequest, _ := jsonClient.MakeRequestArgsForCall(0)

			authHeader := receivedRequest.Header["Authorization"]
			Expect(authHeader).To(HaveLen(1))
			Expect(authHeader[0]).To(Equal("bearer fake-token"))
		})

		Context("this only supports 5000 spaces", func() {
			It("requests 5000 results per page", func() {
				_, err := ccClient.ListSpaces(token)
				Expect(err).To(Not(HaveOccurred()))
				receivedRequest, _ := jsonClient.MakeRequestArgsForCall(0)
				Expect(receivedRequest.URL.Query()["per_page"]).To(Equal([]string{"5000"}))
			})
			It("errors if there is more than one page of results", func() {
				body := `{ "pagination": { "total_pages": 2 } }`
				jsonClient.MakeRequestStub = func(req *http.Request, responseStruct interface{}) error {
					return json.Unmarshal([]byte(body), responseStruct)
				}

				_, err := ccClient.ListSpaces(token)
				Expect(err).To(MatchError(ContainSubstring("too many results, paging not implemented")))
			})
		})

		Context("when the json client returns an error", func() {
			BeforeEach(func() {
				jsonClient.MakeRequestReturns(errors.New("potato"))
			})

			It("returns a helpful error", func() {
				_, err := ccClient.ListSpaces(token)
				Expect(err).To(MatchError(ContainSubstring("potato")))
			})
		})

		Context("when the url is malformed", func() {
			BeforeEach(func() {
				ccClient.BaseURL = "%%%%%%%"
			})

			It("returns a helpful error", func() {
				_, err := ccClient.ListSpaces(token)
				Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
			})
		})
	})
})
