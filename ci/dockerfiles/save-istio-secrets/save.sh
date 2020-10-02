#!/bin/bash


main() {
  secret_namespace="${1}"
  echo "$@"
  if [ -z "${secret_namespace}" ]; then
    echo "secret_namespace is required"
    echo "Usage: ./save.sh <secret_namespace>"
    exit 1
  fi

    if [ -n "${DEBUG}" ]; then
      cat /etc/istio-certs/*
    fi

    if [[ -s "/etc/istio-certs/root-cert.pem" && -s "/etc/istio-certs/cert-chain.pem" && -s "/etc/istio-certs/key.pem" ]]; then
      echo "Secrets are in-place"
      kubectl create secret tls istio-client-certs --cert=/etc/istio-certs/cert-chain.pem --key=/etc/istio-certs/key.pem --namespace "${secret_namespace}" || echo "probably exists"
      kubectl create secret generic istio-ca-cert --from-file=ca.crt=/etc/istio-certs/root-cert --namespace "${secret_namespace}" || echo "probably exists"
    else
      echo "no secrets were found, terminating"
      exit 1
    fi
}

main "$@"
