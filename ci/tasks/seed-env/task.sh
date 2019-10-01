#!/bin/bash

set -e -u -o pipefail

: "${BBL_STATE_DIR:?}"
: "${DOMAIN:?}"

lb_output_path="certs/load-balancer"

function write_load_balancer_certs() {
  if [ ! -d "${lb_output_path}" ]; then
    mkdir -p "${lb_output_path}"
    pushd "${lb_output_path}"
      local cert_cn
      cert_cn="*.${DOMAIN}"
      certstrap --depot-path "." init --passphrase '' --common-name "server-ca"
      certstrap --depot-path "." request-cert --passphrase '' --common-name "${cert_cn}" --csr "$DOMAIN.csr" --key "${DOMAIN}.key"
      certstrap --depot-path "." sign --CA "server-ca" "${DOMAIN}"

      mv "${DOMAIN}.csr" "server.csr"
      mv "${DOMAIN}.crt" "server.crt"
      mv "${DOMAIN}.key" "server.key"
    popd
  fi
}

function commit_bbl_state() {
  if [[ -n $(git status --porcelain) ]]; then
    git config user.name "CI Bot"
    git config user.email "cf-routing-eng@pivotal.io"

    git add .
    git commit -m "Create LB certs for '${BBL_STATE_DIR}'"
  fi
}

git clone "bbl-state" "updated-bbl-state"

mkdir -p "updated-bbl-state/${BBL_STATE_DIR}"
pushd "updated-bbl-state/${BBL_STATE_DIR}"
	write_load_balancer_certs

	commit_bbl_state
popd


