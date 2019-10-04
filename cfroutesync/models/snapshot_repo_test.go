package models_test

import (
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SnapshotRepo", func() {

	It("returns the snapshot that we put in", func() {
		repo := models.SnapshotRepo{}
		thing := &models.RouteSnapshot{
			Routes: []*models.Route{
				&models.Route{
					Guid: "foo",
				},
				&models.Route{
					Guid: "bar",
				},
			},
		}
		repo.Put(thing)

		snapshot, ok := repo.Get()

		Expect(ok).To(BeTrue())
		Expect(snapshot).To(Equal(thing))

	})

})
