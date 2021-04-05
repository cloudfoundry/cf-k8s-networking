#!/bin/bash

while true; do
  sleep 1
  deposets=$(kubectl get daemonsets,pods,deployments -n istio-system -l "cloudfoundry.org/istio_version notin ($ISTIO_VERSION)" | wc -l)
  if [[ $deposets == 0 ]]; then
    break
  fi
  echo "Didn't quite find it this time... will try again in a sec"
done

kubectl -n cf-workloads rollout restart statefulsets
kubectl -n cf-workloads delete jobs -l "cloudfoundry.org/istio_version notin ($ISTIO_VERSION)"

kubectl -n cf-system rollout restart daemonsets/fluentd
kubectl -n cf-system rollout status daemonsets/fluentd
