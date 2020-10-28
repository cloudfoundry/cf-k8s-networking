#!/usr/bin/env bash

set -euo pipefail

: "${NUMBER_OF_APPS:?}"

function login_and_target() {
    cf api --skip-ssl-validation "https://api.$(cat env-metadata/dns-domain.txt)"
    CF_USERNAME=admin CF_PASSWORD=$(cat env-metadata/cf-admin-password.txt) cf auth
}

function routecontroller_image() {
    image="$(< cf-k8s-networking/config/values.yaml yq -r '.routecontroller.image' | cut -d'@' -f1)"
    version="$(cat cf-k8s-networking/.git/ref)"
    echo -n "${image}:${version}"
}

function run_scale_test() {
    export DOMAIN="apps.ci-scale-testing.routing.lol"
    export CLEANUP="true" #Remove when we run these tests regularly after they start to pass
    export NUMBER_OF_APPS=${NUMBER_OF_APPS}
    ROUTECONTROLLER_IMAGE="$(routecontroller_image)"
    export ROUTECONTROLLER_IMAGE
    export INGRESS_PROVIDER

    pushd cf-k8s-networking/test/scale
        ginkgo -v .
    popd
}

function main() {
    login_and_target
    run_scale_test
}

main
