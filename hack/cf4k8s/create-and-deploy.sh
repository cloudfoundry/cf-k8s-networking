#!/usr/bin/env bash

set -euo pipefail

ROOT="$(cd $(dirname $0) && pwd)"

source "${ROOT}/methods.sh"

# ENV
CLUSTER_NAME=${CLUSTER_NAME:-$1}
CF_DOMAIN=${CF_DOMAIN:-$CLUSTER_NAME.routing.lol}
: "${SHARED_DNS_ZONE_NAME:="routing-lol"}"
: "${GCP_PROJECT:="cf-routing"}"


function main() {
  create_and_target_cluster
  credhub_get_gcp_service_account_key
  deploy_cf_for_k8s
  configure_dns
  target_cf
  enable_docker
}

main
