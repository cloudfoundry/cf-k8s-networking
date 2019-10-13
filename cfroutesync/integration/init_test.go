package integration_test

import (
	"math/rand"
	"testing"

	. "github.com/onsi/ginkgo"
	ginkgoConfig "github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var (
	binaryPathCFRouteSync string
)

var _ = SynchronizedBeforeSuite(func() []byte {
	binaryPathCFRouteSync, err := gexec.Build(
		"code.cloudfoundry.org/cf-k8s-networking/cfroutesync",
		"-race",
	)
	Expect(err).NotTo(HaveOccurred())
	return []byte(binaryPathCFRouteSync)
}, func(data []byte) {
	binaryPathCFRouteSync = string(data)
	rand.Seed(ginkgoConfig.GinkgoConfig.RandomSeed + int64(GinkgoParallelNode()))
})

var _ = SynchronizedAfterSuite(func() {}, func() {
	gexec.CleanupBuildArtifacts()
})
