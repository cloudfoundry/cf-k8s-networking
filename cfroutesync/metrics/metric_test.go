package metrics_test

import (
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/metrics"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Metrics", func() {
	It("has a LastUpdatedAt gauge", func() {
		m := metrics.DefaultMetrics
		Expect(m.ObservedValues.LastUpdatedAt.Desc().String()).To(ContainSubstring("cfroutesync_last_updated_at"))
	})

	It("has a NumberOfRoutes gauge", func() {
		m := metrics.DefaultMetrics
		Expect(m.ObservedValues.NumberOfRoutes.Desc().String()).To(ContainSubstring("cfroutesync_fetched_routes"))
	})
})
