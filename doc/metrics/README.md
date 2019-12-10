# Metrics

To be able to view all current metrics from `indicator.yml` you can import [dashboard.json](./dashboard.json) to Grafana.

```bash
# Enable Grafana

helm template ~/workspace/istio/install/kubernetes/helm/istio --name istio --namespace istio-system -f ~/workspace/cf-k8s-networking/install/istio-values.yaml -v grafna.enabled=true | kubectl apply -f -

# Port forward Grafana port

istioctl dashboard grafana
```


Import [dashboard.json](./dashboard.json) by Create (plus sign) -> Import.

