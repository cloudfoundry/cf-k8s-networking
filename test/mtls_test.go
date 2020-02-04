package acceptance_test

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const workloadsNamespace = "cf-workloads"
const systemNamespace = "cf-system"

const CurlSuccessfulExitCode = 0

var _ = Describe("mTLS setup on a CF-k8s env", func() {
	const cfAppContainerName = "opi"
	const proxyContainerName = "istio-proxy"
	const systemComponentContainerName = "system-component"
	const systemComponentServiceAccountName = "system-component-sa"

	Context("when auto mTLS is enabled and the MeshPolicy is STRICT", func() {
		Describe("for requests from app pod to system component pod", func() {
			var (
				appPodName       string
				sysComponentAddr string
				appPodSelector   string
			)

			BeforeEach(func() {
				var err error
				appPodSelector = "cloudfoundry.org/guid=" + AppGuid
				appPodName, err = getPodNameBySelector(workloadsNamespace, appPodSelector)
				Expect(err).NotTo(HaveOccurred())
				sysComponentAddr, err = getSvcHTTPAddrBySelector(systemNamespace, SysComponentSelector)
				Expect(err).NotTo(HaveOccurred())
			})

			Describe("when sending request from the app container to a system component", func() {
				It("successfully establishes connection with the system component over mTLS", func() {
					By("checking that the request headers on receiving side contains the SVID for the application")
					output, exitCode, _, err := curlInPod(workloadsNamespace, appPodName, cfAppContainerName, fmt.Sprintf("http://%s/headers", sysComponentAddr))
					Expect(err).NotTo(HaveOccurred())
					Expect(exitCode).To(Equal(CurlSuccessfulExitCode))

					svid := parseSVID(output)
					Expect(svid).To(Equal("URI=spiffe://cluster.local/ns/" + workloadsNamespace + "/sa/default"))
				})
			})

			Describe("when sending request from the proxy container in the app pod to a system component", func() {
				Describe("over HTTP", func() {
					It("cannot establish connection with the system component", func() {
						_, exitCode, _, err := curlInPod(workloadsNamespace, appPodName, proxyContainerName, fmt.Sprintf("http://%s/headers", sysComponentAddr))
						Expect(err).NotTo(HaveOccurred())
						Expect(exitCode).NotTo(Equal(CurlSuccessfulExitCode))
					})
				})

				Describe("over HTTPS without client credentials", func() {
					It("cannot establish connection with the system component", func() {
						_, exitCode, _, err := curlInPod(workloadsNamespace, appPodName, proxyContainerName, fmt.Sprintf("https://%s/headers", sysComponentAddr), "-k")
						Expect(err).NotTo(HaveOccurred())
						Expect(exitCode).NotTo(Equal(CurlSuccessfulExitCode))
					})
				})

				Describe("over HTTPS with client credentials", func() {
					It("successfully establishes connection with the system component over mTLS", func() {
						By("checking that the request headers on receiving side contains the SVID for the application")
						output, exitCode, _, err := curlInPod(workloadsNamespace, appPodName, proxyContainerName, fmt.Sprintf("https://%s/headers", sysComponentAddr), "-k", "--cacert", "/etc/certs/root-cert.pem", "--key", "/etc/certs/key.pem", "--cert", "/etc/certs/cert-chain.pem")
						Expect(err).NotTo(HaveOccurred())
						Expect(exitCode).To(Equal(CurlSuccessfulExitCode))

						svid := parseSVID(output)
						Expect(svid).To(Equal("URI=spiffe://cluster.local/ns/" + workloadsNamespace + "/sa/default"))
					})
				})
			})

		})

		Describe("for requests from system component pod to app pod", func() {
			var (
				systemComponentPod string
				appAddr            string
				appSvcSelector     string
			)

			BeforeEach(func() {
				var err error
				appSvcSelector = "cloudfoundry.org/app=" + AppGuid
				systemComponentPod, err = getPodNameBySelector(systemNamespace, SysComponentSelector)
				Expect(err).NotTo(HaveOccurred())
				appAddr, err = getSvcHTTPAddrBySelector(workloadsNamespace, appSvcSelector)
				Expect(err).NotTo(HaveOccurred())
			})

			Describe("when sending request from the system component container to the app", func() {
				It("successfully establishes connection with the system component over mTLS", func() {
					By("checking that the request headers on receiving side contains the SVID for the system component")
					output, exitCode, _, err := curlInPod(systemNamespace, systemComponentPod, systemComponentContainerName, fmt.Sprintf("http://%s/headers", appAddr))
					Expect(err).NotTo(HaveOccurred())
					Expect(exitCode).To(Equal(CurlSuccessfulExitCode))

					svid := parseSVID(output)
					Expect(svid).To(Equal("URI=spiffe://cluster.local/ns/" + systemNamespace + "/sa/" + systemComponentServiceAccountName))
				})
			})

			Describe("when sending request from the proxy container in the pod to a system component", func() {
				Describe("over HTTP", func() {
					It("cannot establish connection with the system component", func() {
						_, exitCode, _, err := curlInPod(systemNamespace, systemComponentPod, proxyContainerName, fmt.Sprintf("http://%s/headers", appAddr))
						Expect(err).NotTo(HaveOccurred())
						Expect(exitCode).NotTo(Equal(CurlSuccessfulExitCode))
					})
				})

				Describe("over HTTPS without client credentials", func() {
					It("cannot establish connection with the system component", func() {
						_, exitCode, _, err := curlInPod(systemNamespace, systemComponentPod, proxyContainerName, fmt.Sprintf("https://%s/headers", appAddr), "-k")
						Expect(err).NotTo(HaveOccurred())
						Expect(exitCode).NotTo(Equal(CurlSuccessfulExitCode))
					})
				})

				Describe("over HTTPS with client credentials", func() {
					It("successfully establishes connection with the system component over mTLS", func() {
						By("checking that the request headers on receiving side contains the SVID for the application")
						output, exitCode, _, err := curlInPod(systemNamespace, systemComponentPod, proxyContainerName, fmt.Sprintf("https://%s/headers", appAddr), "-k", "--cacert", "/etc/certs/root-cert.pem", "--key", "/etc/certs/key.pem", "--cert", "/etc/certs/cert-chain.pem")
						Expect(err).NotTo(HaveOccurred())
						Expect(exitCode).To(Equal(CurlSuccessfulExitCode))

						svid := parseSVID(output)
						Expect(svid).To(Equal("URI=spiffe://cluster.local/ns/" + systemNamespace + "/sa/" + systemComponentServiceAccountName))
					})
				})
			})

		})

	})
})

