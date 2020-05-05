# 2. Directly Create Istio Resources

Date: 2019-10-22

## Status

Superseded by [ADR 10](0010-route-crd-and-kubebuilder-instead-of-metacontroller.md)

## Context

In the [original proposal](https://docs.google.com/document/d/1EYRBVuQedU1r0zexgi8oMSOEFgaMNzM8JWBje3XuweU) for basic http 
ingress routing for CF on Kubernetes, we proposed writing a controller that wrote custom Route resources to the Kubernetes
API. Additionally we would develop a second controller that read the Route CRDs and would create k8s Services and Istio 
VirtualServices.

We discovered several issues with this design. First, we realized that we must have a single VirtualService per FQDN.
While multiple VirtualServices for the same FQDN are [technically permitted by Istio](https://istio.io/docs/ops/traffic-management/deploy-guidelines/#multiple-virtual-services-and-destination-rules-for-the-same-host),
the order in which the match rules for the paths are applied is non-deterministic. In CF we expect that the longest path
prefix is matched first, so this behavior did not suit our needs. 

Since we had to aggregate multiple Route resources to construct a single VirtualService, this meant we could not use
Metacontroller for our second controller. Having multiple "parent" Routes for a single set of "children" VirtualServices would
violate Metacontroller's assumptions.

While we could build a custom second controller using Kubebuilder, we decided that for simplicity and expediency we could
just omit the creation of the Route CRDs for the time being.

## Decision

* CF Route Syncer will directly create k8s Services and Istio VirtualServices instead of creating intermediate Route CRDs

## Consequences

* We will be able to implement a demo-able MVP more rapidly
* We will only have to maintain a single metacontroller webhook rather than
  implement a second controller that aggregates Route CRDs
* We will no longer have a close-representation of Cloud Controller routes in
  the k8s API
* This will couple us more tightly with Istio, but we believe we can easily undo
  this decision

