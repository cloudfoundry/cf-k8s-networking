#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 1 ]; then
  echo 'Usage ./deploy [ENVIRONMENT_NAME]'
  exit 1
fi
script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cfroutesync_dir="${script_dir}/.."
cf_k8s_networking_dir="${cfroutesync_dir}/.."
environment="$1"

echo 'Fetching environment variables for credhub...'
pushd "${HOME}/workspace/networking-oss-deployments/environments/${environment}/"
  eval "$(bbl print-env)"
popd

echo 'Building image...'
img=$(docker build -q -f "${script_dir}/Dockerfile" "${cfroutesync_dir}")
echo 'Tagging and pushing image'
docker tag $img gcr.io/cf-routing/cf-k8s-networking/cfroutesync:${environment}
docker push gcr.io/cf-routing/cf-k8s-networking/cfroutesync:${environment}

echo 'Applying routebulksync CRD...'
kubectl apply -f "${cfroutesync_dir}/crds/routebulksync.yaml"

echo 'Deploying to Kubernetes...'
helm template "${cf_k8s_networking_dir}/install/helm/networking/" --values <("${cf_k8s_networking_dir}/install/scripts/generate_values.rb" "${HOME}/workspace/networking-oss-deployments/environments/${environment}/bbl-state.json") | kubectl apply -f-

echo restarting...
kubectl delete pods -ncf-system -l app=cfroutesync
echo done

