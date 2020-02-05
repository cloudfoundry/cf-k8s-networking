#!/bin/bash

set -euo pipefail

function main() {
  mkdir -p $PWD/kubeconfig
  kubeconfig_path="${PWD}/kubeconfig/config"
  bbl_state="bbl-state/${BBL_STATE_DIR}"

  source "cf-k8s-networking-ci/tasks/k8s/utils.sh"
  extract_kubeconfig $kubeconfig_path $bbl_state
}


main
