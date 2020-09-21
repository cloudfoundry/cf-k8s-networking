#!/usr/bin/env bash

set -euo pipefail

# move cf-for-k8s to output dir
cp -r cf-for-k8s cf-for-k8s-scaled

# run ytt on istioctl-values and on values
ytt -f cf-k8s-networking-ci/tasks/scale/istioctl-values-overlay.yml -f cf-for-k8s/build/istio/istioctl-values.yaml > cf-for-k8s-scaled/build/istio/istioctl-values.yaml
ytt -f cf-k8s-networking-ci/tasks/scale/istio-values-overlay.yml -f cf-for-k8s/build/istio/values.yaml > cf-for-k8s-scaled/build/istio/values.yaml

# generate new XXX files
cf-for-k8s-scaled/build/istio/build.sh
