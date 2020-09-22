#!/usr/bin/env bash

set -euo pipefail

# move cf-for-k8s to output dir
cp -r cf-for-k8s/* cf-for-k8s-scaled/

# scale-up istio deployments via istio-values.yaml
ytt -f cf-k8s-networking-ci/ci/tasks/scale/istioctl-values-overlay.yml -f cf-for-k8s/build/istio/istioctl-values.yaml > cf-for-k8s-scaled/build/istio/istioctl-values.yaml

# replace values.yaml file to change istio version from 1.6.4 to 1.6.9
# TODO remove this once istio has been upgraded
mv cf-k8s-networking-ci/ci/tasks/scale/istio-values.yml cf-for-k8s-scaled/build/istio/values.yaml

# remove the overlay that turns the ingressgateways into a daemonset
rm cf-for-k8s-scaled/build/istio/overlays/ingressgateway-daemonset.yaml

# remove the overlay that scales pilot replicas to 1
# we don't use the hpas for this pipeline, but it's easier to just delete the whole file.
# TODO: removing this while we revert to a working version of cf-for-k8s. The working version doesn't include this overlay. See this comment for details: https://www.pivotaltracker.com/story/show/171678788/comments/218182506
# rm cf-for-k8s-scaled/config/istio/remove-hpas-and-scale-istiod.yml

# generate new XXX files
cf-for-k8s-scaled/build/istio/build.sh
