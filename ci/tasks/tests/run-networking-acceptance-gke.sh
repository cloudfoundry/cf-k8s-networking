#!/bin/bash

set -euo pipefail

# ENV
: "${INTEGRATION_CONFIG_FILE:?}"
: "${CLUSTER_NAME:?}"
: "${CLOUDSDK_COMPUTE_REGION:?}"
: "${CLOUDSDK_COMPUTE_ZONE:?}"
: "${GCP_SERVICE_ACCOUNT_KEY:?}"
: "${GCP_PROJECT:?}"

function main() {
    export KUBECONFIG=kubeconfig/config

    gcloud auth activate-service-account --key-file=<(echo "${GCP_SERVICE_ACCOUNT_KEY}") --project="${GCP_PROJECT}" 1>/dev/null 2>&1
    gcloud container clusters get-credentials ${CLUSTER_NAME} 1>/dev/null 2>&1

    local kubeconfig_path="${PWD}/${KUBECONFIG}"
    local config="${PWD}/integration-config/${INTEGRATION_CONFIG_FILE}"

    pushd cf-k8s-networking/networking-acceptance-tests > /dev/null
        ./bin/test_local "${config}" "${kubeconfig_path}"
    popd
}

main
