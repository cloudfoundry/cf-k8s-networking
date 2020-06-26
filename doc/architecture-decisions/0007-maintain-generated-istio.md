# 7. Maintain Generated Istio

Date: 2020-02-19

## Status

Accepted

## Context 🤔
Cf-k8s-networking was designed to be integrated with [cf-for-k8s](https://github.com/cloudfoundry/cf-for-k8s/). The Istio installation used to be maintained by [cf-for-k8s](https://github.com/cloudfoundry/cf-for-k8s/), but the networking team needed to be able to easily make changes to [Istio](https://istio.io/) configuration to enable more networking features for [Cloud Foundry](https://www.cloudfoundry.org/).


## Decision
We decided to move the scripts to build Istio configuration, and maintain a generated Istio configuration within the cf-k8s-networking repository. 

The build scripts and `ytt` overlays for Istio live in [config/istio](../../config/istio). The [`generate.sh`](../../config/istio/generate.sh) script generates Istio installation yaml to stdout and the [`build.sh`](../../config/istio/build.sh) script generates the maintained Istio installation yaml in [config/istio-generated/xxx-generated-istio.yaml](../../config/istio-generated/xxx-generated-istio.yaml).

[cf-for-k8s](https://github.com/cloudfoundry/cf-for-k8s/) pulls in this generated Istio configuration using [`vendir`](https://github.com/k14s/vendir) as seen in [vendir.yml](https://github.com/cloudfoundry/cf-for-k8s/blob/master/vendir.yml).

See [updating Istio docs](https://github.com/cloudfoundry/cf-k8s-networking/blob/develop/doc/update-istio.md) for instructions.

## Consequences
When making changes to anything related to the Istio installation (build scripts, `ytt` overlays, Istio configuration), developers need to also generate the new corresponding Istio yaml following the doc [doc/update-istio.md](../update-istio.md)
