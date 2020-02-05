#!/bin/bash

set -euo pipefail

if [[ -z $1 || -z $2 ]]; then
    echo "Usage: ./test_local.sh <test_config_path> <bbl_state>"
fi

test_config_path="$1"
bbl_state="$2"

# On networking program workstations, the kubeconfig
# should be stored in the $HOME directory
kubeconfig_path="${HOME}/.kube/config"


#cf_k8s_networking_dir =$(dirname)

# the extract_kubeconfig function takes a path to where the kubeconfig
# should be stored, and a path to the directory containing bbl-state.json
source "../ci/tasks/k8s/utils.sh"
extract_kubeconfig $kubeconfig_path $bbl_state


CONFIG="$test_config_path" ginkgo -v .
