# custom bosh-lite director override taking pieces from the following:
# * https://github.com/cloudfoundry/bosh-bootloader/blob/b4cc74ff6134248055b6d41b6bd6aac4fa5663f3/plan-patches/bosh-lite-gcp/create-director-override.sh
# * https://github.com/cloudfoundry-community/eirini-bosh-release/blob/5d0a14262bd5b74dfc4ee51f6be2859d4dd56a53/plan-patches/shared/create-director-override.sh

#!/bin/sh
bosh create-env \
  ${BBL_STATE_DIR}/bosh-deployment/bosh.yml \
  --state  ${BBL_STATE_DIR}/vars/bosh-state.json \
  --vars-store  ${BBL_STATE_DIR}/vars/director-vars-store.yml \
  --vars-file  ${BBL_STATE_DIR}/vars/director-vars-file.yml \
  --var-file gcp_credentials_json="${BBL_GCP_SERVICE_ACCOUNT_KEY_PATH}" \
  -v project_id="${BBL_GCP_PROJECT_ID}" \
  -v zone="${BBL_GCP_ZONE}" \
  -o  ${BBL_STATE_DIR}/bosh-deployment/gcp/cpi.yml \
  -o  ${BBL_STATE_DIR}/bosh-deployment/jumpbox-user.yml \
  -o  ${BBL_STATE_DIR}/bosh-deployment/uaa.yml \
  -o  ${BBL_STATE_DIR}/bosh-deployment/credhub.yml \
  -o  ${BBL_STATE_DIR}/bosh-deployment/bosh-lite.yml \
  -o  ${BBL_STATE_DIR}/bosh-deployment/bosh-lite-runc.yml \
  -o  ${BBL_STATE_DIR}/bosh-deployment/gcp/bosh-lite-vm-type.yml \
  -o  ${BBL_STATE_DIR}/external-ip-gcp.yml \
  -o  ${BBL_STATE_DIR}/ip-forwarding.yml

bosh_director_name="$(bbl outputs | bosh int - --path=/director_name)"
k8s_host_url="$(bbl outputs | bosh int - --path=/k8s_host_url)"
k8s_service_username="$(bbl outputs | bosh int - --path=/k8s_service_username)"
k8s_service_token="$(bbl outputs | bosh int - --path=/k8s_service_account_data/token)"
k8s_ca="$(bbl outputs | bosh int - --path=/k8s_ca)"

eval "$(bbl print-env -s ${BBL_STATE_DIR})"
credhub set --name=/${bosh_director_name}/cf/k8s_host_url --value="${k8s_host_url}" -t value
credhub set --name=/${bosh_director_name}/cf/k8s_service_username --value="${k8s_service_username}" -t value
credhub set --name=/${bosh_director_name}/cf/k8s_service_token --value="${k8s_service_token}" -t value
credhub set --name=/${bosh_director_name}/cf/k8s_node_ca --value="${k8s_ca}" -t value
