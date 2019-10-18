#!/bin/bash

set -euo pipefail

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cfroutesync_dir="${script_dir}/.."

usage="Usage: $0 {local|docker|docker-shell}"

image="gcr.io/cf-routing/cf-k8s-networking/cfroutesync-integration-test-env"

if [ $# -ne 1 ]; then
  echo 1>&2 "$usage"
  exit 3
fi

if [ "$1" = "local" ]; then
  cd "$script_dir"
  ginkgo .
elif [ "$1" = "docker" ]; then
  docker run --rm \
    -v "$cfroutesync_dir":/cfroutesync \
    --workdir /cfroutesync \
    $image \
    /bin/bash ./integration/run-tests.sh local
elif [ "$1" = "docker-shell" ]; then
  docker run --rm \
    -it \
    -v "$cfroutesync_dir":/cfroutesync \
    --workdir /cfroutesync \
    $image \
    /bin/bash
else
  echo 1>&2 "$usage"
  exit 3
fi
