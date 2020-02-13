#!/usr/bin/env bash

set -euo pipefail

# ENV
: "${CLUSTER_NAME:?}"
: "${CLOUDSDK_COMPUTE_REGION:?}"
: "${CLOUDSDK_COMPUTE_ZONE:?}"
: "${GCP_SERVICE_ACCOUNT_KEY:?}"
: "${GCP_PROJECT:?}"


function extract_kubeconfig() {
     export KUBECONFIG=kubeconfig/config

     gcloud auth activate-service-account --key-file=<(echo "${GCP_SERVICE_ACCOUNT_KEY}") --project="${GCP_PROJECT}" 1>/dev/null 2>&1
     gcloud container clusters get-credentials ${CLUSTER_NAME} 1>/dev/null 2>&1

    echo "kubeconfig extracted! 🤗"
}

function main() {
    extract_kubeconfig
}

main
