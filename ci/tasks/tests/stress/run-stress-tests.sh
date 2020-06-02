#!/usr/bin/env bash

set -euox pipefail

echo "Starting stress tests..."

cp routecontroller-stress-results/results.json cf-k8s-networking/routecontroller/stress/

concourse-dcind/entrypoint.sh cf-k8s-networking/routecontroller/scripts/stress

git config user.name "${GIT_COMMIT_USERNAME}"
git config user.email "${GIT_COMMIT_EMAIL}"

shopt -s dotglob
pushd routecontroller-stress-results
    git add .
    git commit -m "Stress test results"
popd

cp  cf-k8s-networking/routecontroller/stress/ routecontroller-stress-results/results.json
cp -r routecontroller-stress-results/* routecontroller-stress-results-modified

