package models_test

import (
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Route", func() {
	Describe("FQDN()", func() {
		Context("when the route does not have a host", func() {
			It("returns an fqdn containing only the domain name", func() {
				route := models.Route{
					Host: "",
					Path: "/path",
					Domain: models.Domain{
						Name: "domain.example.com",
					},
				}

				Expect(route.FQDN()).To(Equal("domain.example.com"))
			})
		})

		Context("when the route has a host", func() {
			It("returns an fqdn containing the host and domain name", func() {
				route := models.Route{
					Host: "host",
					Path: "/path",
					Domain: models.Domain{
						Name: "domain.example.com",
					},
				}

				Expect(route.FQDN()).To(Equal("host.domain.example.com"))
			})
		})
	})
})
