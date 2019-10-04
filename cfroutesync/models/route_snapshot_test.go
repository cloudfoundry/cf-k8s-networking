package models_test

import (
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RouteSnapshot", func() {
	Describe("Get", func() {
		It("returns a *RouteSnapshot", func() {
			routeSnapshotRepo := models.RouteSnapshot{}
			Expect(routeSnapshotRepo.Get()).To(Equal(&models.RouteSnapshot{}))
		})
	})
})
