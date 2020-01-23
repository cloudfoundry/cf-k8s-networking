# 4. Strategy for Securing Network Traffic

Date: 2020-01-23

## Status

Proposal

## Context

We have a goal to have all egress/ingress network traffic within the platform to be encrypted. 
We describe the different situations for source and destination and how we can accomplish this goal.

## Decision

We've identified three categories of sources/destinations:

* **External:** Traffic originating from/destined outside of the platform.
For example, the `cf cli` to Cloud Controller, web browsers to apps, UAA to off platform identity providers, etc.
* **Apps:** Unprivileged `cf push`-ed app workloads running on the platform.
For example, [dora](https://github.com/cloudfoundry/cf-acceptance-tests/tree/master/assets/dora).
* **System Components:** Cloud Foundry core services that implement power the platform.
For example, UAA, Cloud Controller, Log-Cache, etc.

We expect all of these components to want to communicate with each other in some fashion and we will leverage some existing Istio functionality to accomplish securing internal traffic.

See the following matrix:

| Source\Destination | External                                                                                                                      | Apps                                                                                                                                    | System Components                                                                                                                                                                                                   |
|--------------------|-------------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| External           | N/A                                                                                                                           | Operator provides wildcard certs to Ingress Gateway. Gateway to backend is secured using [Istio Mesh mTLS](#istio-mesh-mtls) | Operator provides wildcard certs for the system domain to Ingress Gateway. Gateway to backend is secured using [Istio Mesh mTLS](#istio-mesh-mtls)                                                                                 |
| Apps               | Mesh can be leveraged with Istio Destination Rules and Service Entries. [Istio defaults to PERMISSIVE egress](https://istio.io/docs/tasks/traffic-management/egress/egress-control/).                                                       | App-to-app traffic will be denied by default. If enabled, will be secured using [Istio Mesh mTLS](#istio-mesh-mtls) | [Istio Mesh mTLS](#istio-mesh-mtls)<sup>0</sup>                                                                                                                     |
| System Components  | CAs for external destinations can be provided<sup>1</sup> | [Istio Mesh mTLS](#istio-mesh-mtls)<sup>2</sup> | [Istio Mesh mTLS](#istio-mesh-mtls) |



### Istio Mesh mTLS
When we refer to "Istio Mesh mTLS" in the matrix above, we are assuming the following Istio functionality is leveraged:

* [Istio Auto mTLS](https://istio.io/docs/tasks/security/authentication/auto-mtls/)
* Have a default Istio mesh [policy](https://istio.io/docs/tasks/security/authentication/authn-policy/) enforcing `STRICT` mTLS
* Istio [sidecar autoinjection](https://istio.io/docs/setup/additional-setup/sidecar-injection/) enabled

#### Footnotes

* 0 - System components that do not expect to receive traffic from apps could be protected using `NetworkPolicies` or Istio `AuthorizationPolicies`
* 1 - The exact mechanism for providing manually providing certificated to system components is to be determined.
* 2 -  Primary use case for system component to app communication that we can think of may be a platform managed Prometheus scraping apps' `/metrics` endpoints
