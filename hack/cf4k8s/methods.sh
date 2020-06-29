#!/usr/bin/env bash

function latest_cluster_version() {
  gcloud container get-server-config --zone us-west1-a --project ${GCP_PROJECT} 2>/dev/null | yq .validMasterVersions[0] -r
}

function credhub_get_gcp_service_account_key() {
  source ~/workspace/networking-oss-deployments/scripts/script_helpers.sh
  concourse_credhub_login
  credhub get -n /concourse/cf-k8s/gcp_gcr_service_account_key -j | jq -r .value > /tmp/cf-k8s-networking-service-account-key.json
  export GCP_SERVICE_ACCOUNT_KEY=/tmp/cf-k8s-networking-service-account-key.json
}

function create_and_target_huge_cluster() {
    if gcloud container clusters describe ${CLUSTER_NAME} --project ${GCP_PROJECT} --zone us-west1-a > /dev/null; then
        echo "${CLUSTER_NAME} already exists! Continuing..."
    else
        echo "Creating cluster: ${CLUSTER_NAME} ..."
        gcloud container clusters create ${CLUSTER_NAME} \
          --project ${GCP_PROJECT} \
          --cluster-version=$(latest_cluster_version) \
          --zone us-west1-a \
          --machine-type=n1-standard-8 \
          --enable-network-policy \
          --labels team=cf-k8s-networking \
          --enable-ip-alias \
          --num-nodes 100
    fi
    gcloud container clusters get-credentials --project ${GCP_PROJECT} ${CLUSTER_NAME} --zone us-west1-a
}

function create_and_target_cluster() {
    if gcloud container clusters describe ${CLUSTER_NAME} --project ${GCP_PROJECT} --zone us-west1-a > /dev/null; then
        echo "${CLUSTER_NAME} already exists! Continuing..."
    else
        echo "Creating cluster: ${CLUSTER_NAME} ..."
        gcloud container clusters create ${CLUSTER_NAME} \
          --project ${GCP_PROJECT} \
          --zone us-west1-a \
          --machine-type=n1-standard-4 \
          --num-nodes 5 \
          --enable-network-policy \
          --labels team=cf-k8s-networking \
          --cluster-version=$(latest_cluster_version)
    fi
    gcloud container clusters get-credentials --project ${GCP_PROJECT} ${CLUSTER_NAME} --zone us-west1-a
}

function deploy_cf_for_k8s() {
    clone_if_not_exist https://github.com/cloudfoundry/cf-for-k8s.git "${HOME}/workspace/cf-for-k8s"
    pushd "${HOME}/workspace/cf-for-k8s" > /dev/null
        mkdir -p "/tmp/${CLUSTER_NAME}"
        if [ ! -f "/tmp/${CLUSTER_NAME}/cf-values.yml" ]; then
          ./hack/generate-values.sh -d ${CF_DOMAIN} -g "${GCP_SERVICE_ACCOUNT_KEY}" > "/tmp/${CLUSTER_NAME}/cf-values.yml"
        fi
        kapp deploy -a cf -f <(ytt -f config -f "/tmp/${CLUSTER_NAME}/cf-values.yml") -y
    popd
}

function target_cf() {
    echo "Targeting CF!"
    cf api --skip-ssl-validation "https://api.${CF_DOMAIN}"
    cf auth admin "$(cat "/tmp/${CLUSTER_NAME}/cf-values.yml" | grep cf_admin_password | awk '{print $2}')"
    cf create-org o
    cf create-space -o o s
    cf target -o o -s s
    echo "Successfully targeted CF!"
}

function enable_docker() {
    cf enable-feature-flag diego_docker
}

clone_if_not_exist() {
  local remote=$1
  local dst_dir="$2"
  local branch_name="${3:-master}"
  echo "Cloning $remote into $dst_dir"
  if [[ ! -d $dst_dir ]]; then
    if [[ -n $branch_name ]]
      then
        git clone --branch "$branch_name" "$remote" "$dst_dir"
      else
        git clone "$remote" "$dst_dir"
    fi
  fi
}

function configure_dns() {
  echo "Discovering Istio Gateway LB IP"
  external_static_ip=""
  while [ -z $external_static_ip ]; do
      sleep 1
      external_static_ip=$(kubectl get services/istio-ingressgateway -n istio-system --output="jsonpath={.status.loadBalancer.ingress[0].ip}")
  done

  echo "Configuring DNS for external IP: ${external_static_ip}"
  gcloud dns record-sets transaction start --project ${GCP_PROJECT} --zone="${SHARED_DNS_ZONE_NAME}"
  gcp_records_json="$( gcloud dns record-sets list --project ${GCP_PROJECT} --zone "${SHARED_DNS_ZONE_NAME}" --name "*.${CF_DOMAIN}" --format=json )"
  record_count="$( echo "${gcp_records_json}" | jq 'length' )"
  if [ "${record_count}" != "0" ]; then
    existing_record_ip="$( echo "${gcp_records_json}" | jq -r '.[0].rrdatas | join(" ")' )"
    gcloud dns record-sets transaction remove --project ${GCP_PROJECT} --name "*.${CF_DOMAIN}" --type=A --zone="${SHARED_DNS_ZONE_NAME}" --ttl=300 "${existing_record_ip}" --verbosity=debug
  fi
  gcloud dns record-sets transaction add --project ${GCP_PROJECT} --name "*.${CF_DOMAIN}" --type=A --zone="${SHARED_DNS_ZONE_NAME}" --ttl=300 "${external_static_ip}" --verbosity=debug

  echo "Contents of transaction.yaml:"
  cat transaction.yaml
  gcloud dns record-sets transaction execute --project ${GCP_PROJECT} --zone="${SHARED_DNS_ZONE_NAME}" --verbosity=debug

  resolved_ip=''
  set +o pipefail
  sleep_time=5
  while [ "$resolved_ip" != "$external_static_ip" ]; do
    echo "Waiting $sleep_time seconds for DNS to propagate..."
    sleep $sleep_time
    resolved_ip=$(nslookup "*.${CF_DOMAIN}" | (grep ${external_static_ip} || true) | cut -d ' ' -f2)
    echo "Resolved IP: ${resolved_ip}, Actual IP: ${external_static_ip}"
    sleep_time=$(($sleep_time + 5))
  done
  set -o pipefail
  echo "We did it! DNS propagated! 🥳"
}
