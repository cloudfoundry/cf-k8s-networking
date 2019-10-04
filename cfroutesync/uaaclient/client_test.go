package uaaclient_test

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/uaaclient"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/uaaclient/fakes"
)

var _ = Describe("Client", func() {
	var (
		client     *uaaclient.Client
		jsonClient *fakes.JSONClient
	)

	Describe("GetToken", func() {
		BeforeEach(func() {
			jsonClient = &fakes.JSONClient{}
			client = &uaaclient.Client{
				BaseURL:    "https://some.base.url",
				Name:       "some-name",
				Secret:     "some-secret",
				JSONClient: jsonClient,
			}

			body := `{ "access_token" : "valid-token" }`

			jsonClient.MakeRequestStub = func(req *http.Request, responseStruct interface{}) error {
				return json.Unmarshal([]byte(body), responseStruct)
			}
		})

		It("Returns the token", func() {
			token, err := client.GetToken()
			Expect(err).NotTo(HaveOccurred())
			Expect(token).To(Equal("valid-token"))
		})

		It("forms the required request", func() {
			_, err := client.GetToken()
			Expect(err).NotTo(HaveOccurred())
			Expect(jsonClient.MakeRequestCallCount()).To(Equal(1))
			receivedRequest, _ := jsonClient.MakeRequestArgsForCall(0)
			Expect(receivedRequest.Method).To(Equal("POST"))
			Expect(receivedRequest.URL.RawQuery).To(BeEmpty())
			receivedBytes, _ := ioutil.ReadAll(receivedRequest.Body)
			Expect(receivedBytes).To(Equal([]byte("client_id=some-name&grant_type=client_credentials")))

			authHeader := receivedRequest.Header["Authorization"]
			Expect(authHeader).To(HaveLen(1))
			Expect(authHeader[0]).To(Equal("Basic c29tZS1uYW1lOnNvbWUtc2VjcmV0"))

			contentType := receivedRequest.Header.Get("Content-Type")
			Expect(contentType).To(Equal("application/x-www-form-urlencoded"))
		})

		Context("when the json client returns an error", func() {
			BeforeEach(func() {
				jsonClient.MakeRequestReturns(errors.New("potato"))
			})

			It("returns a helpful error", func() {
				_, err := client.GetToken()
				Expect(err).To(MatchError(ContainSubstring("potato")))
			})
		})

		Context("when the url is malformed", func() {
			BeforeEach(func() {
				client.BaseURL = "%%%%%%%"
			})

			It("returns a helpful error", func() {
				_, err := client.GetToken()
				Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
			})
		})
	})
})
