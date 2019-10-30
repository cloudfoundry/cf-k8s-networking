#!/usr/bin/env bash

set -euo pipefail

# ENV
: "${KUBECONFIG_CONTEXT:?}"
: "${FILES_TO_APPLY:?}"

function kubectl_apply_all() {
  workspace=${PWD}
  export KUBECONFIG="${workspace}/kubeconfig/config"

  pushd k8s-config-dir > /dev/null
    kubectl config use-context ${KUBECONFIG_CONTEXT}

    for file in ${FILES_TO_APPLY}
    do
        echo "Applying ${file}"
        kubectl apply -f $file
        sleep 5  # give k8s time to converge
    done
  popd
}

function main() {
  kubectl_apply_all
}

main
