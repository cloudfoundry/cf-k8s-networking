#!/usr/bin/env bash
set -eu

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

${SCRIPT_DIR}/generate.sh "$@"  > "${SCRIPT_DIR}/../istio-generated/xxx-generated-istio.yaml"
