#!/usr/bin/env bash

set -euo pipefail

function login() {
    cf api --skip-ssl-validation "https://api.$(cat env-metadata/dns-domain.txt)"
    CF_USERNAME=admin CF_PASSWORD=$(cat env-metadata/cf-admin-password.txt) cf auth
}

function prepare_cf_foundation() {
    cf enable-feature-flag diego_docker
    cf update-quota default -r 3000 -m 3T

    export ORG_NAME="scale-tests"
    export SPACE_NAME="${ORG_NAME}"
    cf create-org "${ORG_NAME}"
    cf create-space -o "${ORG_NAME}" "${SPACE_NAME}"
    cf target -o "${ORG_NAME}" -s "${SPACE_NAME}"
}

function deploy_apps() {
    for n in {0..99}
    do
      for i in {0..9}
      do
        name="bin-$((n * 10 + i))"
        echo $name
        cf push $name -o cfrouting/httpbin8080 -m 256M -k 256M -i 2 &
      done
      wait
    done
}

function main() {
    login
    prepare_cf_foundation
    deploy_apps
}

main
