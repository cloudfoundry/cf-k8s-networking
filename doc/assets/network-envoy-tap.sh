#!/usr/bin/env bash

#################################################################
# This script creates a rudimentary EnvoyFilter for
# tapping into (inbound) HTTP traffc of a CF app.
# One file per http request is created unter /etc/istio/proxy/
#################################################################


set -eo pipefail

if [ -z "$KUBECONFIG" ]; then
  echo "KUBECONFIG not set."
  exit 1
fi

APP_NAME=$1

if [ -z "$APP_NAME" ]; then
  echo "Usage: $0 <app_name>"
  exit 1
fi

APP_GUID=$(cf app "$APP_NAME" --guid)

kubectl -n cf-workloads apply -f - <<EOF
apiVersion: networking.istio.io/v1alpha3
kind: EnvoyFilter
metadata:
  name: $APP_GUID-tap
  namespace: cf-workloads
spec:
  workloadSelector:
    labels:
      cloudfoundry.org/app_guid: $APP_GUID
  configPatches:
    - applyTo: HTTP_FILTER
      match:
        context: SIDECAR_INBOUND
        listener:
          name: "virtualInbound"
          filterChain:
            filter:
              name: "envoy.http_connection_manager"
              subFilter:
                name: "envoy.router"
      patch:
        operation: INSERT_BEFORE
        value:
          name: envoy.filters.http.tap
          config:
            common_config:
              static_config:
                match_config:
                  any_match: true
                output_config:
                  sinks:
                    - format: JSON_BODY_AS_BYTES
                      file_per_tap:
                        path_prefix: /etc/istio/proxy/tap
EOF