func parseSVID(headers string) string {
	re := regexp.MustCompile("URI\\=spiffe.*sa/[a-z-]*|response.*")
	return re.FindString(headers)
}

func curlInPod(namespace string, podName string, containerName string, url string, args ...string) (output string, exitCode int, responseCode int, err error) {
	curlCommand := "curl --silent " + url + " --write-out \"response_code:%{http_code}\\n\" " + strings.Join(args, " ")

	output, exitCode, err = execInPod(namespace, podName, containerName, curlCommand)
	if err != nil {
		return "", 0, 0, nil
	}

	lines := strings.Split(output, "\n")

	if len(lines) != 0 {
		// last line should be "response_code:NUM"
		t := strings.SplitN(lines[len(lines)-1], ":", 2)
		if len(t) == 2 {
			responseCode, _ = strconv.Atoi(t[1])
		}

		lines = lines[0 : len(lines)-1]
		output = strings.Join(lines, "\n")

		fmt.Println("=========OUTPUT IS============")
		fmt.Println(output)
		fmt.Println("==============================")
	}

	return output, exitCode, responseCode, nil
}

func execInPod(namespace string, podName string, containerName string, command string) (string, int, error) {
	stdout, err := kubectl.Run("-n", namespace, "exec", podName, "-c", containerName, "--", "bash", "-c", command)
	output := strings.TrimSpace(string(stdout))

	fmt.Println("=========ERROR IS============")
	fmt.Printf("%+v\n", err)
	fmt.Println("=============================")

	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return output, ee.ExitCode(), nil
		}

		return "", 0, err
	}

	return output, 0, nil
}

func getSvcHTTPAddrBySelector(namespace string, selector string) (string, error) {
	output, err := kubectl.Run("-n", namespace, "get", "svc", "-l", selector, fmt.Sprintf(
		"-ojsonpath='%s.%s.svc.cluster.local:%s'",
		"{.items[0].metadata.name}", // name path
		namespace,
		"{.items[0].spec.ports[?(@.name==\"http\")].port}", // http port path
	))
	if err != nil {
		return "", err
	}

	return strings.Trim(string(output), "'"), nil
}

func getPodNameBySelector(namespace string, selector string) (string, error) {
	output, err := kubectl.Run("-n", namespace, "get", "pods", "-l", selector, "-ojsonpath='{.items[0].metadata.name}'")
	if err != nil {
		return "", err
	}

	return strings.Trim(string(output), "'"), nil
}
