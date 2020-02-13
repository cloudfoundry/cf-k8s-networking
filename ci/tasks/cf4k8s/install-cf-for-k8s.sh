#!/usr/bin/env bash

set -euo pipefail

# ENV
: "${CLUSTER_NAME:?}"
: "${CF_DOMAIN:?}"
: "${CLOUDSDK_COMPUTE_REGION:?}"
: "${CLOUDSDK_COMPUTE_ZONE:?}"
: "${GCP_SERVICE_ACCOUNT_KEY:?}"
: "${GCP_PROJECT:?}"
: "${SHARED_DNS_ZONE_NAME:?}"


function install_cf() {
     export KUBECONFIG=kube-config.yml

     gcloud auth activate-service-account --key-file=<(echo "${GCP_SERVICE_ACCOUNT_KEY}") --project="${GCP_PROJECT}" 1>/dev/null 2>&1
     gcloud container clusters get-credentials ${CLUSTER_NAME} 1>/dev/null 2>&1

     echo "Generating install values..."
     cf-for-k8s-master/hack/generate-values.sh "${CF_DOMAIN}" > cf-install-values.yml

     echo "Installing CF..."
     cf-for-k8s-master/bin/install-cf.sh cf-install-values.yml

     bosh interpolate --path /cf_admin_password cf-install-values.yml > env-metadata/cf-admin-password.txt
     echo "${CF_DOMAIN}" > env-metadata/dns-domain.txt
}

function configure_dns() {
    echo "Discovering Istio Gateway LB IP"
    external_static_ip=""
    while [ -z $external_static_ip ]; do
      sleep 1
      external_static_ip=$(kubectl get services/istio-ingressgateway -n istio-system --output="jsonpath={.status.loadBalancer.ingress[0].ip}")
    done

    echo "Configuring DNS for external IP: ${external_static_ip}"
    gcloud dns record-sets transaction start --zone="${SHARED_DNS_ZONE_NAME}"
    gcp_records_json="$( gcloud dns record-sets list --zone "${SHARED_DNS_ZONE_NAME}" --name "*.${CF_DOMAIN}" --format=json )"
    record_count="$( echo "${gcp_records_json}" | jq 'length' )"
    if [ "${record_count}" != "0" ]; then
    existing_record_ip="$( echo "${gcp_records_json}" | jq -r '.[0].rrdatas | join(" ")' )"
    gcloud dns record-sets transaction remove --name "*.${CF_DOMAIN}" --type=A --zone="${SHARED_DNS_ZONE_NAME}" --ttl=300 "${existing_record_ip}" --verbosity=debug
    fi
    gcloud dns record-sets transaction add --name "*.${CF_DOMAIN}" --type=A --zone="${SHARED_DNS_ZONE_NAME}" --ttl=300 "${external_static_ip}" --verbosity=debug

    echo "Contents of transaction.yaml:"
    cat transaction.yaml
    gcloud dns record-sets transaction execute --zone="${SHARED_DNS_ZONE_NAME}" --verbosity=debug

    resolved_ip=''
    while [ "$resolved_ip" != "$external_static_ip" ]; do
      echo "Waiting for DNS to propagate..."
      sleep 5
      resolved_ip=$(nslookup "*.${CF_DOMAIN}" | grep Address | grep -v ':53' | cut -d ' ' -f2)
    done
}

function main() {
    install_cf
    configure_dns
}

main
