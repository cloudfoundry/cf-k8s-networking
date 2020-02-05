package acceptance_test

import (
	"code.cloudfoundry.org/cf-k8s-networking/acceptance/cfg"
	"fmt"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	"github.com/onsi/gomega/gexec"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	kubectl              *KubeCtl
	TestSetup            *workflowhelpers.ReproducibleTestSuiteSetup
	AppGuid              string
	SysComponentSelector string
)

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)

	config, err := cfg.NewConfig(
		os.Getenv("CONFIG"),
		os.Getenv("KUBECONFIG"),
		os.Getenv("CONFIG_KEEP_CLUSTER") != "",
		os.Getenv("CONFIG_KEEP_CF") != "",
	)

	if err != nil {
		defer GinkgoRecover()
		fmt.Println("Failed to load config.")
		t.Fail()
	}

	kubectl = &KubeCtl{kubeConfigPath: config.KubeConfigPath}

	var _ = SynchronizedBeforeSuite(func() []byte {
		_, err := kubectl.Run("cluster-info")
		if err != nil {
			panic(err)
		}

		SysComponentSelector = createSystemComponent()

		TestSetup = workflowhelpers.NewTestSuiteSetup(config)
		TestSetup.Setup()
		AppGuid = pushApp(generator.PrefixedRandomName("ACCEPTANCE", "app"))
		return []byte{}
	}, func(b []byte) {
		SetDefaultEventuallyTimeout(1 * time.Minute)
		SetDefaultEventuallyPollingInterval(1 * time.Second)
	})

	SynchronizedAfterSuite(func() {}, func() {
		if TestSetup != nil && !config.KeepCFChanges {
			TestSetup.Teardown()
		}

		if !config.KeepClusterChanges {
			destroySystemComponent()
		}
	})

	RunSpecs(t, "Acceptance Suite")
}

type KubeCtl struct {
	kubeConfigPath string
}

func (kc *KubeCtl) Run(args ...string) ([]byte, error) {
	cmd := exec.Command("kubectl", args...)
	cmd.Env = []string{
		fmt.Sprintf("KUBECONFIG=%s", kc.kubeConfigPath),
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
		fmt.Sprintf("HOME=%s", filepath.Dir(kc.kubeConfigPath)), // fixme because kubectl will create another .kube folder inside our provided kubeConfgPath
	}
	fmt.Fprintf(GinkgoWriter, "+ Run %s\n", strings.Join(args, " "))
	output, err := cmd.CombinedOutput()
	GinkgoWriter.Write(output)
	return output, err
}

func createSystemComponent() string {
	cmd := exec.Command("kapp", "deploy", "-n", systemNamespace, "-a", "system-component", "-y", "-f", "./assets/system-component.yml")
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	session.Wait(2 * time.Minute)

	return "app=test-system-component"
}

func destroySystemComponent() {
	cmd := exec.Command("kapp", "delete", "-n", systemNamespace, "-a", "system-component", "-y")
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	session.Wait(1 * time.Minute)
}

func pushApp(name string) string {
	session := cf.Cf("push",
		name,
		"-o", "cfrouting/httpbin8080",
		"-u", "http",
	)
	Expect(session.Wait(120 * time.Second)).To(gexec.Exit(0))

	guid := strings.TrimSpace(string(cf.Cf("app", name, "--guid").Wait().Out.Contents()))

	return guid
}
