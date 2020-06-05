package uptime_test

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	FLOAT_BIT_SIZE = 64
)

var (
	upgradeDiscoveryTimeout                        time.Duration
	dataPlaneSLIAppRouteURL                        string
	dataPlaneSLOMaxRequestLatency                  time.Duration
	dataPlaneSLOPercentage                         float64
	dataPlaneAppName                               string
	controlPlaneSLORoutePropagationTime            time.Duration
	controlPlaneSLOSampleCaptureTime               time.Duration
	controlPlaneSLODataPlaneAvailabilityPercentage float64
	cfAppDomain                                    string
	controlPlaneAppName                            string
	httpClient                                     http.Client
)

func TestUptimeTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UptimeTests Suite")
}

var _ = BeforeSuite(func() {
	var err error

	upgradeDiscoveryTimeout, err = time.ParseDuration(os.Getenv("UPGRADE_DISCOVERY_TIMEOUT"))
	Expect(err).NotTo(HaveOccurred(), "UPGRADE_DISCOVERY_TIMEOUT malformed")

	dataPlaneSLOMaxRequestLatency, err = time.ParseDuration(os.Getenv("DATA_PLANE_SLO_MAX_REQUEST_LATENCY"))
	Expect(err).NotTo(HaveOccurred(), "DATA_PLANE_SLO_MAX_REQUEST_LATENCY malformed")

	dataPlaneSLOPercentage, err = strconv.ParseFloat(os.Getenv("DATA_PLANE_SLO_PERCENTAGE"), FLOAT_BIT_SIZE)
	Expect(err).NotTo(HaveOccurred(), "DATA_PLANE_SLO_PERCENTAGE malformed")

	controlPlaneSLORoutePropagationTime, err = time.ParseDuration(os.Getenv("CONTROL_PLANE_SLO_MAX_ROUTE_PROPAGATION_TIME"))
	Expect(err).NotTo(HaveOccurred(), "CONTROL_PLANE_SLO_MAX_ROUTE_PROPAGATION_TIME malformed")

	controlPlaneSLOSampleCaptureTime, err = time.ParseDuration(os.Getenv("CONTROL_PLANE_SLO_SAMPLE_CAPTURE_TIME"))
	Expect(err).NotTo(HaveOccurred(), "CONTROL_PLANE_SLO_SAMPLE_CAPTURE_TIME malformed")

	controlPlaneSLODataPlaneAvailabilityPercentage, err = strconv.ParseFloat(os.Getenv("CONTROL_PLANE_SLO_DATA_PLANE_AVAILABILITY_PERCENTAGE"), FLOAT_BIT_SIZE)
	Expect(err).NotTo(HaveOccurred(), "CONTROL_PLANE_SLO_DATA_PLANE_AVAILABILITY_PERCENTAGE malformed")

	var found bool
	cfAppDomain, found = os.LookupEnv("CF_APP_DOMAIN")
	Expect(found).To(BeTrue(), "CF_APP_DOMAIN required but not set")

	controlPlaneAppName, found = os.LookupEnv("CONTROL_PLANE_APP_NAME")
	Expect(found).To(BeTrue(), "CONTROL_PLANE_APP_NAME required but not set")

	dataPlaneAppName, found = os.LookupEnv("DATA_PLANE_APP_NAME")
	Expect(found).To(BeTrue(), "DATA_PLANE_APP_NAME required but not set")

	dataPlaneSLIAppRouteURL = fmt.Sprintf("http://%s.%s", dataPlaneAppName, cfAppDomain)

	httpClient = http.Client{Timeout: 1 * time.Second}
})

func timeGetRequest(requestURL string) (*http.Response, error, time.Duration) {
	start := time.Now()
	resp, err := httpClient.Get(requestURL)
	requestLatency := time.Since(start)
	return resp, err, requestLatency
}
