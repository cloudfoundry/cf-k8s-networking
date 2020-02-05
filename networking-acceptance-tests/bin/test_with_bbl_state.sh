#!/bin/bash

set -euo pipefail

set +u
if [[ -z $1 || -z $2 ]]; then
  echo "Usage: ./test_with_bbl_state.sh <test_config_path> <bbl_state> [kube_config_path]"
  exit 1
fi
set -u

test_config_path="$1"
bbl_state="$2"
kube_config_path=${3:="${HOME}/.kube/config"}

# the extract_kubeconfig function takes a path to where the kubeconfig
# should be stored, and a path to the directory containing bbl-state.json
source "../ci/tasks/k8s/utils.sh"
extract_kubeconfig_from_bbl_state "${kube_config_path}" "$bbl_state"

./bin/test_local.sh "${test_config_path}" "${kube_config_path}"
