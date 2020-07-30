package scale_test

import (
	"os"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestScale(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Scale Suite")
}

var (
	domain          string
	cleanup         bool
	numApps         int
	numAppsPerSpace int
	orgNamePrefix   string
	spaceNamePrefix string
)

var _ = BeforeSuite(func() {
	var found bool
	var err error

	orgNamePrefix = "scale-tests"
	spaceNamePrefix = "scale-tests"

	domain, found = os.LookupEnv("DOMAIN")
	Expect(found).To(BeTrue(), "DOMAIN environment variable required but not set")

	cleanupStr := os.Getenv("CLEANUP")
	cleanup = cleanupStr == "true" || cleanupStr == "1"

	numAppsStr, found := os.LookupEnv("NUMBER_OF_APPS")
	Expect(found).To(BeTrue(), "NUMBER_OF_APPS environment variable required but not set")
	numApps, err = strconv.Atoi(numAppsStr)
	Expect(err).NotTo(HaveOccurred(), "NUMBER_OF_APPS environment variable malformed")

	numAppsPerSpace = 10
})
