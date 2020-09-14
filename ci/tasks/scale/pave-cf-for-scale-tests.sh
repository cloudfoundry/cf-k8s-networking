#!/usr/bin/env bash

set -euo pipefail

: "${NUMBER_OF_APPS:?}"

function login() {
    cf api --skip-ssl-validation "https://api.$(cat env-metadata/dns-domain.txt)"
    CF_USERNAME=admin CF_PASSWORD=$(cat env-metadata/cf-admin-password.txt) cf auth
}

function prepare_cf_foundation() {
    cf enable-feature-flag diego_docker
    cf update-quota default -r 3000 -m 3T
}

function deploy_apps() {
    org_name_prefix="scale-tests"
    space_name_prefix="scale-tests"
    number_of_org_spaces="$((NUMBER_OF_APPS / 10))"
    number_of_apps_per_org_space="$((NUMBER_OF_APPS / number_of_org_spaces))"

    for n in $(seq 0 ${number_of_org_spaces})
    do
      org_name="${org_name_prefix}-${n}"
      space_name="${space_name_prefix}-${n}"
      cf create-org "${org_name}"
      cf create-space -o "${org_name}" "${space_name}"
      cf target -o "${org_name}" -s "${space_name}"

      for i in $(seq 0 ${number_of_apps_per_org_space})
      do
        name="bin-$((n * 100 + i))"
        echo $name
        cf push $name -o cfrouting/proxy -m 128M -k 256M -i 2 &
        sleep 2
      done
      wait
    done
}

function main() {
    sleep 10
    # hopefully wait for til it works?
    curl -vvv --retry 300 -k "https://api.$(cat env-metadata/dns-domain.txt)"

    login
    prepare_cf_foundation
    deploy_apps
}

main
