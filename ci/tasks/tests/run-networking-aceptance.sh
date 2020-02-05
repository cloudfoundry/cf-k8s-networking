#!/bin/bash

set -euo pipefail

# ENV
: "${INTEGRATION_CONFIG_FILE:?}"

function main() {
  local kubeconfig="${PWD}/kubeconfig/config"
  local config="${PWD}/integration-config/${INTEGRATION_CONFIG_FILE}"

  cd cf-k8s-networking/networking-acceptance-tests
  ./bin/test_local "${config}" "${kubeconfig}"
}


main
