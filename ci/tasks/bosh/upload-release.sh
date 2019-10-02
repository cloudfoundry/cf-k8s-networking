#!/bin/bash
set -eu

function setup_bosh_env_vars() {
  pushd "bbl-state/${BBL_STATE_DIR}"
    eval "$(bbl print-env)"
  popd
}

function upload_release() {
  for filename in release-tarball/*.tgz; do
    bosh upload-release --sha2 "$filename"
  done
}

function main() {
  setup_bosh_env_vars
  upload_release
}

main
