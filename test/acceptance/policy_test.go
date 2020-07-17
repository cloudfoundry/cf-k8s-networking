package acceptance_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Policy and mesh connectivity", func() {
	var (
		app1name string
		app2name string
		app2guid string
		domain   string
	)

	BeforeEach(func() {
		app1name = generator.PrefixedRandomName("ACCEPTANCE", "proxy1")
		app2name = generator.PrefixedRandomName("ACCEPTANCE", "proxy2")

		_ = pushProxy(app1name)
		app2guid = pushProxy(app2name)

		domain = globals.AppsDomain
	})

	AfterEach(func() {
		cf.Cf("delete", app1name)
		cf.Cf("delete", app2name)
	})

	Context("to metrics / stats endpoints", func() {
		It("succeeds", func() {
			route := fmt.Sprintf("http://%s.%s/proxy/%s", app1name, domain, url.QueryEscape("istiod.istio-system:15014/metrics"))
			fmt.Printf("Attempting to reach %s", route)
			resp, err := http.Get(route)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(body)).To(ContainSubstring("component=\"pilot\""))
		})

	})

	Context("from apps", func() {
		Context("to istio control plane components", func() {
			It("fails", func() {
				route := fmt.Sprintf("http://%s.%s/proxy/%s", app1name, domain, url.QueryEscape("istiod.istio-system:8080/debug/edsz"))
				expectConnectError(route)
			})
		})

		Context("to other apps over the internal network", func() {
			It("fails", func() {
				service, err := getSvcHTTPAddrBySelector("cf-workloads", fmt.Sprintf("cloudfoundry.org/app_guid=%s", app2guid))
				Expect(err).NotTo(HaveOccurred())

				route := fmt.Sprintf("http://%s.%s/proxy/%s", app1name, domain, url.QueryEscape(service))
				expectConnectError(route)
			})
		})

		Context("to other apps via hairpinning", func() {
			It("succeeds", func() {
				route := fmt.Sprintf("http://%s.%s/proxy/%s.%s", app1name, domain, app2name, domain)
				fmt.Printf("Attempting to reach %s", route)
				resp, err := http.Get(route)
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(200))

				buf := new(bytes.Buffer)
				buf.ReadFrom(resp.Body)
				bodyStr := buf.String()
				fmt.Println(bodyStr)

				Expect(bodyStr).To(MatchRegexp("ListenAddresses"))

				defer resp.Body.Close()
			})
		})
	})
})

func expectConnectError(route string) {
	fmt.Printf("Attempting to reach %s", route)
	resp, err := http.Get(route)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(200))

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	bodyStr := buf.String()
	fmt.Println(bodyStr)

	Expect(bodyStr).To(MatchRegexp("connect error"))

	defer resp.Body.Close()
}
