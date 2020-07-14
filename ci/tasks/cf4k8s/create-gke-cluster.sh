#!/usr/bin/env bash

set -euo pipefail

# ENV
: "${CLUSTER_NAME:?}"
: "${GCP_SERVICE_ACCOUNT_KEY:?}"

: "${CLOUDSDK_COMPUTE_REGION:?}"
: "${CLOUDSDK_COMPUTE_ZONE:?}"
: "${ENABLE_IP_ALIAS:?}"
: "${GCP_PROJECT:?}"
: "${MACHINE_TYPE:?}"
: "${NUM_NODES:?}"

function latest_cluster_version() {
  gcloud container get-server-config --zone us-west1-a 2>/dev/null | yq .validMasterVersions[0] -r
}

function create_cluster() {
    gcloud auth activate-service-account --key-file=<(echo "${GCP_SERVICE_ACCOUNT_KEY}") --project="${GCP_PROJECT}" 1>/dev/null 2>&1

    if gcloud container clusters describe ${CLUSTER_NAME} > /dev/null; then
        echo "${CLUSTER_NAME} already exists! Destroying..."
        gcloud container clusters delete ${CLUSTER_NAME} --quiet
    fi


    additional_args=()
    if [ "${ENABLE_IP_ALIAS}" = true ]; then
        additional_args+=("--enable-ip-alias")
    fi

    echo "Creating cluster: ${CLUSTER_NAME} ..."
    gcloud container clusters create ${CLUSTER_NAME} \
        --cluster-version=$(latest_cluster_version) \
        --machine-type=${MACHINE_TYPE} \
        --labels team=cf-k8s-networking-ci \
        --enable-network-policy \
        --num-nodes "${NUM_NODES}" \
        "${additional_args[@]}"
}

function main() {
    create_cluster
}

main
