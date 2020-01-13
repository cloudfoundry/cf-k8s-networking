#!/bin/bash

set -euo pipefail

# ENV
: "${KUBECONFIG_CONTEXT:?}"

function install_grafana_dashboard() {
  export KUBECONFIG="${PWD}/kubeconfig/config"
  kubectl config use-context ${KUBECONFIG_CONTEXT}

  dashboard_file="${PWD}/cf-k8s-networking/doc/metrics/dashboard.json"

  jq -n '{ "dashboard": input }'  $dashboard_file | jq '.dashboard.id = null' | jq '.dashboard.uid = "indicators"' -c > "/tmp/dashboard.json"

  kubectl proxy --port=8080 &
  proxy_pid=$!

  # Delete old dashboard
  curl -H 'Content-Type: application/json' -H 'Accept: application/json' -XDELETE http://localhost:8080/api/v1/namespaces/istio-system/services/grafana:http/proxy/api/dashboards/uid/indicators
  # Create dashboard
  curl -H 'Content-Type: application/json' -H 'Accept: application/json' -XPOST http://localhost:8080/api/v1/namespaces/istio-system/services/grafana:http/proxy/api/dashboards/db -d "@/tmp/dashboard.json"

  kill ${proxy_pid}
}


function main() {
  install_grafana_dashboard
}

main
