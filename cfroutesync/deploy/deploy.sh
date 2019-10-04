#!/usr/bin/env bash
set -euo pipefail

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cfroutesync_dir="${script_dir}/.."

echo 'Building image'
img=$(docker build -q -f "${script_dir}/Dockerfile" "${cfroutesync_dir}")
echo 'Tagging and pushing image'
docker tag $img gcr.io/cf-routing-desserts/cfroutesync
docker push gcr.io/cf-routing-desserts/cfroutesync

echo 'Applying routebulksync and route custom resources'
kubectl apply -f "${cfroutesync_dir}/config/routebulksync.yaml"
kubectl apply -f "${cfroutesync_dir}/config/route.yaml"

echo 'Deploying to Kubernetes'
pushd ~/workspace/lagunabeach/cfroutesync-uaa
  kubectl -n pas-system delete secret cfroutesync-uaa || true
  kubectl -n pas-system create secret generic cfroutesync-uaa --from-file=ca --from-file=clientName --from-file=clientSecret --from-file uaaBaseUrl
popd
kubectl apply -f "${script_dir}/cfroutesync.yaml"

echo restarting...
kubectl delete pods -l app=cfroutesync -npas-system
echo done

