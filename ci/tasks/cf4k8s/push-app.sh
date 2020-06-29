#!/usr/bin/env bash

set -euo pipefail

# ENV
: "${APP_NAME:?}"
: "${ORG_NAME:?}"
: "${SPACE_NAME:?}"

ROOT="$(cd "$(dirname "${0}")/../../../.." && pwd)"

function target_cf() {
    local cf_domain=$(cat "${ROOT}/cf-install-values/cf-install-values.yml" | \
        grep system_domain | awk '{print $2}' | tr -d '"')

    cf api --skip-ssl-validation "https://api.${cf_domain}"
    local password=$(cat "${ROOT}/cf-install-values/cf-install-values.yml" | \
        grep cf_admin_password | awk '{print $2}')
    cf auth "admin" "${password}"
}

function create_org_and_space() {
    cf create-org "${ORG_NAME}"
    cf create-space -o "${ORG_NAME}" "${SPACE_NAME}"
}

function deploy_app() {
    local name="${1}"
    cf push "${name}" -o "cfrouting/httpbin"
}

function main() {
    target_cf
    create_org_and_space
    cf target -o "${ORG_NAME}" -s "${SPACE_NAME}"
    cf enable-feature-flag diego_docker
    deploy_app "${APP_NAME}"
}

main
