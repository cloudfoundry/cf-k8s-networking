package cfg_test

import (
	"os"

	"code.cloudfoundry.org/cf-k8s-networking/routecontroller/cfg"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	Describe("Load", func() {
		BeforeEach(func() {
			err := os.Setenv("ISTIO_GATEWAY_NAME", "some-gateway")
			Expect(err).NotTo(HaveOccurred())
		})

		It("loads the config", func() {
			config, err := cfg.Load()
			Expect(err).NotTo(HaveOccurred())

			Expect(config.Istio.Gateway).To(Equal("some-gateway"))
		})

		Context("when the ISTIO_GATEWAY_NAME env var is not set", func() {
			BeforeEach(func() {
				err := os.Unsetenv("ISTIO_GATEWAY_NAME")
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error", func() {
				_, err := cfg.Load()
				Expect(err).To(MatchError("ISTIO_GATEWAY_NAME not configured"))
			})
		})
	})
})
