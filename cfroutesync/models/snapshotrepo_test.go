package models_test

import (
	"sync"

	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SnapshotRepo", func() {
	Specify("Get returns the snapshot that was Put in", func() {
		repo := models.SnapshotRepo{}
		thing := &models.RouteSnapshot{
			Routes: []models.Route{
				{Guid: "foo"},
				{Guid: "bar"},
			},
		}
		repo.Put(thing)

		snapshot, ok := repo.Get()

		Expect(ok).To(BeTrue())
		Expect(snapshot).To(Equal(thing))
	})

	Context("when no snapshot has been Put into the repo", func() {
		Specify("Get returns nil,false", func() {
			repo := models.SnapshotRepo{}

			snapshot, ok := repo.Get()
			Expect(snapshot).To(BeNil())

			Expect(ok).To(BeFalse())
		})
	})

	// this test is only meaningful if run using the -race flag
	Specify("the repo is safe for concurrent access", func() {
		repo := &models.SnapshotRepo{}
		const numCalls = 100

		var complete sync.WaitGroup
		complete.Add(2)

		go func(repo *models.SnapshotRepo) {
			for i := 0; i < numCalls; i++ {
				thing := &models.RouteSnapshot{
					Routes: []models.Route{
						{Guid: "foo"},
						{Guid: "bar"},
					},
				}
				repo.Put(thing)
			}
			complete.Done()
		}(repo)

		go func(repo *models.SnapshotRepo) {
			for i := 0; i < numCalls; i++ {
				repo.Get()
			}
			complete.Done()
		}(repo)

		complete.Wait()
		// if we've made it this far without a race detected, then
		// the snapshotRepo is safe for concurrent use!
	})
})
