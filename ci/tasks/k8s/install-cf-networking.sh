#!/usr/bin/env bash

set -euo pipefail

# ENV
: "${KUBECONFIG_CONTEXT:?}"
: "${BBL_STATE_DIR:?}"

function install() {
  workspace=${PWD}
  export KUBECONFIG="${workspace}/kubeconfig/config"
  kubectl config use-context ${KUBECONFIG_CONTEXT}

  tmp_dir="$(mktemp -d /tmp/helm-secrets.XXXXXXXX)"
  secrets_yaml="${tmp_dir}/secrets.yaml"

  echo 'Fetching environment variables for credhub...'
  pushd "bbl-state/${BBL_STATE_DIR}" > /dev/null
    eval "$(bbl print-env)"
  popd

  ./cf-k8s-networking/install/scripts/generate_values.rb "bbl-state/${BBL_STATE_DIR}/bbl-state.json" > $secrets_yaml

  echo 'Applying CRDs...'
  kubectl apply -f "cf-k8s-networking/cfroutesync/crds/routebulksync.yaml"

  echo 'Deploying to Kubernetes...'
  helm template cf-k8s-networking/install/helm/networking/ --values $secrets_yaml | kubectl apply -f-
}

function main() {
  install
}

main
