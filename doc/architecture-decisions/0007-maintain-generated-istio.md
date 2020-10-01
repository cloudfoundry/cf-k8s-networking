# 7. Maintain Generated Istio

Date: 2020-02-19

## Status

Partially superseded by [ADR
17](./0017-moving-istio-configuration-out-of-this-repo.md). To look at the files
at the moment this ADR was written you can browse files at [3f55af5
commit](https://github.com/cloudfoundry/cf-k8s-networking/tree/3f55af54912a527de16a8f70645018e4f13f9dba).

## Context 🤔
Cf-k8s-networking was designed to be integrated with
[cf-for-k8s](https://github.com/cloudfoundry/cf-for-k8s/). The Istio
installation used to be maintained by
[cf-for-k8s](https://github.com/cloudfoundry/cf-for-k8s/), but the networking
team needed to be able to easily make changes to [Istio](https://istio.io/)
configuration to enable more networking features for [Cloud
Foundry](https://www.cloudfoundry.org/).


## Decision
We decided to move the scripts to build Istio configuration, and maintain a
generated Istio configuration within the cf-k8s-networking repository. 

The build scripts and `ytt` overlays for Istio live in this repo (links removed
as they are no longer relevant or accurate). **UPDATE** This configuration has
moved as a result of [ADR
017](./0017-moving-istio-configuration-out-of-this-repo.md).

## Consequences
When making changes to anything related to the Istio installation (build scripts, `ytt` overlays, Istio configuration), developers need to also generate the new corresponding Istio yaml following the doc [doc/update-istio.md](../update-istio.md)

