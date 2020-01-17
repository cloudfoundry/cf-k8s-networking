#!/bin/bash

set -euo pipefail

# ENV
: "${KUBECONFIG_CONTEXT:?}"

function enabled_sidecar_injection() {
  workspace=${PWD}
  export KUBECONFIG="${PWD}/kubeconfig/config"

  # Enable Istio Sidecar Injection for app workloads
  kubectl label namespace cf-workloads istio-injection=enabled --overwrite=true
  kubectl label namespace cf-system istio-injection=enabled --overwrite=true
  kubectl label namespace metacontroller istio-injection=enabled --overwrite=true
}

function main() {
  enabled_sidecar_injection
}

main
