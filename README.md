cf-k8s-networking
---
Routing and networking for Cloud Foundry running on Kubernetes.

## Deploying

CF-K8s-Networking is a component of CF-for-K8s. To deploy CF-for-K8s reference
the following documentation:

* [Deploy Cloud Foundry on
  Kubernetes](https://github.com/cloudfoundry/cf-for-k8s/blob/master/docs/deploy.md)
* [Deploy Cloud Foundry
  Locally](https://github.com/cloudfoundry/cf-for-k8s/blob/6e4ba5cc0514481a0675ea83731449c752b1dcad/docs/deploy-local.md)

## Architecture

![Architecture Diagram of
CF-K8s-Networking](doc/assets/routecontroller-data-flow-diagram.png)

* **RouteController:** Watches the Kubernetes API for Route CRs and translates
  the Route CRs into Istio Virtual Service CRs and Kubernetes Services
  accordingly to enable routing to applications deployed by Cloud Foundry.

* **Istio:** CF-K8s-Networking currently depends on [Istio](https://istio.io/).
  * Istio serves as both our gateway router for ingress networking, replacing
    the role of the Gorouters in CF for VMs, and service mesh for (eventually)
    container-to-container networking policy enforcement.
  * We provide a manifest for installing our custom configuration for Istio,
    [here](config/istio/generated/xxx-generated-istio.yaml).
  * Istio provides us with security features out of the box, such as:
    * Automatic Envoy sidecar injection for system components and application workloads
    * `Sidecar` Kubernetes resources that can limit egress traffic from workload `Pod`s
    * Transparent mutal TLS (mTLS) everywhere
    * (Eventually) app identity certificates using [SPIFFE](https://spiffe.io/) issued by Istio Citadel
  * Istio should be treated as an "implementation detail" of the platform and
    our reliance on it is subject to change

## Contributing
For information about how to contribute, develop against our codebase, and run
our various test suites, check out our [Contributing guidelines](CONTRIBUTING.md).

