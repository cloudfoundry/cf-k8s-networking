package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"path/filepath"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type httpProxy struct {
	Spec httpProxySpec
}

type httpProxySpec struct {
	VirtualHost virtualHost
	Routes     []hpRoute
}

type virtualHost struct {
	Fqdn string
}

type hpRoute struct {
	Services   []hpService
	Conditions []matchCondition
}

type hpService struct {
	Name string
	Port int64
}

type matchCondition struct {
	Prefix string
}

var _ = Describe("Integration with Contour", func() {
	var (
		session *gexec.Session

		clusterName    string
		kubeConfigPath string
		namespace      string
		gateway        string

		yamlToApply string

		kubectlGetHTTPProxies func() ([]httpProxy, error)
		kubectlGetServices    func() ([]service, error)
	)

	BeforeEach(func() {
		clusterName = fmt.Sprintf("test-%d-%d", GinkgoParallelNode(), rand.Uint64())
		namespace = "cf-k8s-networking-tests"
		gateway = "cf-test-gateway"

		kubeConfigPath = createKindCluster(clusterName)
		output, err := kubectlWithConfig(kubeConfigPath, nil, "create", "namespace", namespace)
		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("kubectl create namespace failed with err: %s", string(output)))

		contourCRDPath := filepath.Join("fixtures", "contour-httpproxy-crd.yaml")
		output, err = kubectlWithConfig(kubeConfigPath, nil, "-n", namespace, "apply", "-f", contourCRDPath)
		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("kubectl apply crd failed with err: %s", string(output)))

		// Generate the YAML for the Route CRD with Kustomize, and then apply it with kubectl apply.
		kustomizeOutput, err := kustomizeConfigCRD()
		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("kustomize failed to render CRD yaml: %s", string(kustomizeOutput)))
		kustomizeOutputReader := bytes.NewReader(kustomizeOutput)

		output, err = kubectlWithConfig(kubeConfigPath, kustomizeOutputReader, "-n", namespace, "apply", "-f", "-")
		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("kubectl apply crd failed with err: %s", string(output)))

		session = startRouteController(kubeConfigPath, gateway, "contour")

		kubectlGetHTTPProxies = func() ([]httpProxy, error) {
			output, err := kubectlWithConfig(kubeConfigPath, nil, "-n", namespace, "-o", "json", "get", "httpproxies")
			if err != nil {
				return nil, err
			}

			var resultList struct {
				Items []httpProxy
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

		It("does not create an HTTPProxy", func() {
			Eventually(kubectlGetHTTPProxies).Should(HaveLen(0))
		})
	})

	When("there is a single route CR with a single destination", func() {
		BeforeEach(func() {
			yamlToApply = filepath.Join("fixtures", "single-route-with-single-destination.yaml")
		})

		It("handles the route", func() {
			By("creating a service")
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

			By("creating an HTTPProxy")
			Eventually(kubectlGetHTTPProxies).Should(ConsistOf(
				httpProxy{
					Spec: httpProxySpec{
						VirtualHost: virtualHost{Fqdn: "hostname.apps.example.com"},
						Routes: []hpRoute{
							{
								Conditions: []matchCondition{
									{
										Prefix: "/some/path",
									},
								},
								Services: []hpService{
									{
										Name: "s-destination-guid-1",
										Port: 8080,
									},
								},
							},
						},
					},
				},
			))
		})

		It("handles removing a destination from the route correctly", func() {
			// check that service and vs exists
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

			Eventually(kubectlGetHTTPProxies).Should(ConsistOf(
				httpProxy{
					Spec: httpProxySpec{
						VirtualHost: virtualHost{Fqdn: "hostname.apps.example.com"},
						Routes: []hpRoute{
							{
								Conditions: []matchCondition{
									{
										Prefix: "/some/path",
									},
								},
								Services: []hpService{
									{
										Name: "s-destination-guid-1",
										Port: 8080,
									},
								},
							},
						},
					},
				},
			))

			yamlToApply = filepath.Join("fixtures", "single-route-with-no-destination.yaml")
			output, err := kubectlWithConfig(kubeConfigPath, nil, "-n", namespace, "apply", "-f", yamlToApply)
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("kubectl apply CR failed with err: %s", string(output)))

			By("removing the destination from the httpproxy")
			Eventually(kubectlGetHTTPProxies).Should(ConsistOf(
				httpProxy{
					Spec: httpProxySpec{
						VirtualHost: virtualHost{Fqdn: "hostname.apps.example.com"},
						Routes: []hpRoute{{
							Conditions: []matchCondition{{Prefix: "/some/path"}},
							Services:   []hpService{{Name: "no-destinations", Port: 8080}},
						}},
					},
				},
			))

			By("deleting the service associated with the destination")
			resultList, err := kubectlGetServices()
			Expect(err).NotTo(HaveOccurred())
			Eventually(len(resultList)).Should(Equal(0))
		})
	})

	When("there are multiple route CRs with the same hostname and domain", func() {
		BeforeEach(func() {
			yamlToApply = filepath.Join("fixtures", "multiple-routes-with-same-fqdn.yaml")
		})

		It("creates a single httpproxy", func() {
			Eventually(kubectlGetHTTPProxies).Should(ConsistOf(
				httpProxy{
					Spec: httpProxySpec{
						VirtualHost: virtualHost{Fqdn: "hostname.apps.example.com"},
						Routes: []hpRoute{
							{
								Conditions: []matchCondition{
									{
										Prefix: "/some/path",
									},
								},
								Services: []hpService{
									{
										Name: "s-destination-guid-1",
										Port: 8080,
									},
									{
										Name: "s-additional-destination-for-route-1",
										Port: 8080,
									},
								},
							},
							{
								Conditions: []matchCondition{
									{
										Prefix: "/some/different/path",
									},
								},
								Services: []hpService{
									{
										Name: "s-destination-guid-2",
										Port: 8080,
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

		It("creates multiple httpproxy", func() {
			Eventually(kubectlGetHTTPProxies).Should(ConsistOf(
				httpProxy{
					Spec: httpProxySpec{
						VirtualHost: virtualHost{Fqdn: "hostname-1.apps.example.com"},
						Routes: []hpRoute{
							{
								Conditions: []matchCondition{
									{
										Prefix: "/some/path",
									},
								},
								Services: []hpService{
									{
										Name: "s-destination-guid-1",
										Port: 8080,
									},
								},
							},
						},
					},
				},
				httpProxy{
					Spec: httpProxySpec{
						VirtualHost: virtualHost{Fqdn: "hostname-2.apps.example.com"},
						Routes: []hpRoute{
							{
								Conditions: []matchCondition{
									{
										Prefix: "/some/different/path",
									},
								},
								Services: []hpService{
									{
										Name: "s-destination-guid-2",
										Port: 8080,
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

		It("creates a single httpproxy and multiple services", func() {
			Eventually(kubectlGetHTTPProxies).Should(ConsistOf(
				httpProxy{
					Spec: httpProxySpec{
						VirtualHost: virtualHost{Fqdn: "hostname.apps.example.com"},
						Routes: []hpRoute{
							{
								Conditions: []matchCondition{
									{
										Prefix: "/some/path",
									},
								},
								Services: []hpService{
									{
										Name: "s-destination-guid-1",
										Port: 8080,
									},
									{
										Name: "s-destination-guid-2",
										Port: 8080,
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

		It("updates the httpproxy for that FQDN", func() {
			Eventually(kubectlGetHTTPProxies).Should(ConsistOf(
				httpProxy{
					Spec: httpProxySpec{
						VirtualHost: virtualHost{Fqdn: "hostname.apps.example.com"},
						Routes: []hpRoute{
							{
								Conditions: []matchCondition{
									{
										Prefix: "/hello",
									},
								},
								Services: []hpService{
									{
										Name: "s-destination-guid-1",
										Port: 8080,
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

			Eventually(kubectlGetHTTPProxies).Should(ConsistOf(
				httpProxy{
					Spec: httpProxySpec{
						VirtualHost: virtualHost{Fqdn: "hostname.apps.example.com"},
						Routes: []hpRoute{
							{
								Conditions: []matchCondition{
									{
										Prefix: "/hello/world",
									},
								},
								Services: []hpService{
									{
										Name: "s-destination-guid-2",
										Port: 8080,
									},
								},
							},
							{
								Conditions: []matchCondition{
									{
										Prefix: "/hello",
									},
								},
								Services: []hpService{
									{
										Name: "s-destination-guid-1",
										Port: 8080,
									},
								},
							},
						},
					},
				},
			))
		})
	})

	Context("For an existing route with a single destination", func() {
		BeforeEach(func() {
			yamlToApply = filepath.Join("fixtures", "single-route-with-single-destination.yaml")
		})

		When("adding an additional destination to the Route", func() {
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

				Eventually(kubectlGetHTTPProxies).Should(ConsistOf(
					httpProxy{
						Spec: httpProxySpec{
							VirtualHost: virtualHost{Fqdn: "hostname.apps.example.com"},
							Routes: []hpRoute{
								{
									Conditions: []matchCondition{
										{
											Prefix: "/some/path",
										},
									},
									Services: []hpService{
										{
											Name: "s-destination-guid-1",
											Port: 8080,
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

				Eventually(kubectlGetHTTPProxies).Should(ConsistOf(
					httpProxy{
						Spec: httpProxySpec{
							VirtualHost: virtualHost{Fqdn: "hostname.apps.example.com"},
							Routes: []hpRoute{
								{
									Conditions: []matchCondition{
										{
											Prefix: "/some/path",
										},
									},
									Services: []hpService{
										{
											Name: "s-destination-guid-1",
											Port: 8080,
										},
										{
											Name: "s-destination-guid-2",
											Port: 8080,
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

		When("changing the destination on the Route", func() {
			It("updates the service and virtual service", func() {
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

				Eventually(kubectlGetHTTPProxies).Should(ConsistOf(
					httpProxy{
						Spec: httpProxySpec{
							VirtualHost: virtualHost{Fqdn: "hostname.apps.example.com"},
							Routes: []hpRoute{
								{
									Conditions: []matchCondition{
										{
											Prefix: "/some/path",
										},
									},
									Services: []hpService{
										{
											Name: "s-destination-guid-1",
											Port: 8080,
										},
									},
								},
							},
						},
					},
				))

				secondYAMLToApply := filepath.Join("fixtures", "single-route-with-updated-single-destination.yaml")
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
									TargetPort: 9090,
								},
							},
						},
					},
				))

				Eventually(kubectlGetHTTPProxies).Should(ConsistOf(
					httpProxy{
						Spec: httpProxySpec{
							VirtualHost: virtualHost{Fqdn: "hostname.apps.example.com"},
							Routes: []hpRoute{
								{
									Conditions: []matchCondition{
										{
											Prefix: "/some/path",
										},
									},
									Services: []hpService{
										{
											Name: "s-destination-guid-1",
											Port: 8080,
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

	When("deleting a route", func() {
		Context("that is the only route for a given domain", func() {
			BeforeEach(func() {
				yamlToApply = filepath.Join("fixtures", "single-route-with-single-destination.yaml")
			})

			It("deletes the associated services and virtual services", func() {
				Eventually(kubectlGetHTTPProxies).Should(ConsistOf(
					httpProxy{
						Spec: httpProxySpec{
							VirtualHost: virtualHost{Fqdn: "hostname.apps.example.com"},
							Routes: []hpRoute{
								{
									Conditions: []matchCondition{
										{
											Prefix: "/some/path",
										},
									},
									Services: []hpService{
										{
											Name: "s-destination-guid-1",
											Port: 8080,
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
				))

				output, err := kubectlWithConfig(kubeConfigPath, nil, "-n", namespace, "delete", "routes", "cc-route-guid-1")
				Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("kubectl delete route CR failed with err: %s", string(output)))

				Eventually(kubectlGetHTTPProxies, "5s", "1s").Should(BeEmpty())

				Eventually(kubectlGetServices, "5s", "1s").Should(BeEmpty())
			})
		})

		Context("that applies to an FQDN with other routes that are not deleted", func() {
			BeforeEach(func() {
				yamlToApply = filepath.Join("fixtures", "multiple-routes-with-same-fqdn.yaml")
			})

			It("deletes the associated services, and updates the virtual service", func() {
				Eventually(kubectlGetHTTPProxies).Should(ConsistOf(
					httpProxy{
						Spec: httpProxySpec{
							VirtualHost: virtualHost{Fqdn: "hostname.apps.example.com"},
							Routes: []hpRoute{
								{
									Conditions: []matchCondition{
										{
											Prefix: "/some/path",
										},
									},
									Services: []hpService{
										{
											Name: "s-destination-guid-1",
											Port: 8080,
										},
										{
											Name: "s-additional-destination-for-route-1",
											Port: 8080,
										},
									},
								},
								{
									Conditions: []matchCondition{
										{
											Prefix: "/some/different/path",
										},
									},
									Services: []hpService{
										{
											Name: "s-destination-guid-2",
											Port: 8080,
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
							Name: "s-additional-destination-for-route-1",
						},
						Spec: serviceSpec{
							Ports: []serviceSpecPort{
								{
									TargetPort: 9090,
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
									TargetPort: 8080,
								},
							},
						},
					},
				))

				output, err := kubectlWithConfig(kubeConfigPath, nil, "-n", namespace, "delete", "routes", "cc-route-guid-1")
				Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("kubectl delete route CR failed with err: %s", string(output)))

				Eventually(kubectlGetHTTPProxies).Should(ConsistOf(
					httpProxy{
						Spec: httpProxySpec{
							VirtualHost: virtualHost{Fqdn: "hostname.apps.example.com"},
							Routes: []hpRoute{
								{
									Conditions: []matchCondition{
										{
											Prefix: "/some/different/path",
										},
									},
									Services: []hpService{
										{
											Name: "s-destination-guid-2",
											Port: 8080,
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
							Name: "s-destination-guid-2",
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
			})
		})
	})

	When("removing a destination from an existing route", func() {
		BeforeEach(func() {
			yamlToApply = filepath.Join("fixtures", "multiple-routes-with-same-fqdn.yaml")
		})

		It("removes the Service/updates the httpproxy without deleting Services owned by other routes", func() {
			Eventually(kubectlGetHTTPProxies).Should(ConsistOf(
				httpProxy{
					Spec: httpProxySpec{
						VirtualHost: virtualHost{Fqdn: "hostname.apps.example.com"},
						Routes: []hpRoute{
							{
								Conditions: []matchCondition{
									{
										Prefix: "/some/path",
									},
								},
								Services: []hpService{
									{
										Name: "s-destination-guid-1",
										Port: 8080,
									},
									{
										Name: "s-additional-destination-for-route-1",
										Port: 8080,
									},
								},
							},
							{
								Conditions: []matchCondition{
									{
										Prefix: "/some/different/path",
									},
								},
								Services: []hpService{
									{
										Name: "s-destination-guid-2",
										Port: 8080,
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
						Name: "s-additional-destination-for-route-1",
					},
					Spec: serviceSpec{
						Ports: []serviceSpecPort{
							{
								TargetPort: 9090,
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
								TargetPort: 8080,
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
				service{
					Metadata: metadata{
						Name: "s-destination-guid-2",
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

			Eventually(kubectlGetHTTPProxies).Should(ConsistOf(
				httpProxy{
					Spec: httpProxySpec{
						VirtualHost: virtualHost{Fqdn: "hostname.apps.example.com"},
						Routes: []hpRoute{
							{
								Conditions: []matchCondition{
									{
										Prefix: "/some/path",
									},
								},
								Services: []hpService{
									{
										Name: "s-destination-guid-1",
										Port: 8080,
									},
								},
							},
							{
								Conditions: []matchCondition{
									{
										Prefix: "/some/different/path",
									},
								},
								Services: []hpService{
									{
										Name: "s-destination-guid-2",
										Port: 8080,
									},
								},
							},
						},
					},
				},
			))
		})
	})

	When("a Route's child resources are deleted unexpectedly", func() {
		BeforeEach(func() {
			yamlToApply = filepath.Join("fixtures", "single-route-with-single-destination.yaml")
		})

		It("recreates the httpproxy and Service for the Route", func() {
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

			Eventually(kubectlGetHTTPProxies).Should(ConsistOf(
				httpProxy{
					Spec: httpProxySpec{
						VirtualHost: virtualHost{Fqdn: "hostname.apps.example.com"},
						Routes: []hpRoute{
							{
								Conditions: []matchCondition{
									{
										Prefix: "/some/path",
									},
								},
								Services: []hpService{
									{
										Name: "s-destination-guid-1",
										Port: 8080,
									},
								},
							},
						},
					},
				}))

			output, err := kubectlWithConfig(kubeConfigPath, nil, "-n", namespace, "delete", "services", "--all", "--wait=true")
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("kubectl delete services failed with err: %s", string(output)))

			output, err = kubectlWithConfig(kubeConfigPath, nil, "-n", namespace, "delete", "httpproxies", "--all", "--wait=true")
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("kubectl delete httpproxies failed with err: %s", string(output)))

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

			Eventually(kubectlGetHTTPProxies).Should(ConsistOf(
				httpProxy{
					Spec: httpProxySpec{
						VirtualHost: virtualHost{Fqdn: "hostname.apps.example.com"},
						Routes: []hpRoute{
							{
								Conditions: []matchCondition{
									{
										Prefix: "/some/path",
									},
								},
								Services: []hpService{
									{
										Name: "s-destination-guid-1",
										Port: 8080,
									},
								},
							},
						},
					},
				}))
		})
	})
})
