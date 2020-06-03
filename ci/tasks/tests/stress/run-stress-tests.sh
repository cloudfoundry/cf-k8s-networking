#!/usr/bin/env bash

set -euox pipefail

echo "Starting stress tests..."

cp routecontroller-stress-results/results.json cf-k8s-networking/routecontroller/stress/

concourse-dcind/entrypoint.sh cf-k8s-networking/routecontroller/scripts/stress

cp cf-k8s-networking/routecontroller/stress/results.json routecontroller-stress-results/results.json

pushd cf-k8s-networking > /dev/null
    git_sha="$(cat .git/ref)"
popd

pushd routecontroller-stress-results
    git config user.name "${GIT_COMMIT_USERNAME}"
    git config user.email "${GIT_COMMIT_EMAIL}"
    git add .
    git commit -m "Stress test results for cf-k8s-networking commit SHA ${git_sha}"
popd

shopt -s dotglob
cp -r routecontroller-stress-results/* routecontroller-stress-results-modified

