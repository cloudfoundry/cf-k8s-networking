package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/onsi/gomega/gexec"

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

var _ = Describe("Integration with Istio", func() {
	var (
		session *gexec.Session

		clusterName    string
		kubeConfigPath string
		namespace      string
		gateway        string

		yamlToApply string

		kubectlGetVirtualServices func() ([]virtualService, error)
		kubectlGetServices        func() ([]service, error)
	)

	BeforeEach(func() {
		clusterName = fmt.Sprintf("test-%d-%d", GinkgoParallelNode(), rand.Uint64())
		namespace = "cf-k8s-networking-tests"
		gateway = "cf-test-gateway"

		kubeConfigPath = createKindCluster(clusterName)
		output, err := kubectlWithConfig(kubeConfigPath, nil, "create", "namespace", namespace)
		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("kubectl create namespace failed with err: %s", string(output)))

		istioCRDPath := filepath.Join("fixtures", "istio-virtual-service.yaml")
		output, err = kubectlWithConfig(kubeConfigPath, nil, "-n", namespace, "apply", "-f", istioCRDPath)
		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("kubectl apply crd failed with err: %s", string(output)))

		// Generate the YAML for the Route CRD with Kustomize, and then apply it with kubectl apply.
		kustomizeOutput, err := kustomizeConfigCRD()
		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("kustomize failed to render CRD yaml: %s", string(kustomizeOutput)))
		kustomizeOutputReader := bytes.NewReader(kustomizeOutput)

		output, err = kubectlWithConfig(kubeConfigPath, kustomizeOutputReader, "-n", namespace, "apply", "-f", "-")
		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("kubectl apply crd failed with err: %s", string(output)))

		session = startRouteController(kubeConfigPath, gateway, "istio")

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

			By("creating a virtualservice")
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

			yamlToApply = filepath.Join("fixtures", "single-route-with-no-destination.yaml")
			output, err := kubectlWithConfig(kubeConfigPath, nil, "-n", namespace, "apply", "-f", yamlToApply)
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("kubectl apply CR failed with err: %s", string(output)))

			By("removing the destination from the virtualservice")
			Eventually(kubectlGetVirtualServices).Should(ConsistOf(
				virtualService{
					Spec: virtualServiceSpec{
						Gateways: []string{gateway},
						Hosts:    []string{"hostname.apps.example.com"},
						Http: []http{{
							Match: []match{{Uri: uri{Prefix: "/some/path"}}},
							Route: []route{{Destination: destination{Host: "no-destinations"}}},
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
									route{
										Destination: destination{Host: "s-additional-destination-for-route-1"},
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

	When("deleting a route", func() {
		Context("that is the only route for a given domain", func() {
			BeforeEach(func() {
				yamlToApply = filepath.Join("fixtures", "single-route-with-single-destination.yaml")
			})

			It("deletes the associated services and virtual services", func() {
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

				Eventually(kubectlGetVirtualServices, "5s", "1s").Should(BeEmpty())

				Eventually(kubectlGetServices, "5s", "1s").Should(BeEmpty())
			})
		})

		Context("that applies to an FQDN with other routes that are not deleted", func() {
			BeforeEach(func() {
				yamlToApply = filepath.Join("fixtures", "multiple-routes-with-same-fqdn.yaml")
			})

			It("deletes the associated services, and updates the virtual service", func() {
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
											Destination: destination{Host: "s-additional-destination-for-route-1"},
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

				Eventually(kubectlGetVirtualServices).Should(ConsistOf(
					virtualService{
						Spec: virtualServiceSpec{
							Gateways: []string{gateway},
							Hosts:    []string{"hostname.apps.example.com"},
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

		It("removes the Service/updates the VirtualService without deleting Services owned by other routes", func() {
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
										Destination: destination{Host: "s-additional-destination-for-route-1"},
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

	When("a Route's child resources are deleted unexpectedly", func() {
		BeforeEach(func() {
			yamlToApply = filepath.Join("fixtures", "single-route-with-single-destination.yaml")
		})

		It("recreates the VirtualService and Service for the Route", func() {
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

			output, err := kubectlWithConfig(kubeConfigPath, nil, "-n", namespace, "delete", "services", "--all", "--wait=true")
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("kubectl delete services failed with err: %s", string(output)))

			output, err = kubectlWithConfig(kubeConfigPath, nil, "-n", namespace, "delete", "virtualservices", "--all", "--wait=true")
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("kubectl delete virtualservices failed with err: %s", string(output)))

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

	Describe("kubectl", func() {
		type routeView struct {
			name string
			url  string
			age  string
		}

		When("viewing routes with -owide view mode", func() {
			BeforeEach(func() {
				yamlToApply = filepath.Join("fixtures", "multiple-routes-with-different-fqdn.yaml")
			})

			It("outputs the associated name and URL", func() {
				Eventually(func() ([]routeView, error) {
					output, err := kubectlWithConfig(kubeConfigPath, nil, "-n", namespace, "get", "routes", "-o", "wide")
					if err != nil {
						return nil, err
					}

					// the first line is a header with column names
					lines := strings.Split(strings.TrimSpace(string(output)), "\n")[1:]
					Expect(lines).Should(HaveLen(2))
					routes := make([]routeView, 0, len(lines))

					const (
						nameColumn = 0
						urlColumn  = 1
						ageColumn  = 2
					)

					spaceRe := regexp.MustCompile(`\s+`)
					for _, line := range lines {
						columns := spaceRe.Split(line, -1)
						Expect(columns).Should(HaveLen(3))

						Expect(columns[nameColumn]).ShouldNot(BeEmpty())
						Expect(columns[urlColumn]).ShouldNot(BeEmpty())
						Expect(columns[ageColumn]).ShouldNot(BeEmpty())

						routes = append(routes, routeView{
							name: columns[nameColumn],
							url:  columns[urlColumn],
							// no assertion for age column to prevent flakes
						})
					}

					return routes, nil
				}).Should(ConsistOf(
					routeView{
						name: "cc-route-guid-1",
						url:  "hostname-1.apps.example.com/some/path",
					},
					routeView{
						name: "cc-route-guid-2",
						url:  "hostname-2.apps.example.com/some/different/path",
					},
				))
			})
		})
	})
})
