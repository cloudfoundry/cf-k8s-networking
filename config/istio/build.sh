#!/usr/bin/env bash
set -eu

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

${SCRIPT_DIR}/generate.sh "$@" | kbld -f - > "${SCRIPT_DIR}/generated/xxx-generated-istio.yaml"
