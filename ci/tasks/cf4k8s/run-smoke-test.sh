#!/usr/bin/env bash

set -euo pipefail

ROOT="$(cd "$(dirname "${0}")/../../../.." && pwd)"

FLAKE_ATTEMPTS="${FLAKE_ATTEMPTS:-0}"

function run_smoke_test() {
    DNS_DOMAIN=$(cat env-metadata/dns-domain.txt)
    export SMOKE_TEST_API_ENDPOINT="https://api.${DNS_DOMAIN}"
    export SMOKE_TEST_APPS_DOMAIN="apps.${DNS_DOMAIN}"
    export SMOKE_TEST_USERNAME=admin
    export SMOKE_TEST_PASSWORD=$(cat env-metadata/cf-admin-password.txt)
    cd "${ROOT}/cf-for-k8s/tests/smoke"
    ginkgo -v -flakeAttempts="${FLAKE_ATTEMPTS}" ./
}

function main() {
    run_smoke_test
}

main
