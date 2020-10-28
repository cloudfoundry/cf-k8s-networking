package integration_test

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/onsi/gomega/gexec"
	"sigs.k8s.io/kind/pkg/cluster"

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

type service struct {
	Metadata metadata
	Spec     serviceSpec
}

type serviceSpec struct {
	Ports []serviceSpecPort
}

type serviceSpecPort struct {
	TargetPort int
}

type metadata struct {
	Name string
}

func startRouteController(kubeConfigPath, gateway string, ingressProvider string) *gexec.Session {
	cmd := exec.Command(routeControllerBinaryPath)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%s", kubeConfigPath))
	cmd.Env = append(cmd.Env, fmt.Sprintf("ISTIO_GATEWAY_NAME=%s", gateway))
	cmd.Env = append(cmd.Env, fmt.Sprintf("RESYNC_INTERVAL=%s", "5"))
	cmd.Env = append(cmd.Env, fmt.Sprintf("INGRESS_PROVIDER=%s", ingressProvider))

	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	return session
}

func createKindCluster(name string) string {
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

	kubeConfigPath, err := ioutil.TempFile("", fmt.Sprintf("kubeconfig-%s", name))
	Expect(err).NotTo(HaveOccurred())
	defer kubeConfigPath.Close()

	_, err = kubeConfigPath.Write([]byte(kubeConfig))
	Expect(err).NotTo(HaveOccurred())

	return kubeConfigPath.Name()
}

func deleteKindCluster(name, kubeConfigPath string) {
	provider := cluster.NewProvider()
	err := provider.Delete(name, kubeConfigPath)
	Expect(err).NotTo(HaveOccurred())
}

func kustomizeConfigCRD() ([]byte, error) {
	args := []string{"kustomize", filepath.Join("..", "config", "crd")}
	cmd := exec.Command("kubectl", args...)
	cmd.Stderr = GinkgoWriter

	fmt.Fprintf(GinkgoWriter, "+ kubectl %s\n", strings.Join(args, " "))

	output, err := cmd.Output()

	return output, err
}

func kubectlWithConfig(kubeConfigPath string, stdin io.Reader, args ...string) ([]byte, error) {
	if len(kubeConfigPath) == 0 {
		return nil, errors.New("kubeconfig path cannot be empty")
	}
	cmd := exec.Command("kubectl", args...)
	cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%s", kubeConfigPath))
	if stdin != nil {
		cmd.Stdin = stdin
	}

	fmt.Fprintf(GinkgoWriter, "+ kubectl %s\n", strings.Join(args, " "))
	output, err := cmd.CombinedOutput()
	return output, err
}
