package stress_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os/exec"
	"strings"
	"testing"
	"text/template"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"sigs.k8s.io/kind/pkg/cluster"
)

func TestStress(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Route Controller Stress Tests Suite")
}

var (
	kubectl kubectlRunner
	ytt     yttRunner
)

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(5 * time.Minute)

	kubectl = CreateKindCluster()

	// Deploy Route CRD
	session, err := kubectl.Run("apply", "-f", "../../config/crd/networking.cloudfoundry.org_routes.yaml")
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))

	// Deploy Istio's Virtual Service CRD
	session, err = kubectl.Run("apply", "-f", "../integration/fixtures/istio-virtual-service.yaml")
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))
})

var _ = AfterSuite(func() {
	kubectl.DestroyCluster()
})

type kubectlRunner struct {
	kubeconfigFilePath string
	clusterName        string
}

func CreateKindCluster() kubectlRunner {
	name := fmt.Sprintf("stress-tests-%d", rand.Uint64())
	provider := cluster.NewProvider()
	err := provider.Create(name)
	// retry once
	if err != nil {
		time.Sleep(5 * time.Second)
		err = provider.Create(name)
	}

	Expect(err).NotTo(HaveOccurred())

	kubeConfig, err := provider.KubeConfig(name, false)
	Expect(err).NotTo(HaveOccurred())

	kubeConfigFile, err := ioutil.TempFile("", fmt.Sprintf("kubeconfig-%s", name))
	Expect(err).NotTo(HaveOccurred())
	defer kubeConfigFile.Close()

	_, err = kubeConfigFile.Write([]byte(kubeConfig))
	Expect(err).NotTo(HaveOccurred())

	return kubectlRunner{clusterName: name, kubeconfigFilePath: kubeConfigFile.Name()}
}

func (k kubectlRunner) DestroyCluster() {
	provider := cluster.NewProvider()
	err := provider.Delete(k.clusterName, k.kubeconfigFilePath)
	Expect(err).NotTo(HaveOccurred())
}

func (k kubectlRunner) Run(kubectlCommandArgs ...string) (*gexec.Session, error) {
	cmd := k.generateCommand(nil, kubectlCommandArgs...)
	return gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
}

func (k kubectlRunner) RunWithStdin(stdin io.Reader, kubectlCommandArgs ...string) (*gexec.Session, error) {
	cmd := k.generateCommand(stdin, kubectlCommandArgs...)
	return gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
}

func (k kubectlRunner) generateCommand(stdin io.Reader, kubectlCommandArgs ...string) *exec.Cmd {
	fmt.Fprintf(GinkgoWriter, "+ kubectl %s\n", strings.Join(kubectlCommandArgs, " "))
	cmd := exec.Command("kubectl", kubectlCommandArgs...)
	cmd.Env = append(cmd.Env, "KUBECONFIG="+k.kubeconfigFilePath)
	if stdin != nil {
		cmd.Stdin = stdin
	}

	return cmd
}

func (k kubectlRunner) GetNumberOf(resourceName string) int {
	session, err := k.Run("get", resourceName, "--no-headers")
	if err != nil {
		return 0
	}

	session.Wait(5 * time.Minute)

	if session.ExitCode() != 0 {
		return 0
	}

	return strings.Count(string(session.Out.Contents()), "\n")
}

type yttRunner struct {
}

func (y yttRunner) Run(yttCommandArgs ...string) (*gexec.Session, error) {
	fmt.Fprintf(GinkgoWriter, "+ ytt %s\n", strings.Join(yttCommandArgs, " "))
	cmd := exec.Command("ytt", yttCommandArgs...)
	return gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
}

func buildRoutes(numberOfRoutes int) io.Reader {
	routeTmpl, err := template.ParseFiles("fixtures/route_template.yml")
	Expect(err).NotTo(HaveOccurred())

	type Route struct {
		Name            string
		Host            string
		Path            string
		Domain          string
		DestinationGUID string
		AppGUID         string
	}

	var routesBuilder strings.Builder

	for i := 0; i < numberOfRoutes; i++ {
		route := Route{
			Name:            fmt.Sprintf("route-%d", i),
			Host:            fmt.Sprintf("hostname-%d", i),
			Path:            fmt.Sprintf("/%d", i),
			Domain:          "apps.example.com",
			DestinationGUID: fmt.Sprintf("destination-guid-%d", i),
			AppGUID:         fmt.Sprintf("app-guid-%d", i),
		}

		// Create a new YAML document for each Route definition
		_, err := routesBuilder.WriteString("---\n")
		Expect(err).NotTo(HaveOccurred())

		// Evaluate the Route template and write the resulting Route definition to routesBuilder
		err = routeTmpl.Execute(&routesBuilder, route)
		Expect(err).NotTo(HaveOccurred())
	}

	return strings.NewReader(routesBuilder.String())
}
