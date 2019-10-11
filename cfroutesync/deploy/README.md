# how to install this thing

assemble the secrets required to run `deploy.sh`

```
cd ~/workspace/networking-oss-deployments/environments/eirini-dev-1
eval "$(bbl print-env)"

mkdir -p ~/workspace/eirini-dev-1-config
cd ~/workspace/eirini-dev-1-config

echo "https://api.eirini-dev-1.routing.cf-app.com" > ccBaseUrl
echo "https://uaa.eirini-dev-1.routing.cf-app.com" > uaaBaseUrl
echo "network-policy" > clientName
credhub get -n /bosh-eirini-dev-1/cf/uaa_clients_network_policy_secret -j | jq -r .value > clientSecret
cat ~/workspace/oss-networking-deployments/environments/eirini-dev-1/bbl-state/bbl-state.json | jq -r .lb.cert > ca
```


also you'll need to [install metacontroller](https://metacontroller.app/guide/install/)
