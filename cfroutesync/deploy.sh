#!/usr/bin/env bash
set -euo pipefail

echo 'Building image'
img=$(docker build . -q)
echo 'Tagging and pushing image'
docker tag $img gcr.io/cf-routing-desserts/cfroutesync
docker push gcr.io/cf-routing-desserts/cfroutesync

echo 'Deploying to Kubernetes'
pushd ~/workspace/lagunabeach/cfroutesync-uaa
  kubectl -n pas-system delete secret cfroutesync-uaa || true
  kubectl -n pas-system create secret generic cfroutesync-uaa --from-file=ca --from-file=clientName --from-file=clientSecret --from-file uaaBaseUrl
popd
kubectl apply -f cfroutesync.yaml

echo restarting...
kubectl delete pods -l app=cfroutesync -npas-system
echo done

