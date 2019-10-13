package integration_test

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/ccclient"
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/cfg"
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
			"resources": te.FakeCC.Data.Destinations[routeGUIDs[1]],
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

func NewTestEnv() (*TestEnv, error) {
	configDir, err := ioutil.TempDir("", "cfroutesync-integ-test-config-dir")
	if err != nil {
		return nil, err
	}

	testEnv := &TestEnv{ConfigDir: configDir}

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

	return testEnv, nil
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
