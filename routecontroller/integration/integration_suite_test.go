package integration_test

import (
	"testing"
	"time"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var (
	routeControllerBinaryPath string
)

const (
	defaultTimeout         = 30 * time.Second
	defaultPollingInterval = 1 * time.Second
)

var _ = SynchronizedBeforeSuite(func() []byte {
	binPath, err := gexec.Build(
		"code.cloudfoundry.org/cf-k8s-networking/routecontroller",
		"--race",
	)
	Expect(err).NotTo(HaveOccurred())

	SetDefaultEventuallyTimeout(defaultTimeout)
	SetDefaultEventuallyPollingInterval(defaultPollingInterval)
	SetDefaultConsistentlyDuration(defaultTimeout)
	SetDefaultConsistentlyPollingInterval(defaultPollingInterval)

	return []byte(binPath)
}, func(data []byte) {
	routeControllerBinaryPath = string(data)
})

var _ = SynchronizedAfterSuite(func() {}, func() {
	gexec.CleanupBuildArtifacts()
})
