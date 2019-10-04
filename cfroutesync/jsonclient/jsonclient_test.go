package jsonclient_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/jsonclient"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/jsonclient/fakes"
)

var _ = Describe("JSON Client", func() {
	var (
		client     *jsonclient.JSONClient
		httpClient *fakes.HTTPClient
	)

	BeforeEach(func() {
		httpClient = &fakes.HTTPClient{}
		client = &jsonclient.JSONClient{
			HTTPClient: httpClient,
		}
	})

	Context("when the http client returns an error", func() {
		BeforeEach(func() {
			httpClient.DoReturns(nil, errors.New("potato"))
		})

		It("returns a helpful error", func() {
			err := client.MakeRequest(&http.Request{}, struct{}{})
			Expect(err).To(MatchError(ContainSubstring("http client: potato")))
		})
	})

	Context("if the response status code is not 200", func() {
		BeforeEach(func() {
			httpClient.DoReturns(&http.Response{
				StatusCode: 418,
				Body:       ioutil.NopCloser(strings.NewReader("bad thing")),
			}, nil)
		})

		It("returns the response body in the error", func() {
			err := client.MakeRequest(&http.Request{}, struct{}{})
			Expect(err).To(MatchError("bad response, code 418: bad thing"))
		})
	})

	Context("when the response body is not valid json", func() {
		BeforeEach(func() {
			httpClient.DoReturns(&http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader(`%%%%`)),
			}, nil)
		})

		It("returns a helpful error", func() {
			err := client.MakeRequest(&http.Request{}, struct{}{})
			Expect(err).To(MatchError(ContainSubstring("unmarshal json: invalid character")))
		})
	})
})
