package uptime_test

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/cf-k8s-networking/test/uptime/internal/checker"
	"code.cloudfoundry.org/cf-k8s-networking/test/uptime/internal/collector"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Control Plane Uptime", func() {
	var (
		upgradeChecker   *checker.Upgrade
		requestCollector *collector.Request
		startTime        time.Time
		count            int
	)

	BeforeEach(func() {
		upgradeChecker = &checker.Upgrade{
			PollInterval: 1 * time.Second,
		}
		requestCollector = &collector.Request{
			DataPlaneSLOMaxRequestLatency:                  controlPlaneSLODataPlaneMaxRequestLatency,
			ControlPlaneSLODataPlaneAvailabilityPercentage: controlPlaneSLODataPlaneAvailabilityPercentage,
			Client: httpClient,
		}

		upgradeChecker.Start()
		startTime = time.Now()
		count = 0
	})

	AfterEach(func() {
		upgradeChecker.Stop()

		for i := 0; i < count; i++ {
			cf.Cf("delete-route", "-f", cfAppDomain, "--hostname", fmt.Sprintf("host-%d", i)).Wait(30 * time.Second)
		}
	})

	It("measures the control plane uptime", func() {
		By("checking whether X% of requests are successful within the acceptable response time during an upgrade", func() {
			for {
				if !upgradeChecker.HasFoundUpgrade() && time.Since(startTime) > upgradeDiscoveryTimeout {
					Fail(fmt.Sprintf("failed to find cf upgrade in %s", upgradeDiscoveryTimeout.String()))
				}

				// if the upgrade is finished (learned by checking the "finished at" in
				// kapp app-change ls), stop running the test
				if upgradeChecker.HasFoundUpgrade() && upgradeChecker.IsUpgradeFinished() {
					break
				}

				routeHost := fmt.Sprintf("host-%d", count)
				cf.Cf("map-route", controlPlaneAppName, cfAppDomain, "--hostname", routeHost)

				route := fmt.Sprintf("http://%s.%s", routeHost, cfAppDomain)
				requestCollector.Request(route, controlPlaneSLORoutePropagationTime, controlPlaneSLOSampleCaptureTime)

				count++
				time.Sleep(5 * time.Second)
			}

			requestCollector.Wait()

			results := requestCollector.GetResults()
			results.PrintResults()

			Expect(results.SuccessPercentage()).To(BeNumerically(">=", controlPlaneSLOPercentage))
			Expect(true).To(BeTrue())
		})
	})
})
