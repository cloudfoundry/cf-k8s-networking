package acceptance_test

import (
	"encoding/json"
	"fmt"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	"github.com/onsi/gomega/gexec"
	"io/ioutil"
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

	config, err := NewConfig(os.Getenv("CONFIG"))
	if err != nil {
		defer GinkgoRecover()
		fmt.Println("Failed to load config.")
		t.Fail()
	}

	kubectl = &KubeCtl{kubeConfigPath:config.KubeConfigPath}

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
		if TestSetup != nil {
			TestSetup.Teardown()
		}
		destroySystemComponent()
	})

	RunSpecs(t, "Acceptance Suite")
}

type Config struct {
	KubeConfigPath string `json:"kubeconfig_path"`

	API            string `json:"api"`
	AdminUser      string `json:"admin_user"`
	AdminPassword  string `json:"admin_password"`

	ExistingUser         string `json:"existing_user"`
	ExistingUserPassword string `json:"existing_user_password"`
	ShouldKeepUser       bool   `json:"keep_user_at_suite_end"`
	UseExistingUser      bool   `json:"use_existing_user"`

	UseExistingOrganization bool   `json:"use_existing_organization"`
	ExistingOrganization    string `json:"existing_organization"`
}

func (c *Config) GetAdminUser() string {
	return c.AdminUser
}

func (c *Config) GetAdminPassword() string {
	return c.AdminPassword
}

func (c *Config) GetUseExistingOrganization() bool {
	return c.UseExistingOrganization
}

func (c *Config) GetUseExistingSpace() bool {
	return false
}

func (c *Config) GetExistingOrganization() string {
	return c.ExistingOrganization
}

func (c *Config) GetExistingSpace() string {
	panic("implement me")
}

func (c *Config) GetUseExistingUser() bool {
	return c.UseExistingUser
}

func (c *Config) GetExistingUser() string {
	return c.ExistingUser
}

func (c *Config) GetExistingUserPassword() string {
	return c.ExistingUserPassword
}

func (c *Config) GetShouldKeepUser() bool {
	return c.ShouldKeepUser
}

func (c *Config) GetConfigurableTestPassword() string {
	return ""
}

func (c *Config) GetAdminClient() string {
	return ""
}

func (c *Config) GetAdminClientSecret() string {
	return ""
}

func (c *Config) GetExistingClient() string {
	return ""
}

func (c *Config) GetExistingClientSecret() string {
	panic("implement me")
}

func (c *Config) GetApiEndpoint() string {
	return c.API
}

func (c *Config) GetSkipSSLValidation() bool {
	return true
}

func (c *Config) GetNamePrefix() string {
	return "ACCEPTANCE"
}

func (c *Config) GetScaledTimeout(timeout time.Duration) time.Duration {
	return time.Duration(float64(timeout) * 2)
}

func NewConfig(configPath string) (*Config, error) {
	configFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config %v", err)
	}

	config := &Config{}
	err = json.Unmarshal([]byte(configFile), config)

	if err != nil {
		return nil, fmt.Errorf("error parsing json %v", err)
	}

	return config, nil
}

type KubeCtl struct {
	kubeConfigPath string
}

func (kc *KubeCtl) Run(args ...string) ([]byte, error) {
	cmd := exec.Command("kubectl", args...)
	cmd.Env = []string{
		fmt.Sprintf("KUBECONFIG=%s", kc.kubeConfigPath),
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
		fmt.Sprintf("HOME=%s", filepath.Dir(kc.kubeConfigPath)), // fixme
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
		"-o", "cfrouting/httpbin8080")
	Expect(session.Wait(120 * time.Second)).To(gexec.Exit(0))

	guid := strings.TrimSpace(string(cf.Cf("app", name, "--guid").Wait().Out.Contents()))

	return guid
}
