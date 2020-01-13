# Metrics

Istio and `cfroutesync` produce metrics that help an operator understand how the routing system is functioning.

These metrics are emitted with prometheus by default when you install Istio and CF-K8s-Networking.  See [indicators.yml](indicators.yml) for recommended `promql` queries to watch.

## Viewing metrics
You can view and explore Istio and `cfroutesync` metrics and queries via a Grafana or Prometheus dashboard. Prometheus comes installed by default with Istio, but for Grafana you will have to enable `--set values.grafana.enabled=true` on your Istio installation.

### Grafana
We use grafana to see these metrics.  To get a dashboard:

1. Install grafana.  This is optional with Istio.  We have a CI job for this ([script](../../ci/tasks/istio/deploy-istio.sh))

1. Install our dashboard.  We have a CI job ([task](../../ci/tasks/istio/install-grafana-dashboard.yml), [script](../../ci/tasks/istio/install-grafana-dashboard.sh)).

1. View the dashboard:

   ```bash
   istioctl dashboard grafana
   ```
   
   (this is a `kubectl port-forward` but simpler).
   
   
1. Go to "Indicator" dashboard.

### Prometheus
1. View the dashboard:
   ```bash
   istioctl dashboard prometheus
   ```
1. See the [indicators](indicators.yml) and paste in the promql queries to view graphs

## Dev workflow
To export:
1. Go to "Indicator" dashboard.
1. Click on "share" icon in the top right corner.
1. Select export and save file to [dashboard.json](./dashboard.json)
