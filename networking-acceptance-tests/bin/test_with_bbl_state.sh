#!/bin/bash

set -euo pipefail


set +u
if [[ -z $1 || -z $2 ]]; then
  echo "Usage: ./test_with_bbl_state.sh <test_config_path> <bbl_state>"
  exit 1
fi
set -u

test_config_path="$1"
bbl_state="$2"

# On networking program workstations, the kubeconfig
# should be stored in the $HOME directory
kubeconfig_path="${HOME}/.kube/config"


# the extract_kubeconfig function takes a path to where the kubeconfig
# should be stored, and a path to the directory containing bbl-state.json
source "../ci/tasks/k8s/utils.sh"
extract_kubeconfig_from_bbl_state "$kubeconfig_path" "$bbl_state"

./bin/test_local.sh "$test_config_path"
