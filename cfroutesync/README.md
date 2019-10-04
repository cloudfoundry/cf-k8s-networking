# cf route sync

manual testing

```
mkdir -p /tmp/uaa-config

echo "network-policy" > /tmp/uaa-config/clientName
echo "https://uaa.dev-full-2.routing.cf-app.com" > /tmp/uaa-config/uaaBaseUrl
credhub get -n /bosh-dev-full-2/cf/uaa_clients_network_policy_secret -j | jq -r .value > /tmp/uaa-config/clientSecret
pushd ~/workspace/deployments-routing/dev-full-2/bbl-state
  jq -r .lb.cert < bbl-state.json > /tmp/uaa-config/ca
popd

go run main.go -c /tmp/uaa-config > /tmp/bearer

curl -v --cacert /tmp/uaa-config/ca -H "Authorization: bearer $(cat /tmp/bearer)" https://api.dev-full-2.routing.cf-app.com/v3/apps
```
