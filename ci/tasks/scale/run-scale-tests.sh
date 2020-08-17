#!/usr/bin/env bash

set -euo pipefail

function login_and_target() {
    cf api --skip-ssl-validation "https://api.$(cat env-metadata/dns-domain.txt)"
    CF_USERNAME=admin CF_PASSWORD=$(cat env-metadata/cf-admin-password.txt) cf auth
}

function run_scale_test() {
    export DOMAIN="apps.ci-scale-testing.routing.lol"
    export CLEANUP="true" #Remove when we run these tests regularly after they start to pass

    pushd cf-k8s-networking/test/scale
        ginkgo .
    popd
}

function hack_dns() {
    gcloud auth activate-service-account --key-file=<(echo "${GCP_SERVICE_ACCOUNT_KEY}") --project="${GCP_PROJECT}" 1>/dev/null 2>&1
    gcloud container clusters get-credentials ${CLUSTER_NAME} 1>/dev/null 2>&1
    system_ip=$(kubectl get svc -n istio-system -l "istio=istio-system-ingressgateway" -ojsonpath='{.items[0].status.loadBalancer.ingress[0].ip}')
    cat >> /etc/hosts <<EOF
${system_ip} api.$(cat env-metadata/dns-domain.txt)
${system_ip} login.$(cat env-metadata/dns-domain.txt)
${system_ip} uaa.$(cat env-metadata/dns-domain.txt)
EOF
}

function main() {
    hack_dns
    login_and_target
    run_scale_test
}

main
