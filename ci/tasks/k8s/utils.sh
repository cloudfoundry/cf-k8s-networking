#!/bin/bash

set -euo pipefail


function extract_kubeconfig() {
  if [[ -z $1 || -z $2 ]]; then
      echo "Usage: extract_kubeconfig <kubeconfig_path> <bbl_state>"
  fi

  kubeconfig_path="$1"
  bbl_state="$2"

  export KUBECONFIG="$kubeconfig_path"

  tmp_dir="$(mktemp -d /tmp/kubernetes-certs.XXXXXXXX)"

  pushd "$bbl_state" > /dev/null
    bosh_director_name="$(bbl outputs | bosh int - --path=/director_name)"
    k8s_host_url="$(bbl outputs | bosh int - --path=/k8s_host_url)"
    k8s_service_username="$(bbl outputs | bosh int - --path=/k8s_service_username)"
    k8s_service_token="$(bbl outputs | bosh int - --path=/k8s_service_account_data/token)"

    k8s_ca_path="${tmp_dir}/k8s-ca"
    bbl outputs | bosh int - --path=/k8s_ca > $k8s_ca_path

    cluster_name="${bosh_director_name}-cluster"

    echo "Configuring kubectl"
    kubectl config set-credentials cf-k8s-networking-installer --user=$k8s_service_username --token=$k8s_service_token
    kubectl config set-cluster $cluster_name --embed-certs=true --server=$k8s_host_url --certificate-authority=$k8s_ca_path
    kubectl config set-context $cluster_name --cluster=$cluster_name --user=cf-k8s-networking-installer

    echo "Testing kubeconfig"
    kubectl config use-context $cluster_name
    kubectl get pods --all-namespaces
  popd

}
