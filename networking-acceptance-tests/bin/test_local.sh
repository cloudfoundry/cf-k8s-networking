#!/bin/bash

set -euo pipefail

if [[ -z $1 ]]; then
    echo "Usage: ./test_local.sh <test_config_path>"
fi

test_config_path="$1"

CONFIG="$test_config_path" ginkgo -v .
