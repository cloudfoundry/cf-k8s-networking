package integration_test

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/onsi/gomega/gexec"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	k8sApiServer "sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/ccclient"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/cfg"
)

type TestEnv struct {
	lock sync.Mutex

	TempDir              string
	CfRouteSyncConfigDir string

	FakeUAA struct {
		Handler http.Handler
		Server  *httptest.Server
	}
	FakeCC struct {
		Handler http.Handler
		Server  *httptest.Server
		Data    struct {
			Domains      []ccclient.Domain
			Routes       []ccclient.Route
			Destinations map[string][]ccclient.Destination
		}
	}
	K8sApiServerEnv *k8sApiServer.Environment
	KubeConfigPath  string

	GalleySession         *gexec.Session
	MetaControllerSession *gexec.Session

	TestOutput io.Writer
}

func (te *TestEnv) FakeUAAServeHTTP(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(struct {
		AccessToken string `json:"access_token"`
	}{"fake-access-token"})
}

func (te *TestEnv) FakeCCServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.Contains(r.URL.Path, "domains"):
		json.NewEncoder(w).Encode(map[string]interface{}{
			"resources": te.FakeCC.Data.Domains,
		})
	case strings.Contains(r.URL.Path, "destinations"):
		routeGUIDs := regexp.MustCompile("/v3/routes/(.*)/destinations").FindStringSubmatch(r.URL.Path)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"destinations": te.FakeCC.Data.Destinations[routeGUIDs[1]],
		})
	case strings.Contains(r.URL.Path, "routes"):
		json.NewEncoder(w).Encode(map[string]interface{}{
			"resources": te.FakeCC.Data.Routes,
		})
	default:
		log.WithFields(log.Fields{"server": "fakeCC", "request": r}).Error("unrecognized request")
		panic("request for unimplemented route on fake CC")
	}
}

func NewTestEnv(testOutput io.Writer) (*TestEnv, error) {
	tempDir, err := ioutil.TempDir("", "cfroutesync-integ-test")
	if err != nil {
		return nil, err
	}

	te := &TestEnv{
		TempDir:    tempDir,
		TestOutput: testOutput,
	}

	te.FakeUAA.Handler = http.HandlerFunc(te.FakeUAAServeHTTP)
	te.FakeUAA.Server = httptest.NewTLSServer(te.FakeUAA.Handler)
	te.FakeCC.Handler = http.HandlerFunc(te.FakeCCServeHTTP)
	te.FakeCC.Server = httptest.NewTLSServer(te.FakeCC.Handler)

	if err := te.setupConfigDirForCfroutesync(); err != nil {
		return nil, fmt.Errorf("setup config for cfroutesync: %w", err)
	}

	logf.SetLogger(logf.ZapLoggerTo(te.TestOutput, true /* development */))
	te.K8sApiServerEnv = &k8sApiServer.Environment{
		KubeAPIServerFlags: getApiServerFlags(),
	}
	apiServerConfig, err := te.K8sApiServerEnv.Start()
	if err != nil {
		return nil, fmt.Errorf("starting fake api server: %w", err)
	}
	if err := te.createKubeConfig(apiServerConfig); err != nil {
		return nil, fmt.Errorf("writing kube config: %w", err)
	}

	if _, err = te.kubectl("apply", "-f", "fixtures/crds/istio_crds.yaml"); err != nil {
		return nil, err
	}
	if err := te.startGalley(); err != nil {
		return nil, fmt.Errorf("starting galley: %w", err)
	}
	if _, err = te.kubectl("apply", "-f", "fixtures/istio-validating-admission-webhook.yaml"); err != nil {
		return nil, err
	}

	if err = te.setupAndStartMetaController(); err != nil {
		return nil, fmt.Errorf("starting metacontroller: %w", err)
	}

	if te.GalleySession.ExitCode() >= 0 {
		return nil, fmt.Errorf("galley exited unexpectedly with code: %d", te.GalleySession.ExitCode())
	}

	if err := te.eventuallyGalleyIsValidating(10, 2*time.Second); err != nil {
		return nil, err
	}

	return te, nil
}

