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
CF-K8s-Networking](doc/assets/routecontroller-design.png)

* **RouteController:** Watches the Kubernetes API for Route CRs and translates
  the Route CRs into Istio Virtual Service CRs and Kubernetes Services
  accordingly to enable routing to applications deployed by Cloud Foundry.
  * TODO: Explain the relationships between what happens in CAPI to Route CR and
    Route CR to Virtual Service and K8s Service
  * TODO: Add explanation of how Eirini creates stateful sets and how we
    reference the pods of those stateful sets in routecontroller

* **Istio:** TODO: Explain that Istio is a dependency of the cf-k8s-networking
  subsystem
  * Explain that Sidecars are used and required by the cf-k8s-networking
    subsystem
    * Explain that it is required because of mTLS

* TODO: Explain other default networking configuration (namespace network
  policy?)

* TODO: Explain that prometheus thing (?)

## Testing

* TODO: Add references to readmes on all our various tests and how to run them

## Contributing
Check out our [Contributing guidelines](CONTRIBUTING.md).
