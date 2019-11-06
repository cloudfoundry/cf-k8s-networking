package integration_test

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"code.cloudfoundry.org/cf-networking-helpers/testsupport/ports"

	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/ccclient"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/webhook"
)

var _ = Describe("Integration of cfroutesync with UAA, CC and Meta Controller", func() {
	var (
		te                 *TestEnv
		cfroutesyncSession *gexec.Session
	)

	BeforeEach(func() {
		var err error
		te, err = NewTestEnv(GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		_, err = te.kubectl("apply", "-f", "fixtures/crds/routebulksync.yaml")
		Expect(err).NotTo(HaveOccurred())

		_, err = te.kubectl("create", "namespace", "cf-workloads")
		Expect(err).NotTo(HaveOccurred())

		// apply the parent object that metacontroller watches
		_, err = te.kubectl("apply", "-f", "fixtures/routebulksync.yaml")
		Expect(err).NotTo(HaveOccurred())

		initializeFakeData(te)

		cfroutesyncSession = startAndRegister(te)
	})

	AfterEach(func() {
		cfroutesyncSession.Terminate().Wait("2s")
		te.Cleanup()
	})

	Specify("cfroutesync creates the expected k8s resources", func() {
		serviceMap := map[string]webhook.Service{}
		Eventually(func() (map[string]webhook.Service, error) {
			err := te.getResourcesByName("services", "cf-workloads", serviceMap)
			return serviceMap, err
		}, "1s", "0.1s").Should(HaveLen(3))
		Expect(serviceMap).To(HaveKey("s-destination-0"))
		Expect(serviceMap).To(HaveKey("s-destination-1"))
		Expect(serviceMap).To(HaveKey("s-destination-2"))

		virtualServiceMap := map[string]webhook.VirtualService{}
		Eventually(func() (map[string]webhook.VirtualService, error) {
			err := te.getResourcesByName("virtualservices", "cf-workloads", virtualServiceMap)
			return virtualServiceMap, err
		}, "1s", "0.1s").Should(HaveLen(3))
		Expect(virtualServiceMap).To(HaveKey("route-0-host.domain0.example.com"))
		Expect(virtualServiceMap).To(HaveKey("route-1-host.domain1.apps.internal"))
		Expect(virtualServiceMap).To(HaveKey(fmt.Sprintf("%s.domain1.apps.internal", longHostname)))
	})
})

func startAndRegister(te *TestEnv) *gexec.Session {
	webhookListenAddr := fmt.Sprintf("127.0.0.1:%d", ports.PickAPort())
	cmd := exec.Command(binaryPathCFRouteSync,
		"-c", te.CfRouteSyncConfigDir,
		"-l", webhookListenAddr,
		"-v", "6")
	cfroutesyncSession, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(cfroutesyncSession.Out).Should(gbytes.Say("starting webhook server"))
	Eventually(cfroutesyncSession.Out).Should(gbytes.Say("starting cc fetch loop"))
	Eventually(cfroutesyncSession.Out, 10*time.Second).Should(gbytes.Say("Fetched and put snapshot"))

	compositeController := fmt.Sprintf(`---
apiVersion: metacontroller.k8s.io/v1alpha1
kind: CompositeController
metadata:
  name: cfroutesync
spec:
  resyncPeriodSeconds: 5
  parentResource:
    apiVersion: apps.cloudfoundry.org/v1alpha1
    resource: routebulksyncs
  childResources:
    - apiVersion: v1
      resource: services
      updateStrategy:
        method: InPlace
    - apiVersion: networking.istio.io/v1alpha3
      resource: virtualservices
      updateStrategy:
        method: InPlace
  hooks:
    sync:
      webhook:
        url: http://%s/sync`, webhookListenAddr)
	Expect(te.kubectlApplyResource(compositeController)).To(Succeed())
	Eventually(cfroutesyncSession.Out, 10*time.Second).Should(gbytes.Say("metacontroller webhook request received"))
	return cfroutesyncSession
}

const DNSLabelMaxLength = 63

var longHostname = strings.Repeat("a", DNSLabelMaxLength)

func initializeFakeData(te *TestEnv) {
	te.FakeCC.Data.Routes = []ccclient.Route{
		ccclient.Route{
			Guid: "route-0-guid",
			Host: "route-0-host",
			Path: "route-0-path",
			Url:  "route-0-url",
		},
		ccclient.Route{
			Guid: "route-1-guid",
			Host: "route-1-host",
			Path: "route-1-path",
			Url:  "route-1-url",
		},
		ccclient.Route{
			Guid: "route-2-guid",
			Host: longHostname,
			Path: "route-2-path",
			Url:  "route-2-url",
		},
	}
	te.FakeCC.Data.Routes[0].Relationships.Domain.Data.Guid = "domain-0"
	te.FakeCC.Data.Routes[1].Relationships.Domain.Data.Guid = "domain-1"
	te.FakeCC.Data.Routes[2].Relationships.Domain.Data.Guid = "domain-1"
	te.FakeCC.Data.Domains = []ccclient.Domain{
		{
			Guid:     "domain-0",
			Name:     "domain0.example.com",
			Internal: false,
		},
		{
			Guid:     "domain-1",
			Name:     "domain1.apps.internal",
			Internal: true,
		},
	}
	te.FakeCC.Data.Destinations = map[string][]ccclient.Destination{}
	te.FakeCC.Data.Destinations["route-0-guid"] = []ccclient.Destination{
		{
			Guid:   "destination-0",
			Port:   8000,
			Weight: nil,
		},
	}
	te.FakeCC.Data.Destinations["route-0-guid"][0].App.Guid = "destination-0-app-guid"
	te.FakeCC.Data.Destinations["route-0-guid"][0].App.Process.Type = "destination-0-process-type"
	te.FakeCC.Data.Destinations["route-1-guid"] = []ccclient.Destination{
		{
			Guid:   "destination-1",
			Port:   8000,
			Weight: nil,
		},
	}
	te.FakeCC.Data.Destinations["route-1-guid"][0].App.Guid = "destination-1-app-guid"
	te.FakeCC.Data.Destinations["route-1-guid"][0].App.Process.Type = "destination-1-process-type"
	te.FakeCC.Data.Destinations["route-2-guid"] = []ccclient.Destination{
		{
			Guid:   "destination-2",
			Port:   8000,
			Weight: nil,
		},
	}
	te.FakeCC.Data.Destinations["route-2-guid"][0].App.Guid = "destination-2-app-guid"
	te.FakeCC.Data.Destinations["route-2-guid"][0].App.Process.Type = "destination-2-process-type"
}