func (te *TestEnv) setupConfigDirForCfroutesync() error {
	te.CfRouteSyncConfigDir = filepath.Join(te.TempDir, "cfroutesync-config")
	if err := os.MkdirAll(te.CfRouteSyncConfigDir, 0644); err != nil {
		return err
	}

	fakeUAACertBytes, err := tlsCertToPem(te.FakeUAA.Server.Certificate())
	if err != nil {
		return err
	}

	fakeCCCertBytes, err := tlsCertToPem(te.FakeCC.Server.Certificate())
	if err != nil {
		return err
	}

	for filename, contents := range map[string]string{
		cfg.FileUAABaseURL:      te.FakeUAA.Server.URL,
		cfg.FileUAAClientName:   "fake-uaa-client-name",
		cfg.FileUAAClientSecret: "fake-uaa-client-secret",
		cfg.FileUAACA:           string(fakeUAACertBytes),
		cfg.FileCCBaseURL:       te.FakeCC.Server.URL,
		cfg.FileCCCA:            string(fakeCCCertBytes),
	} {
		if err := ioutil.WriteFile(filepath.Join(te.CfRouteSyncConfigDir, filename), []byte(contents), 0644); err != nil {
			return err
		}
	}
	return nil
}

func (te *TestEnv) setupAndStartMetaController() error {
	var err error
	if _, err = te.kubectl("apply", "-f", "fixtures/crds/metacontroller_crds.yaml"); err != nil {
		return err
	}
	cmd := exec.Command("metacontroller",
		"-logtostderr",
		"-client-config-path", te.KubeConfigPath,
		"-v", "6",
		"-discovery-interval", "1s")
	te.MetaControllerSession, err = gexec.Start(cmd, te.TestOutput, te.TestOutput)
	return err
}

func getApiServerFlags() []string {
	apiServerFlags := make([]string, len(k8sApiServer.DefaultKubeAPIServerFlags))
	copy(apiServerFlags, k8sApiServer.DefaultKubeAPIServerFlags)
	for i, current := range apiServerFlags {
		if strings.HasPrefix(current, "--admission-control") {
			apiServerFlags[i] = "--enable-admission-plugins=ValidatingAdmissionWebhook"
		}
	}
	return apiServerFlags
}

func (te *TestEnv) startGalley() error {
	cmd := exec.Command("galley",
		"server",
		"--enable-server=false",
		"--enable-validation=true",
		"--validation-webhook-config-file", "./fixtures/istio-validating-admission-webhook.yaml",
		"--caCertFile", "./fixtures/galley-certs/galley-ca.crt",
		"--tlsCertFile", "./fixtures/galley-certs/galley-webhook.crt",
		"--tlsKeyFile", "./fixtures/galley-certs/galley-webhook.key",
		"--insecure",
		"--kubeconfig", te.KubeConfigPath,
	)
	var err error
	te.GalleySession, err = gexec.Start(cmd, te.TestOutput, te.TestOutput)
	if err != nil {
		return err
	}
	return nil
}

func (te *TestEnv) checkGalleyIsValidating() error {
	// attempt to apply invalid data
	outBytes, err := te.kubectl("apply", "-f", "./fixtures/invalid-virtual-service.yaml")
	out := string(outBytes)
	if err == nil {
		// it succeeded, clean-up
		_, errOnDelete := te.kubectl("delete", "-f", "./fixtures/invalid-virtual-service.yaml")
		if errOnDelete != nil {
			return fmt.Errorf("applying invalid data was successful (bad) and then we errored when attempting to delete it (even worse!): %w", err)
		}
		return fmt.Errorf("invalid virtual-service was admitted to the K8s API: %s", out)
	}

	const expectedErrorSnippet = `admission webhook "pilot.validation.istio.io" denied the request`
	if strings.Contains(out, expectedErrorSnippet) {
		fmt.Fprintf(te.TestOutput, "invalid data was rejected, it appears that the istio galley validating admission webhook is working\n")
		return nil
	}

	return fmt.Errorf("unexpected condition while applying invalid VirtualService: %w: %s", err, out)
}

func (te *TestEnv) eventuallyGalleyIsValidating(numPolls int, pollInterval time.Duration) error {
	var err error
	for i := 0; i < numPolls; i++ {
		err = te.checkGalleyIsValidating()
		if err == nil {
			return nil
		}
		time.Sleep(pollInterval)
	}
	return fmt.Errorf("timed out waiting for galley to start validating.  last error: %w", err)
}

