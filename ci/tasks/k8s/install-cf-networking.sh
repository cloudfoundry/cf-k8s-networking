#!/usr/bin/env bash

set -euo pipefail

# ENV
: "${KUBECONFIG_CONTEXT:?}"
: "${BBL_STATE_DIR:?}"
: "${SYSTEM_NAMESPACE:?}"

function install() {
  workspace=${PWD}
  export KUBECONFIG="${workspace}/kubeconfig/config"
  kubectl config use-context ${KUBECONFIG_CONTEXT}

  tmp_dir="$(mktemp -d /tmp/values.XXXXXXXX)"
  values_yml="${tmp_dir}/values.yaml"

  echo 'Fetching environment variables for credhub...'
  pushd "bbl-state/${BBL_STATE_DIR}" > /dev/null
    eval "$(bbl print-env)"
  popd

  ./cf-k8s-networking/install/scripts/generate_values.rb "bbl-state/${BBL_STATE_DIR}/bbl-state.json" > ${values_yml}

  pushd cf-k8s-networking > /dev/null
    git_sha="$(cat .git/ref)"
  popd
  image_repo="gcr.io/cf-networking-images/cf-k8s-networking/cfroutesync:${git_sha}"

  echo "Deploying image '${image_repo}' to Kubernetes..."
  ytt -f cf-k8s-networking/install/ytt/networking/-f ${values_yml} \
    --data-value-yaml cfroutesync.image=${image_repo} | \
    kapp deploy -n "${SYSTEM_NAMESPACE}" -a cfroutesync \
    -f cf-k8s-networking/cfroutesync/crds/routebulksync.yaml \
    -f - \
    -y
}

function main() {
  install
}

main
