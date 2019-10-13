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
kubectl apply -f "${cfroutesync_dir}/crds/routebulksync.yaml"
kubectl apply -f "${cfroutesync_dir}/crds/route.yaml"

echo 'Deploying to Kubernetes'
pushd ~/workspace/eirini-dev-1-config
  kubectl -n cf-system delete secret cfroutesync || true
  kubectl -n cf-system create secret generic cfroutesync --from-file=ca --from-file=clientName --from-file=clientSecret --from-file=uaaBaseUrl --from-file=ccBaseUrl
popd
kubectl apply -f "${script_dir}/cfroutesync.yaml"

echo restarting...
kubectl delete pods -ncf-system -l app=cfroutesync
echo done

