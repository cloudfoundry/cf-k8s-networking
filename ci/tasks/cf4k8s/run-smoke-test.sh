#!/usr/bin/env sh

set -e

function run_smoke_test() {
    DNS_DOMAIN=$(cat env-metadata/dns-domain.txt)
    export SMOKE_TEST_API_ENDPOINT="https://api.${DNS_DOMAIN}"
    export SMOKE_TEST_APPS_DOMAIN="${DNS_DOMAIN}"
    export SMOKE_TEST_USERNAME=admin
    export SMOKE_TEST_PASSWORD=$(cat env-metadata/cf-admin-password.txt)
    cf-for-k8s-master/hack/run-smoke-tests.sh
}

function main() {
    run_smoke_test
}

main
