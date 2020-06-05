#!/usr/bin/env bash

set -euo pipefail

source cf-k8s-networking-ci/ci/tasks/helpers.sh

function run_upgrade_uptime_tests() {
    pushd "cf-k8s-networking/test/uptime"
        ginkgo -v -r -p .
    popd
}

function main() {
    target_k8s_cluster #from helpers.sh
    INSTALL_VALUES_FILEPATH=cf-install-values/cf-install-values.yml target_cf_with_install_values #from helpers.sh
    run_upgrade_uptime_tests
}

main
