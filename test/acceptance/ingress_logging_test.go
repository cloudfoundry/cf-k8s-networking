package acceptance_test

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("IngressLogging", func() {
	When("a user sends a GET request to an app", func() {
		It("produces an RTR log visible in cf logs", func() {
			logSession := cf.Cf("logs", globals.AppName)

			customTransport := http.DefaultTransport.(*http.Transport).Clone()
			customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
			client := &http.Client{Transport: customTransport}

			_, err := client.Get(fmt.Sprintf("https://%s.%s", globals.AppName, globals.AppsDomain))
			Expect(err).NotTo(HaveOccurred())

			Eventually(logSession, 10*time.Second).Should(gbytes.Say("RTR.*" + globals.AppGuid))

			logSession.Kill()
			Eventually(logSession).Should(gexec.Exit())
		})
	})
})
