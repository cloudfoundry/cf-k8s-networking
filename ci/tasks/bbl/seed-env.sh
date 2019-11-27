#!/bin/bash

set -e -u -o pipefail

: "${BBL_STATE_DIR:?}"
: "${SYSTEM_DOMAIN:?}"
: "${APPS_DOMAIN:?}"

lb_output_path="certs/load-balancer"

cats_output_path="cats_integration_config.json"

function write_cats_config() {
  if [ ! -f "${cats_output_path}" ]; then
    cat <<- EOF > "${cats_output_path}"
{
    "api": "api.${SYSTEM_DOMAIN}",
    "apps_domain": "${APPS_DOMAIN}",
    "admin_user": "admin",
    "admin_password": "will-be-overridden-by-later-task",
    "skip_ssl_validation": true,
    "use_http": true,
    "backend": "diego",
    "default_timeout": 60,
    "include_apps": true,
    "include_v3": true,
    "include_capi_experimental": true,
    "include_capi_no_bridge": true,
    "include_routing": true,
    "include_detect": true,
    "include_sso": true,
    "include_container_networking": true,
    "include_backend_compatability": false,
    "include_credhub": false,
    "include_deployments": false,
    "include_docker": false,
    "include_internet_dependent": true,
    "include_isolation_segments": false,
    "include_private_docker_registry": false,
    "include_route_services": false,
    "include_routing_isolation_segments": false,
    "include_security_groups": false,
    "include_service_discovery": false,
    "include_services": false,
    "include_service_instance_sharing": false,
    "include_tasks": false,
    "include_ssh": false,
    "include_zipkin": false,
    "use_existing_user": false,
    "keep_user_at_suite_end": false
}
EOF
  fi
}

smoke_config_output_path="smoke_cats_integration_config.json"

function write_smoke_test_config() {
  if [ ! -f "${smoke_config_output_path}" ]; then
    cat <<- EOF > "${smoke_config_output_path}"
{
    "api": "api.${SYSTEM_DOMAIN}",
    "apps_domain": "${APPS_DOMAIN}",
    "admin_user": "admin",
    "admin_password": "will-be-overridden-by-later-task",
    "skip_ssl_validation": true,
    "use_http": true,
    "backend": "diego",
    "default_timeout": 60,
    "include_apps": false,
    "include_v3": false,
    "include_capi_experimental": false,
    "include_capi_no_bridge": false,
    "include_routing": true,
    "include_detect": false,
    "include_sso": false,
    "include_container_networking": false,
    "include_backend_compatability": false,
    "include_credhub": false,
    "include_deployments": false,
    "include_docker": false,
    "include_internet_dependent": false,
    "include_isolation_segments": false,
    "include_private_docker_registry": false,
    "include_route_services": false,
    "include_routing_isolation_segments": false,
    "include_security_groups": false,
    "include_service_discovery": false,
    "include_services": false,
    "include_service_instance_sharing": false,
    "include_tasks": false,
    "include_ssh": false,
    "include_zipkin": false,
    "use_existing_user": false,
    "keep_user_at_suite_end": false
}
EOF
  fi
}

apps_domain_vars_file="apps_domains_vars.yml"
function write_apps_domains_vars_file() {
  if [ ! -f "${apps_domain_vars_file}" ]; then
    cat <<- EOF > "${apps_domain_vars_file}"
app_domains:
- ${APPS_DOMAIN} # DNS points to istio ingress router
- ${SYSTEM_DOMAIN} # DNS points to gorouter
smoke_test_app_domain: ${APPS_DOMAIN}
EOF
  fi
}

function write_load_balancer_certs() {
  if [ ! -d "${lb_output_path}" ]; then
    mkdir -p "${lb_output_path}"
    pushd "${lb_output_path}"
      local cert_cn
      cert_cn="*.${SYSTEM_DOMAIN}"
      certstrap --depot-path "." init --passphrase '' --common-name "server-ca"
      certstrap --depot-path "." request-cert --passphrase '' --common-name "${cert_cn}" --csr "$SYSTEM_DOMAIN.csr" --key "${SYSTEM_DOMAIN}.key"
      certstrap --depot-path "." sign --CA "server-ca" "${SYSTEM_DOMAIN}"

      mv "${SYSTEM_DOMAIN}.csr" "server.csr"
      mv "${SYSTEM_DOMAIN}.crt" "server.crt"
      mv "${SYSTEM_DOMAIN}.key" "server.key"
    popd
  fi
}

function commit_bbl_state() {
  if [[ -n $(git status --porcelain) ]]; then
    git config user.name "CI Bot"
    git config user.email "cf-networking@pivotal.io"

    git add .
    git commit -m "Seeding CATS config, certs, etc. in '${BBL_STATE_DIR}'"
  fi
}

git clone "bbl-state" "updated-bbl-state"

output_path="updated-bbl-state/${BBL_STATE_DIR}"
mkdir -p "${output_path}"
pushd ${output_path}
    write_cats_config
    write_smoke_test_config
    write_apps_domains_vars_file
    write_load_balancer_certs

    commit_bbl_state
popd