func (te *TestEnv) Cleanup() {
	if te == nil {
		return
	}
	te.lock.Lock()
	defer te.lock.Unlock()

	if len(te.TempDir) > 0 {
		os.RemoveAll(te.TempDir)
		te.TempDir = ""
	}

	if te.FakeUAA.Server != nil {
		te.FakeUAA.Server.Close()
		te.FakeUAA.Server = nil
	}

	if te.FakeCC.Server != nil {
		te.FakeCC.Server.Close()
		te.FakeCC.Server = nil
	}

	if te.K8sApiServerEnv != nil {
		te.K8sApiServerEnv.Stop()
		te.K8sApiServerEnv = nil
	}

	if te.GalleySession != nil {
		te.GalleySession.Terminate().Wait("2s")
		te.GalleySession = nil
	}

	if te.MetaControllerSession != nil {
		te.MetaControllerSession.Terminate().Wait("2s")
		te.MetaControllerSession = nil
	}
}

func tlsCertToPem(cert *x509.Certificate) ([]byte, error) {
	pemBlock := &pem.Block{
		Type:    "CERTIFICATE",
		Headers: nil,
		Bytes:   cert.Raw,
	}

	buf := new(bytes.Buffer)
	if err := pem.Encode(buf, pemBlock); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (te *TestEnv) kubectl(args ...string) ([]byte, error) {
	cmd := exec.Command("kubectl", args...)
	cmd.Env = []string{
		fmt.Sprintf("KUBECONFIG=%s", te.KubeConfigPath),
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
		fmt.Sprintf("HOME=%s", filepath.Dir(te.KubeConfigPath)),
	}
	fmt.Fprintf(te.TestOutput, "+ kubectl %s\n", strings.Join(args, " "))
	output, err := cmd.CombinedOutput()
	te.TestOutput.Write(output)
	return output, err
}

func (te *TestEnv) kubectlApplyResource(resource string) error {
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Env = []string{
		fmt.Sprintf("KUBECONFIG=%s", te.KubeConfigPath),
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
		fmt.Sprintf("HOME=%s", filepath.Dir(te.KubeConfigPath)),
	}
	cmd.Stdin = bytes.NewReader([]byte(resource))
	fmt.Fprintf(te.TestOutput, "+ kubectl apply with string input\n")
	output, err := cmd.CombinedOutput()
	te.TestOutput.Write(output)
	return err
}

func (te *TestEnv) createKubeConfig(config *rest.Config) error {
	payload := fmt.Sprintf(`apiVersion: v1
clusters:
- cluster:
    server: %s
  name: test-env
contexts:
- context:
    cluster: test-env
    user: test-user
  name: test-env
current-context: test-env
kind: Config
users:
- name: test-user
  user:
    token: %s`, config.Host, config.BearerToken)
	te.KubeConfigPath = filepath.Join(te.TempDir, "kube", "config")
	if err := os.MkdirAll(filepath.Dir(te.KubeConfigPath), 0644); err != nil {
		return err
	}
	fmt.Fprintf(te.TestOutput, "saving kubecfg to %s\n", te.KubeConfigPath)
	return ioutil.WriteFile(te.KubeConfigPath, []byte(payload), 0644)
}

func (te *TestEnv) getResourcesByName(resourceType, namespace string, outMap interface{}) error {
	out, err := te.kubectl("get", resourceType, "-n", namespace, "-o", "json")
	if err != nil {
		return err
	}
	return k8sListResponseByName(out, outMap)
}

func k8sListResponseByName(rawJSON []byte, outMap interface{}) error {
	k8sApiListResponse := reflect.New(reflect.StructOf([]reflect.StructField{
		{
			Name: "Items",
			Type: reflect.SliceOf(reflect.TypeOf(outMap).Elem()),
		},
	})).Interface()
	if err := json.Unmarshal(rawJSON, k8sApiListResponse); err != nil {
		return err
	}
	itemsVal := reflect.ValueOf(k8sApiListResponse).Elem().Field(0)
	outMapVal := reflect.ValueOf(outMap)
	for i := 0; i < itemsVal.Len(); i++ {
		item := itemsVal.Index(i)
		outMapVal.SetMapIndex(item.FieldByName("Name"), item)
	}
	return nil
}
