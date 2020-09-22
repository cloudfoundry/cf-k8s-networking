package uptime_test

import (
	"crypto/tls"
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
func getEnvOrUseDefault(envVarName, defaultValue string) string {
	if envVal, ok := os.LookupEnv(envVarName); ok {
		return envVal
	}
	return defaultValue
}

var _ = BeforeSuite(func() {
	var err error

	upgradeDiscoveryTimeout, err = time.ParseDuration(getEnvOrUseDefault("UPGRADE_DISCOVERY_TIMEOUT", "1m"))
	Expect(err).NotTo(HaveOccurred(), "UPGRADE_DISCOVERY_TIMEOUT malformed")

	dataPlaneSLOMaxRequestLatency, err = time.ParseDuration(getEnvOrUseDefault("DATA_PLANE_SLO_MAX_REQUEST_LATENCY", "300ms"))
	Expect(err).NotTo(HaveOccurred(), "DATA_PLANE_SLO_MAX_REQUEST_LATENCY malformed")

	dataPlaneSLOPercentage, err = strconv.ParseFloat(getEnvOrUseDefault("DATA_PLANE_SLO_PERCENTAGE", "0.95"), FLOAT_BIT_SIZE)
	Expect(err).NotTo(HaveOccurred(), "DATA_PLANE_SLO_PERCENTAGE malformed")

	controlPlaneSLORoutePropagationTime, err = time.ParseDuration(getEnvOrUseDefault("CONTROL_PLANE_SLO_MAX_ROUTE_PROPAGATION_TIME", "10s"))
	Expect(err).NotTo(HaveOccurred(), "CONTROL_PLANE_SLO_MAX_ROUTE_PROPAGATION_TIME malformed")

	controlPlaneSLOSampleCaptureTime, err = time.ParseDuration(getEnvOrUseDefault("CONTROL_PLANE_SLO_SAMPLE_CAPTURE_TIME", "10s"))
	Expect(err).NotTo(HaveOccurred(), "CONTROL_PLANE_SLO_SAMPLE_CAPTURE_TIME malformed")

	controlPlaneSLODataPlaneAvailabilityPercentage, err = strconv.ParseFloat(getEnvOrUseDefault("CONTROL_PLANE_SLO_DATA_PLANE_AVAILABILITY_PERCENTAGE", "0.99"), FLOAT_BIT_SIZE)
	Expect(err).NotTo(HaveOccurred(), "CONTROL_PLANE_SLO_DATA_PLANE_AVAILABILITY_PERCENTAGE malformed")

	cfAppDomain := getEnvOrUseDefault("CF_APP_DOMAIN", "apps.ci-upgrade-cf.routing.lol")

	controlPlaneAppName := getEnvOrUseDefault("CONTROL_PLANE_APP_NAME", "upgrade-control-plane-sli")

	dataPlaneAppName := getEnvOrUseDefault("DATA_PLANE_APP_NAME", "upgrade-data-plane-sli")

	dataPlaneSLIAppRouteURL = fmt.Sprintf("https://%s.%s", dataPlaneAppName, cfAppDomain)

	httpClient = http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 1 * time.Second,
	}
})

func timeGetRequest(requestURL string) (*http.Response, error, time.Duration) {
	start := time.Now()
	resp, err := httpClient.Get(requestURL)
	requestLatency := time.Since(start)
	return resp, err, requestLatency
}
