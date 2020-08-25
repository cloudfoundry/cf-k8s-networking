#!/usr/bin/env bash

set -euo pipefail

# ENV
: "${APP_NAME:?}"
: "${ORG_NAME:?}"
: "${SPACE_NAME:?}"
: "${INSTANCES:?}"

ROOT="$(cd "$(dirname "${0}")/../../../.." && pwd)"

function target_cf() {
    local cf_domain=$(cat "${ROOT}/cf-install-values/cf-install-values.yml" | \
        grep system_domain | awk '{print $2}' | tr -d '"')

    cf_with_retry api --skip-ssl-validation "https://api.${cf_domain}"
    local password=$(cat "${ROOT}/cf-install-values/cf-install-values.yml" | \
        grep cf_admin_password | awk '{print $2}')
    cf_with_retry auth "admin" "${password}"
}

function create_org_and_space() {
    cf_with_retry create-org "${ORG_NAME}"
    cf_with_retry create-space -o "${ORG_NAME}" "${SPACE_NAME}"
}

function deploy_app() {
    local name="${1}"
    cf_with_retry push "${name}" -o "cfrouting/httpbin" -i "${INSTANCES}"
}

function cf_with_retry() {
    cf_command=$*

    set +euo pipefail

    for i in {1..3}
    do
        echo "Running cf ${cf_command}..."
        cf $cf_command && set -euo pipefail && return
        sleep 10
    done

    echo "cf_with_retry command has failed 3 times"
    exit
}

function main() {
    target_cf
    create_org_and_space
    cf_with_retry target -o "${ORG_NAME}" -s "${SPACE_NAME}"
    cf_with_retry enable-feature-flag diego_docker
    deploy_app "${APP_NAME}"
}

main
