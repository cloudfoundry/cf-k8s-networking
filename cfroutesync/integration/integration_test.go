package integration_test

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"code.cloudfoundry.org/cf-networking-helpers/testsupport/ports"

	"encoding/json"

	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/ccclient"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/webhook"
)

var _ = Describe("Integration of cfroutesync with UAA, CC and Meta Controller", func() {
	var (
		te                 *TestEnv
		cfroutesyncSession *gexec.Session
		webhookListenAddr  string
	)

	BeforeEach(func() {
		var err error
		te, err = NewTestEnv(GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		_, err = te.kubectl("create", "namespace", "cf-workloads")
		Expect(err).NotTo(HaveOccurred())

		_, err = te.kubectl("apply", "-f", "../crds/routebulksync.yaml")
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() error {
			_, err = te.kubectl("apply", "-f", "fixtures/routebulksync.yaml")
			return err
		}).Should(Succeed())

		initializeFakeData(te)
		webhookListenAddr = fmt.Sprintf("127.0.0.1:%d", ports.PickAPort())
		cfroutesyncSession = startAndRegister(te, webhookListenAddr)
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
		Expect(virtualServiceMap).To(HaveKey(webhook.VirtualServiceName("route-0-host.domain0.example.com")))
		Expect(virtualServiceMap).To(HaveKey(webhook.VirtualServiceName("route-1-host.domain1.apps.internal")))
		Expect(virtualServiceMap).To(HaveKey(webhook.VirtualServiceName(fmt.Sprintf("%s.domain1.apps.internal", longHostname))))

		// check that there isn't a hot-loop: https://github.com/GoogleCloudPlatform/metacontroller/issues/171
		getParentGenerations := func() map[int64]bool {
			lines := strings.Split(strings.TrimSpace(string(cfroutesyncSession.Out.Contents())), "\n")
			generations := map[int64]bool{}
			for _, line := range lines {
				structured := parseLogLine(line)
				generations[structured.Request.Parent.ObjectMeta.Generation] = true
			}
			return generations
		}
		expectedParentGenerations := map[int64]bool{0: true, 1: true} // a hot loop would make many more generations
		Consistently(getParentGenerations, "2s", "0.5s").Should(Equal(expectedParentGenerations))
	})

	Specify("expose scrapable metrics", func() {
		client := http.Client{}
		resp, err := client.Get(fmt.Sprintf("http://%s/%s", webhookListenAddr, "metrics"))
		Expect(err).NotTo(HaveOccurred())

		output := gbytes.BufferReader(resp.Body)
		Eventually(output).Should(gbytes.Say("cfroutesync_fetched_routes %d", len(te.FakeCC.Data.Routes)))
		Eventually(output).Should(gbytes.Say("cfroutesync_last_updated_at"))
		timestamp1 := parseTimestamp(output)

		By("Deleting a route updates the fetched_routes metric")
		te.FakeCC.Data.Routes = te.FakeCC.Data.Routes[:len(te.FakeCC.Data.Routes)-1]

		Eventually(func() *gbytes.Buffer {
			resp, err := client.Get(fmt.Sprintf("http://%s/%s", webhookListenAddr, "metrics"))
			Expect(err).NotTo(HaveOccurred())

			// we cannot use gbytes.BufferReader(resp.Body) because BufferReader fills the buffer async
			b := gbytes.NewBuffer()
			io.Copy(b, resp.Body)
			b.Close()
			return b
		}, "11s", "1s").Should(gbytes.Say("cfroutesync_fetched_routes %d", len(te.FakeCC.Data.Routes)))

		By("Checking that the last_updated_at metric time changes")
		resp, err = client.Get(fmt.Sprintf("http://%s/%s", webhookListenAddr, "metrics"))
		Expect(err).NotTo(HaveOccurred())

		output2 := gbytes.BufferReader(resp.Body)
		Eventually(output2).Should(gbytes.Say("cfroutesync_last_updated_at"))

		timestamp2 := parseTimestamp(output2)
		Expect(timestamp1).ToNot(Equal(timestamp2))
	})
})

// parseTimestamp returns a string of the timestamp in metrics input
// metrics input contains the following metrics info:
// 		# HELP cfroutesync_last_updated_at Unix timestamp indicating last successful sync
//      # TYPE cfroutesync_last_updated_at gauge
//      cfroutesync_last_updated_at 1.5761903739383752e+09
func parseTimestamp(metrics *gbytes.Buffer) string {
	re := regexp.MustCompile(`cfroutesync_last_updated_at.*?(?:\n)`)
	timestampLine := string(metrics.Contents())
	timestampLine = re.FindAllString(timestampLine, 4)[2]
	return regexp.MustCompile(`\s+`).Split(timestampLine, -1)[1]
}

type WebhookRequestLogLine struct {
	Msg     string
	Request webhook.SyncRequest
}

func parseLogLine(logLine string) WebhookRequestLogLine {
	var res WebhookRequestLogLine
	json.Unmarshal([]byte(logLine), &res)
	return res
}

func startAndRegister(te *TestEnv, webhookListenAddr string) *gexec.Session {
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
			Destinations: []ccclient.Destination{
				{
					Guid:   "destination-0",
					Port:   8000,
					Weight: nil,
					App: ccclient.App{
						Guid: "destination-0-app-guid",
						Process: ccclient.Process{
							Type: "destination-0-process-type",
						},
					},
				},
			},
		},
		ccclient.Route{
			Guid: "route-1-guid",
			Host: "route-1-host",
			Path: "route-1-path",
			Url:  "route-1-url",
			Destinations: []ccclient.Destination{
				{
					Guid:   "destination-1",
					Port:   8000,
					Weight: nil,
					App: ccclient.App{
						Guid: "destination-1-app-guid",
						Process: ccclient.Process{
							Type: "destination-1-process-type",
						},
					},
				},
			},
		},
		ccclient.Route{
			Guid: "route-2-guid",
			Host: longHostname,
			Path: "route-2-path",
			Url:  "route-2-url",
			Destinations: []ccclient.Destination{
				{
					Guid:   "destination-2",
					Port:   8000,
					Weight: nil,
					App: ccclient.App{
						Guid: "destination-2-app-guid",
						Process: ccclient.Process{
							Type: "destination-2-process-type",
						},
					},
				},
			},
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

	te.FakeCC.Data.Routes[0].Relationships.Space.Data.Guid = "space-0-guid"
	te.FakeCC.Data.Routes[1].Relationships.Space.Data.Guid = "space-1-guid"
	te.FakeCC.Data.Routes[2].Relationships.Space.Data.Guid = "space-1-guid"
	te.FakeCC.Data.Spaces = []ccclient.Space{
		{
			Guid: "space-0-guid",
		},
		{
			Guid: "space-1-guid",
		},
	}
	te.FakeCC.Data.Spaces[0].Relationships.Organization.Data.Guid = "org-0-guid"
	te.FakeCC.Data.Spaces[1].Relationships.Organization.Data.Guid = "org-1-guid"
}
