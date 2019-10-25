#!/bin/bash

set -eu

# ENV
: "${CLOUD_CONFIG_PATH:?}"

# INPUTS
pushd "bbl-state/${BBL_STATE_DIR}"
  eval "$(bbl print-env)"
popd

cloud_config="cloud-config/${CLOUD_CONFIG_PATH}"

bosh -n update-cloud-config "${cloud_config}"
