#!/bin/bash

set -euo pipefail

# ENV
: "${KUBECONFIG_CONTEXT:?}"

function install_grafana_dashboard() {
  export KUBECONFIG="${PWD}/kubeconfig/config"
  kubectl config use-context ${KUBECONFIG_CONTEXT}

  dashboard_file="${PWD}/cf-k8s-networking/doc/metrics/dashboard.json"
  kubectl delete jobs create-dashboard  --ignore-not-found=true
  kubectl run --generator=job/v1 create-dashboard --image alpine --restart Never --dry-run -o json --command -- \
    sh -c "apk add curl && curl -H 'Content-Type: application/json' -H 'Accept: application/json' -XPOST http://grafana.istio-system.svc.cluster.local:3000/api/dashboards/db -d '""$(jq -n '{ "dashboard": input }'  $dashboard_file | jq '.dashboard.id = null' | jq '.dashboard.uid = null' -c)""'" | \
    jq '.spec.ttlSecondsAfterFinished = 10' | \
    kubectl apply -f -
}


function main() {
  install_grafana_dashboard
}

main
