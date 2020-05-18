#!/usr/bin/env bash

set -euo pipefail

# ENV
: "${APP_NAME:?}"

function deploy_app() {
    local name="${1}"
    cf push ${name} -o cfrouting/httpbin8080
}

function main() {
    deploy_app "${APP_NAME}"
}
