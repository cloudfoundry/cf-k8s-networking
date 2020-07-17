#!/bin/bash

while true; do
  sleep 1
  configmaps=$(kubectl get configmaps -n istio-system -l "cloudfoundry.org/istio_version=$ISTIO_VERSION" --no-headers | wc -l)
  deposets=$(kubectl get daemonsets,pods,deployments -n istio-system -l "cloudfoundry.org/istio_version notin ($ISTIO_VERSION)" | wc -l)
  if [[ $configmaps == 2 && $deposets == 0 ]]; then
    break
  fi
  echo "Didn't quite find it this time... will try again in a sec"
done

kubectl -n cf-workloads rollout restart statefulset
kubectl -n cf-workloads delete jobs -l "cloudfoundry.org/istio_version notin ($ISTIO_VERSION)"
