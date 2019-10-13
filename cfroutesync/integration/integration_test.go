package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"

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
		te, err = NewTestEnv()
		Expect(err).NotTo(HaveOccurred())

		webhookListenAddr = fmt.Sprintf("127.0.0.1:%d", ports.PickAPort())
	})

	AfterEach(func() {
		te.Cleanup()
	})

	metacontrollerSync := func(req webhook.SyncRequest) (int, *webhook.SyncResponse, error) {
		reqBody := bytes.NewBuffer(nil)
		json.NewEncoder(reqBody).Encode(req)
		resp, err := http.Post(fmt.Sprintf("http://%s/sync", webhookListenAddr), "application/json", reqBody)
		if err != nil {
			return 0, nil, err
		}

		var syncResp webhook.SyncResponse
		err = json.NewDecoder(resp.Body).Decode(&syncResp)
		return resp.StatusCode, &syncResp, err
	}

	Specify("cfroutesync boots and stays running", func() {
		cmd := exec.Command(binaryPathCFRouteSync, "-c", te.ConfigDir, "-l", webhookListenAddr, "-v", "6")
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		defer func() { session.Terminate().Wait("2s") }()

		Eventually(session.Out).Should(gbytes.Say("starting webhook server"))
		Eventually(session.Out).Should(gbytes.Say("starting cc fetch loop"))

		syncReq := webhook.SyncRequest{
			Parent: webhook.BulkSync{
				Spec: webhook.BulkSyncSpec{},
			},
		}
		Eventually(func() error {
			statusCode, _, err := metacontrollerSync(syncReq)
			if err != nil {
				return err
			}
			if statusCode != http.StatusOK {
				return fmt.Errorf("unexpected status code %d", statusCode)
			}
			return nil
		}).Should(Succeed())
	})
})
