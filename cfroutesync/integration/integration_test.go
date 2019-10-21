package integration_test

import (
	"fmt"
	"os/exec"
	"time"

	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/ccclient"
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
		Eventually(session.Out, 10*time.Second).Should(gbytes.Say("metacontroller"))

		out, err := te.kubectl("get", "services", "--all-namespaces")
		Expect(err).NotTo(HaveOccurred())
		fmt.Println(string(out))

	})
})
