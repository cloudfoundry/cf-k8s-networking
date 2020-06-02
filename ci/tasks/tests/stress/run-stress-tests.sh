#!/usr/bin/env bash

set -euox pipefail

concourse-dcind/entrypoint.sh cf-k8s-networking/routecontroller/scripts/stress

shopt -s dotglob
pushd routecontroller-stress-results
    git add .
    git commit -m "Stress test results"
popd

cp -r routecontroller-stress-results/* routecontroller-stress-results-modified

