package cfg_test

import (
	"os"
	"time"

	"code.cloudfoundry.org/cf-k8s-networking/routecontroller/cfg"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	Describe("Load", func() {
		BeforeEach(func() {
			err := os.Setenv("INGRESS_PROVIDER", "istio")
			Expect(err).NotTo(HaveOccurred())
			err = os.Setenv("ISTIO_GATEWAY_NAME", "some-gateway")
			Expect(err).NotTo(HaveOccurred())
		})
		AfterEach(func() {
			os.Clearenv()
		})

		Context("when the RESYNC_INTERVAL env var is set", func() {
			BeforeEach(func() {
				err := os.Setenv("RESYNC_INTERVAL", "15")
				Expect(err).NotTo(HaveOccurred())
			})
			It("is configured correctly ", func() {
				config, err := cfg.Load()
				Expect(err).NotTo(HaveOccurred())

				Expect(config.ResyncInterval).To(Equal(15 * time.Second))
			})
		})

		Context("when the RESYNC_INTERVAL env var is not set", func() {
			It("defaults to 30 seconds", func() {
				config, err := cfg.Load()
				Expect(err).NotTo(HaveOccurred())
				Expect(config.ResyncInterval).To(Equal(30 * time.Second))
			})
		})

		Context("when the INGRESS_PROVIDER env var is not set", func() {
			BeforeEach(func() {
				err := os.Unsetenv("INGRESS_PROVIDER")
				Expect(err).NotTo(HaveOccurred())
			})
			It("returns an error", func() {
				_, err := cfg.Load()
				Expect(err).To(MatchError("INGRESS_PROVIDER not configured"))
			})
		})

		Context("when the INGRESS_PROVIDER env var is set to an invalid provider", func() {
			BeforeEach(func() {
				err := os.Setenv("INGRESS_PROVIDER", "some-ingress-provider")
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error", func() {
				_, err := cfg.Load()
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("INGRESS_PROVIDER=some-ingress-provider not supported"))
			})
		})

		Context("When INGRESS_PROVIDER env var is set to Istio", func() {
			BeforeEach(func() {
				err := os.Setenv("ISTIO_GATEWAY_NAME", "some-gateway")
				Expect(err).NotTo(HaveOccurred())
				err = os.Setenv("INGRESS_PROVIDER", "istio")
				Expect(err).NotTo(HaveOccurred())
			})

			It("is configured correctly ", func() {
				config, err := cfg.Load()
				Expect(err).NotTo(HaveOccurred())

				Expect(config.Istio.Gateway).To(Equal("some-gateway"))
				Expect(config.IngressProvider).To(Equal(cfg.IngressProvider("istio")))
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

		Context("When INGRESS_PROVIDER env var is set to Contour", func() {
			BeforeEach(func() {
				err := os.Setenv("INGRESS_PROVIDER", "contour")
				Expect(err).NotTo(HaveOccurred())
			})

			It("is configured correctly", func() {
				config, err := cfg.Load()
				Expect(err).NotTo(HaveOccurred())

				Expect(config.IngressProvider).To(Equal(cfg.IngressProvider("contour")))
			})

		})
	})
})
