package integration_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/onsi/gomega/gexec"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/kind/pkg/cluster"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type virtualService struct {
	Spec virtualServiceSpec
}

type virtualServiceSpec struct {
	Gateways []string
	Hosts    []string
	Http     []http
}

type http struct {
	Match []match
	Route []route
}

type match struct {
	Uri uri
}

type uri struct {
	Prefix string
}

type route struct {
	Destination destination
}

type destination struct {
	Host string
}

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

var _ = Describe("Integration", func() {
	var (
		session *gexec.Session

		clusterName    string
		kubeConfigPath string
		namespace      string
		gateway        string
		clientset      kubernetes.Interface

		yamlToApply string

		kubectlGetVirtualServices func() ([]virtualService, error)
		kubectlGetServices        func() ([]service, error)
	)

	BeforeEach(func() {
		clusterName = fmt.Sprintf("test-%d-%d", GinkgoParallelNode(), rand.Uint64())
		namespace = "cf-k8s-networking-tests"
		gateway = "cf-test-gateway"

		kubeConfigPath = createKindCluster(clusterName)
		config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		Expect(err).NotTo(HaveOccurred())

		clientset, err = kubernetes.NewForConfig(config)
		Expect(err).NotTo(HaveOccurred())

		// Create testing namespace
		_, err = clientset.CoreV1().Namespaces().Create(&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		})
		Expect(err).NotTo(HaveOccurred())

		istioCRDPath := filepath.Join("fixtures", "istio-virtual-service.yaml")
		output, err := kubectlWithConfig(kubeConfigPath, nil, "-n", namespace, "apply", "-f", istioCRDPath)
		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("kubectl apply crd failed with err: %s", string(output)))

		// Generate the YAML for the Route CRD with Kustomize, and then apply it with kubectl apply.
		kustomizeOutput, err := kustomizeConfigCRD()
		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("kustomize failed to render CRD yaml: %s", string(kustomizeOutput)))
		kustomizeOutputReader := bytes.NewReader(kustomizeOutput)

		output, err = kubectlWithConfig(kubeConfigPath, kustomizeOutputReader, "-n", namespace, "apply", "-f", "-")
		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("kubectl apply crd failed with err: %s", string(output)))

		session = startRouteController(kubeConfigPath, gateway)

		kubectlGetVirtualServices = func() ([]virtualService, error) {
			output, err := kubectlWithConfig(kubeConfigPath, nil, "-n", namespace, "-o", "json", "get", "virtualservices")
			if err != nil {
				return nil, err
			}

			var resultList struct {
				Items []virtualService
			}
			err = json.Unmarshal(output, &resultList)
			if err != nil {
				return nil, err
			}

			return resultList.Items, nil
		}

		kubectlGetServices = func() ([]service, error) {
			output, err := kubectlWithConfig(kubeConfigPath, nil, "-n", namespace, "-o", "json", "get", "services")
			if err != nil {
				return nil, err
			}

			var resultList struct {
				Items []service
			}
			err = json.Unmarshal(output, &resultList)
			if err != nil {
				return nil, err
			}

			return resultList.Items, nil
		}
	})

	AfterEach(func() {
		session.Interrupt()
		Eventually(session, "10s").Should(gexec.Exit())

		deleteKindCluster(clusterName, kubeConfigPath)
	})

	JustBeforeEach(func() {
		if yamlToApply == "" {
			Fail("yamlToApply must be set by the test")
		}
		output, err := kubectlWithConfig(kubeConfigPath, nil, "-n", namespace, "apply", "-f", yamlToApply)
		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("kubectl apply CR failed with err: %s", string(output)))
	})

	When("there is no destination provided in the route CR", func() {
		BeforeEach(func() {
			yamlToApply = filepath.Join("fixtures", "route-without-any-destination.yaml")
		})

		It("does not create a virtualservice", func() {
			Eventually(kubectlGetVirtualServices).Should(HaveLen(0))
		})
	})

	When("there is a single route CR with a single destination", func() {
		BeforeEach(func() {
			yamlToApply = filepath.Join("fixtures", "single-route-with-single-destination.yaml")
		})

		It("creates a virtualservice and a service", func() {
			By("Verifying the Service")
			Eventually(kubectlGetServices).Should(ConsistOf(
				service{
					Metadata: metadata{
						Name: "s-destination-guid-1",
					},
					Spec: serviceSpec{
						Ports: []serviceSpecPort{
							{
								TargetPort: 8080,
							},
						},
					},
				},
			))

			By("Verifying the VirtualService")
			Eventually(kubectlGetVirtualServices).Should(ConsistOf(
				virtualService{
					Spec: virtualServiceSpec{
						Gateways: []string{gateway},
						Hosts:    []string{"hostname.apps.example.com"},
						Http: []http{
							http{
								Match: []match{
									match{
										Uri: uri{Prefix: "/some/path"},
									},
								},
								Route: []route{
									route{
										Destination: destination{Host: "s-destination-guid-1"},
									},
								},
							},
						},
					},
				},
			))
		})
	})

	When("there are multiple route CRs with the same hostname and domain", func() {
		BeforeEach(func() {
			yamlToApply = filepath.Join("fixtures", "multiple-routes-with-same-fqdn.yaml")
		})

		It("creates a single virtualservice", func() {
			Eventually(kubectlGetVirtualServices).Should(ConsistOf(
				virtualService{
					Spec: virtualServiceSpec{
						Gateways: []string{gateway},
						Hosts:    []string{"hostname.apps.example.com"},
						Http: []http{
							http{
								Match: []match{
									match{
										Uri: uri{Prefix: "/some/path"},
									},
								},
								Route: []route{
									route{
										Destination: destination{Host: "s-destination-guid-1"},
									},
								},
							},
							http{
								Match: []match{
									match{
										Uri: uri{Prefix: "/some/different/path"},
									},
								},
								Route: []route{
									route{
										Destination: destination{Host: "s-destination-guid-2"},
									},
								},
							},
						},
					},
				},
			))
		})
	})

	When("there are multiple route CRs with different hostnames or domains", func() {
		BeforeEach(func() {
			yamlToApply = filepath.Join("fixtures", "multiple-routes-with-different-fqdn.yaml")
		})

		It("creates multiple virtualservice", func() {
			Eventually(kubectlGetVirtualServices).Should(ConsistOf(
				virtualService{
					Spec: virtualServiceSpec{
						Gateways: []string{gateway},
						Hosts:    []string{"hostname-1.apps.example.com"},
						Http: []http{
							http{
								Match: []match{
									match{
										Uri: uri{Prefix: "/some/path"},
									},
								},
								Route: []route{
									route{
										Destination: destination{Host: "s-destination-guid-1"},
									},
								},
							},
						},
					},
				},
				virtualService{
					Spec: virtualServiceSpec{
						Gateways: []string{gateway},
						Hosts:    []string{"hostname-2.apps.example.com"},
						Http: []http{
							http{
								Match: []match{
									match{
										Uri: uri{Prefix: "/some/different/path"},
									},
								},
								Route: []route{
									route{
										Destination: destination{Host: "s-destination-guid-2"},
									},
								},
							},
						},
					},
				},
			))
		})
	})

	When("there is a single route CR with multiple destinations", func() {
		BeforeEach(func() {
			yamlToApply = filepath.Join("fixtures", "single-route-with-multiple-destinations.yaml")
		})

		It("creates a single virtualservice and multiple services", func() {
			Eventually(kubectlGetVirtualServices).Should(ConsistOf(
				virtualService{
					Spec: virtualServiceSpec{
						Gateways: []string{gateway},
						Hosts:    []string{"hostname.apps.example.com"},
						Http: []http{
							http{
								Match: []match{
									match{
										Uri: uri{Prefix: "/some/path"},
									},
								},
								Route: []route{
									route{
										Destination: destination{Host: "s-destination-guid-1"},
									},
									route{
										Destination: destination{Host: "s-destination-guid-2"},
									},
								},
							},
						},
					},
				},
			))

			Eventually(kubectlGetServices).Should(ConsistOf(
				service{
					Metadata: metadata{
						Name: "s-destination-guid-1",
					},
					Spec: serviceSpec{
						Ports: []serviceSpecPort{
							{
								TargetPort: 8080,
							},
						},
					},
				},
				service{
					Metadata: metadata{
						Name: "s-destination-guid-2",
					},
					Spec: serviceSpec{
						Ports: []serviceSpecPort{
							{
								TargetPort: 9000,
							},
						},
					},
				},
			))
		})
	})

	When("Route resources are created in succession for the same FQDN", func() {
		BeforeEach(func() {
			yamlToApply = filepath.Join("fixtures", "context-path-route-for-single-fqdn1.yaml")
		})

		It("updates the virtualservice for that FQDN", func() {
			Eventually(kubectlGetVirtualServices).Should(ConsistOf(
				virtualService{
					Spec: virtualServiceSpec{
						Gateways: []string{gateway},
						Hosts:    []string{"hostname.apps.example.com"},
						Http: []http{
							http{
								Match: []match{
									match{
										Uri: uri{Prefix: "/hello"},
									},
								},
								Route: []route{
									route{
										Destination: destination{Host: "s-destination-guid-1"},
									},
								},
							},
						},
					},
				},
			))

			secondYAMLToApply := filepath.Join("fixtures", "context-path-route-for-single-fqdn2.yaml")
			output, err := kubectlWithConfig(kubeConfigPath, nil, "-n", namespace, "apply", "-f", secondYAMLToApply)
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("kubectl apply CR failed with err: %s", string(output)))

			Eventually(kubectlGetVirtualServices).Should(ConsistOf(
				virtualService{
					Spec: virtualServiceSpec{
						Gateways: []string{gateway},
						Hosts:    []string{"hostname.apps.example.com"},
						Http: []http{
							http{
								Match: []match{
									match{
										Uri: uri{Prefix: "/hello/world"},
									},
								},
								Route: []route{
									route{
										Destination: destination{Host: "s-destination-guid-2"},
									},
								},
							},
							http{
								Match: []match{
									match{
										Uri: uri{Prefix: "/hello"},
									},
								},
								Route: []route{
									route{
										Destination: destination{Host: "s-destination-guid-1"},
									},
								},
							},
						},
					},
				},
			))
		})
	})

	When("Adding a destination to an existing route", func() {
		BeforeEach(func() {
			yamlToApply = filepath.Join("fixtures", "single-route-with-single-destination.yaml")
		})

		It("adds a new service for the new destination, and updates the virtual service with the backend", func() {
			Eventually(kubectlGetServices).Should(ConsistOf(
				service{
					Metadata: metadata{
						Name: "s-destination-guid-1",
					},
					Spec: serviceSpec{
						Ports: []serviceSpecPort{
							{
								TargetPort: 8080,
							},
						},
					},
				},
			))

			Eventually(kubectlGetVirtualServices).Should(ConsistOf(
				virtualService{
					Spec: virtualServiceSpec{
						Gateways: []string{gateway},
						Hosts:    []string{"hostname.apps.example.com"},
						Http: []http{
							http{
								Match: []match{
									match{
										Uri: uri{Prefix: "/some/path"},
									},
								},
								Route: []route{
									route{
										Destination: destination{Host: "s-destination-guid-1"},
									},
								},
							},
						},
					},
				},
			))

			secondYAMLToApply := filepath.Join("fixtures", "single-route-with-multiple-destinations.yaml")
			output, err := kubectlWithConfig(kubeConfigPath, nil, "-n", namespace, "apply", "-f", secondYAMLToApply)
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("kubectl apply CR failed with err: %s", string(output)))

			Eventually(kubectlGetVirtualServices).Should(ConsistOf(
				virtualService{
					Spec: virtualServiceSpec{
						Gateways: []string{gateway},
						Hosts:    []string{"hostname.apps.example.com"},
						Http: []http{
							http{
								Match: []match{
									match{
										Uri: uri{Prefix: "/some/path"},
									},
								},
								Route: []route{
									route{
										Destination: destination{Host: "s-destination-guid-1"},
									},
									route{
										Destination: destination{Host: "s-destination-guid-2"},
									},
								},
							},
						},
					},
				},
			))

			Eventually(kubectlGetServices).Should(ConsistOf(
				service{
					Metadata: metadata{
						Name: "s-destination-guid-1",
					},
					Spec: serviceSpec{
						Ports: []serviceSpecPort{
							{
								TargetPort: 8080,
							},
						},
					},
				},
				service{
					Metadata: metadata{
						Name: "s-destination-guid-2",
					},
					Spec: serviceSpec{
						Ports: []serviceSpecPort{
							{
								TargetPort: 9000,
							},
						},
					},
				},
			))

		})
	})

	When("Removing a destination from an existing route", func() {
		BeforeEach(func() {
			yamlToApply = filepath.Join("fixtures", "single-route-with-multiple-destinations.yaml")
		})

		It("removes the service for the removed destination, and updates the virtual service to remove the backend", func() {

			Eventually(kubectlGetVirtualServices).Should(ConsistOf(
				virtualService{
					Spec: virtualServiceSpec{
						Gateways: []string{gateway},
						Hosts:    []string{"hostname.apps.example.com"},
						Http: []http{
							http{
								Match: []match{
									match{
										Uri: uri{Prefix: "/some/path"},
									},
								},
								Route: []route{
									route{
										Destination: destination{Host: "s-destination-guid-1"},
									},
									route{
										Destination: destination{Host: "s-destination-guid-2"},
									},
								},
							},
						},
					},
				},
			))

			Eventually(kubectlGetServices).Should(ConsistOf(
				service{
					Metadata: metadata{
						Name: "s-destination-guid-1",
					},
					Spec: serviceSpec{
						Ports: []serviceSpecPort{
							{
								TargetPort: 8080,
							},
						},
					},
				},
				service{
					Metadata: metadata{
						Name: "s-destination-guid-2",
					},
					Spec: serviceSpec{
						Ports: []serviceSpecPort{
							{
								TargetPort: 9000,
							},
						},
					},
				},
			))

			secondYAMLToApply := filepath.Join("fixtures", "single-route-with-single-destination.yaml")
			output, err := kubectlWithConfig(kubeConfigPath, nil, "-n", namespace, "apply", "-f", secondYAMLToApply)
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("kubectl apply CR failed with err: %s", string(output)))

			Eventually(kubectlGetServices).Should(ConsistOf(
				service{
					Metadata: metadata{
						Name: "s-destination-guid-1",
					},
					Spec: serviceSpec{
						Ports: []serviceSpecPort{
							{
								TargetPort: 8080,
							},
						},
					},
				},
			))

			Eventually(kubectlGetVirtualServices).Should(ConsistOf(
				virtualService{
					Spec: virtualServiceSpec{
						Gateways: []string{gateway},
						Hosts:    []string{"hostname.apps.example.com"},
						Http: []http{
							http{
								Match: []match{
									match{
										Uri: uri{Prefix: "/some/path"},
									},
								},
								Route: []route{
									route{
										Destination: destination{Host: "s-destination-guid-1"},
									},
								},
							},
						},
					},
				},
			))

		})
	})
})

func startRouteController(kubeConfigPath, gateway string) *gexec.Session {
	cmd := exec.Command(routeControllerBinaryPath)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%s", kubeConfigPath))
	cmd.Env = append(cmd.Env, fmt.Sprintf("ISTIO_GATEWAY_NAME=%s", gateway))

	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	return session
}

func createKindCluster(name string) string {
	provider := cluster.NewProvider()
	err := provider.Create(name)
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
