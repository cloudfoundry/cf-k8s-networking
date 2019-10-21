package integration_test

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/ccclient"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/webhook"
	"code.cloudfoundry.org/cf-networking-helpers/testsupport/ports"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Integration of cfroutesync with UAA, CC and Meta Controller", func() {
	var (
		te                *TestEnv
		webhookListenAddr string
	)

	BeforeEach(func() {
		var err error
		te, err = NewTestEnv(GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		webhookListenAddr = fmt.Sprintf("127.0.0.1:%d", ports.PickAPort())

		out, err := te.kubectl("get", "all", "--all-namespaces")
		fmt.Println(string(out))
		Expect(err).NotTo(HaveOccurred())

		// apply the CRDs for metacontroller, istio, and cfroutesync
		out, err = te.kubectl("apply", "-f", "fixtures/crds/metacontroller_crds.yaml")
		fmt.Println(string(out))
		Expect(err).NotTo(HaveOccurred())

		out, err = te.kubectl("apply", "-f", "fixtures/crds/istio_crds.yaml")
		fmt.Println(string(out))
		Expect(err).NotTo(HaveOccurred())

		out, err = te.kubectl("apply", "-f", "fixtures/crds/routebulksync.yaml")
		fmt.Println(string(out))
		Expect(err).NotTo(HaveOccurred())

		// apply the parent object that metacontroller watches
		out, err = te.kubectl("apply", "-f", "fixtures/routebulksync.yaml")
		fmt.Println(string(out))
		Expect(err).NotTo(HaveOccurred())

		compositefile, err := createCompositeController(webhookListenAddr)
		Expect(err).NotTo(HaveOccurred())

		out, err = te.kubectl("apply", "-f", compositefile)
		fmt.Println(string(out))
		Expect(err).NotTo(HaveOccurred())

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
				Host: "route-2-host",
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
	})

	AfterEach(func() {
		te.Cleanup()
	})

	Specify("cfroutesync boots and stays running", func() {
		cmd := exec.Command("metacontroller", "-logtostderr", "-client-config-path", te.KubeConfigPath, "-v", "6")
		metacontrollerSession, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		defer func() {
			metacontrollerSession.Terminate().Wait("2s")
		}()

		Expect(err).NotTo(HaveOccurred())

		cmd = exec.Command(binaryPathCFRouteSync, "-c", te.ConfigDir, "-l", webhookListenAddr, "-v", "6")
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		defer func() {
			session.Terminate().Wait("2s")
		}()

		Eventually(session.Out).Should(gbytes.Say("starting webhook server"))
		Eventually(session.Out).Should(gbytes.Say("starting cc fetch loop"))
		Eventually(session.Out, 10*time.Second).Should(gbytes.Say("Fetched and put snapshot"))
		Eventually(session.Out, 10*time.Second).Should(gbytes.Say("metacontroller"))

		type servicesResponse struct {
			Items []webhook.Service `json:"items"`
		}
		actualServicesResponse := &servicesResponse{}
		Eventually(func() ([]webhook.Service, error) {
			out, err := te.kubectl("get", "services", "-n", "cf-workloads", "-o", "json")
			if err != nil {
				return []webhook.Service{}, err
			}
			err = json.Unmarshal(out, actualServicesResponse)
			if err != nil {
				return []webhook.Service{}, err
			}

			return actualServicesResponse.Items, nil
		}, "1s", "0.1s").Should(HaveLen(3))

		type virtualServicesResponse struct {
			Items []webhook.VirtualService `json:"items"`
		}
		actualVirtualServicesResponse := &virtualServicesResponse{}
		Eventually(func() ([]webhook.VirtualService, error) {
			out, err := te.kubectl("get", "virtualservices", "-n", "cf-workloads", "-o", "json")
			if err != nil {
				return []webhook.VirtualService{}, err
			}
			err = json.Unmarshal(out, actualVirtualServicesResponse)
			if err != nil {
				return []webhook.VirtualService{}, err
			}

			return actualVirtualServicesResponse.Items, nil
		}, "1s", "0.1s").Should(HaveLen(3))
	})
})
