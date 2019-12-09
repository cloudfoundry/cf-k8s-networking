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
image_repo=gcr.io/cf-networking-images/cf-k8s-networking/cfroutesync:${environment}

docker tag "${img}" "${image_repo}"
docker push "${image_repo}"

values_file="$(mktemp -u).yml" # for ytt it's important the file to have .yml extension
"${cf_k8s_networking_dir}/install/scripts/generate_values.rb" "${HOME}/workspace/networking-oss-deployments/environments/${environment}/bbl-state.json" > ${values_file}

echo 'Deploying to Kubernetes...'

ytt -f "${cf_k8s_networking_dir}/install/ytt/networking/" -f "${values_file}" \
    --data-value-yaml cfroutesync.image="${image_repo}" | \
    kapp deploy -n "cf-system" -a cfroutesync \
    -f "${cf_k8s_networking_dir}/cfroutesync/crds/routebulksync.yaml" \
    -f - \
    -y

echo "Done"
