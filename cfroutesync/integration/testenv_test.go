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
	"regexp"
	"strings"
	"sync"

	"k8s.io/client-go/rest"

	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/ccclient"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/cfg"
	log "github.com/sirupsen/logrus"
	fakeapiserver "sigs.k8s.io/controller-runtime/pkg/envtest"
)

type TestEnv struct {
	lock sync.Mutex

	ConfigDir string

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
	FakeApiServerEnv *fakeapiserver.Environment
	KubeConfigPath   string

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
	configDir, err := ioutil.TempDir("", "cfroutesync-integ-test-config-dir")
	if err != nil {
		return nil, err
	}

	testEnv := &TestEnv{
		ConfigDir:  configDir,
		TestOutput: testOutput,
	}

	testEnv.FakeUAA.Handler = http.HandlerFunc(testEnv.FakeUAAServeHTTP)
	testEnv.FakeUAA.Server = httptest.NewTLSServer(testEnv.FakeUAA.Handler)

	testEnv.FakeCC.Handler = http.HandlerFunc(testEnv.FakeCCServeHTTP)
	testEnv.FakeCC.Server = httptest.NewUnstartedServer(testEnv.FakeCC.Handler)
	// hack: ensure FakeCC uses same server cert as FakeUAA
	testEnv.FakeCC.Server.Config.TLSConfig = testEnv.FakeUAA.Server.TLS
	testEnv.FakeCC.Server.StartTLS()

	fakeUAACertBytes, err := tlsCertToPem(testEnv.FakeUAA.Server.Certificate())
	if err != nil {
		return nil, err
	}

	for filename, contents := range map[string]string{
		cfg.FileUAABaseURL:      testEnv.FakeUAA.Server.URL,
		cfg.FileUAAClientName:   "fake-uaa-client-name",
		cfg.FileUAAClientSecret: "fake-uaa-client-secret",
		cfg.FileUAACA:           string(fakeUAACertBytes),
		cfg.FileCCBaseURL:       testEnv.FakeCC.Server.URL,
		//cfg.FileCCCA:            string(fakeUAACertBytes), // currently same as UAA CA
	} {
		if err := ioutil.WriteFile(filepath.Join(testEnv.ConfigDir, filename), []byte(contents), 0644); err != nil {
			return nil, err
		}
	}

	testEnv.FakeApiServerEnv = &fakeapiserver.Environment{}

	testEnvConfig, err = testEnv.FakeApiServerEnv.Start()
	if err != nil {
		return nil, fmt.Errorf("starting fake api server: %w", err)
	}

	testEnv.KubeConfigPath, err = createKubeConfig(testEnvConfig)
	if err != nil {
		return nil, fmt.Errorf("writing kube config: %w", err)
	}

	return testEnv, nil
}

func (te *TestEnv) Cleanup() {
	if te == nil {
		return
	}
	te.lock.Lock()
	defer te.lock.Unlock()

	if len(te.ConfigDir) > 0 {
		os.RemoveAll(te.ConfigDir)
		te.ConfigDir = ""
	}

	if te.FakeUAA.Server != nil {
		te.FakeUAA.Server.Close()
		te.FakeUAA.Server = nil
	}

	if te.FakeCC.Server != nil {
		te.FakeCC.Server.Close()
		te.FakeCC.Server = nil
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
	}
	cmd.Stderr = te.TestOutput

	fmt.Fprintf(te.TestOutput, "+ kubectl %s\n", strings.Join(args, " "))
	output, err := cmd.Output()
	return output, err
}

func createKubeConfig(config *rest.Config) (string, error) {
	kubeConfig, err := ioutil.TempFile("", "kubeconfig")
	if err != nil {
		return "", err
	}
	defer kubeConfig.Close()

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

	_, err = kubeConfig.Write([]byte(payload))
	if err != nil {
		return "", nil
	}

	return kubeConfig.Name(), nil
}

func createCompositeController(webhookHost string) (string, error) {
	compositeControllerYAML, err := ioutil.TempFile("", "compositecontroller.yaml")
	if err != nil {
		return "", err
	}
	defer compositeControllerYAML.Close()

	payload := fmt.Sprintf(`---
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
        url: http://%s/sync`, webhookHost)

	_, err = compositeControllerYAML.Write([]byte(payload))
	if err != nil {
		return "", nil
	}

	return compositeControllerYAML.Name(), nil
}
