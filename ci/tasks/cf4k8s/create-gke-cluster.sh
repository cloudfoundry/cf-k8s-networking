#!/usr/bin/env bash

set -euo pipefail

# ENV
: "${CLUSTER_NAME:?}"
: "${CLOUDSDK_COMPUTE_REGION:?}"
: "${CLOUDSDK_COMPUTE_ZONE:?}"
: "${MACHINE_TYPE:?}"
: "${GCP_SERVICE_ACCOUNT_KEY:?}"
: "${GCP_PROJECT:?}"


function create_cluster() {
    gcloud auth activate-service-account --key-file=<(echo "${GCP_SERVICE_ACCOUNT_KEY}") --project="${GCP_PROJECT}" 1>/dev/null 2>&1

    if gcloud container clusters describe ${CLUSTER_NAME} > /dev/null; then
        echo "${CLUSTER_NAME} already exists! Destroying..."
        gcloud container clusters delete ${CLUSTER_NAME} --quiet
    fi

    echo "Creating cluster: ${CLUSTER_NAME} ..."
    gcloud container clusters create ${CLUSTER_NAME} --machine-type=${MACHINE_TYPE}
}

function main() {
    create_cluster
}

main
