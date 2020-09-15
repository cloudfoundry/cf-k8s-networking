# 17. Moving Istio and Related Configuration to CF-for-K8s Repo

Date: 2020-09-15

## Status

Accepted

## Context

This ADR reverts the decision made in [ADR # 7. Maintain Generated
Istio](./0007-maintain-generated-istio.md).

The networking config and related Istio config is spread widely throughout both
cf-for-k8s and cf-k8s-networking. Having the config in both places has made
processes such as updating networking config, versioning routecontroller,
upgrading Istio, deciding where some networking config should exist (in this repo or
in the cf-for-k8s repo), and so on complicated.


## Decision

Relint and Networking came to a consensus to do move Istio configuration to
cf-for-k8s repo so we reduce the overhead of having incompatibility between
cf-k8s-networking and cf-for-k8s.

* move Istio config generation and overlays folder `istio-install`
* move Istio generated and other networking config folders `config/istio`,
  `config/istio-generated`
* when contributing to networking in cf-for-k8s open PR and tag it with
  `networking` tag
* create CI job to run acceptance tests upon new networking PRs in cf-for-k8s
* update the documentation to reflect the change
* update CI jobs depending on Istio config in this repo (such istio-upgrade,
  images, scaling, etc)


## Consequences

* cf-k8s-networking now mostly only contains routecontroller, CI and tests.
* Istio config now lives in the [cf-for-k8s
  repo](https://github.com/cloudfoundry/cf-for-k8s/tree/master/config/istio) and
  whenever need to make changes to Istio config, we do so through a PR to
  cf-for-k8s.
