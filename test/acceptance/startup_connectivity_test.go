package acceptance_test

import (
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Outbound network connectivity during app startup", func() {
	var (
		app1name string
	)

	BeforeEach(func() {
		app1name = generator.PrefixedRandomName("ACCEPTANCE", "outbound-network-app")
	})

	AfterEach(func() {
		session := cf.Cf("delete", app1name, "-f")
		Expect(session.Wait(TestConfig.DefaultTimeoutDuration())).To(gexec.Exit(0), "expected cf delete to succeed")
	})

	Context("pushing the app", func() {
		It("succeeds", func() {
			session := cf.Cf("push",
				app1name,
				"-p", "assets/outbound-network-request-app",
			)
			Expect(session.Wait(TestConfig.CfPushTimeoutDuration())).To(gexec.Exit(0), "expected app to start successfully")
		})
	})
})
