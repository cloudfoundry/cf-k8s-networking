#!/usr/bin/env bash

set -euo pipefail

: "${NUMBER_OF_APPS:?}"
: "${CLEANUP:?}"

function login_and_target() {
    cf api --skip-ssl-validation "https://api.$(cat env-metadata/dns-domain.txt)"
    CF_USERNAME=admin CF_PASSWORD=$(cat env-metadata/cf-admin-password.txt) cf auth
}

function run_scale_test() {
    export DOMAIN="apps.ci-scale-testing.routing.lol"
    export NUMBER_OF_APPS=${NUMBER_OF_APPS}

    pushd cf-k8s-networking/test/scale
        ginkgo -v .
    popd
}

function main() {
    login_and_target
    run_scale_test
}

main
