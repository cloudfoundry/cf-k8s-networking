# Managing Certificates for Application Ingress

For each domain in CF:

1. Create a secret with the certificate and the key in `istio-system` namespace:
_Note: The secret name should not start with `istio` or `prometheus` for dynamic reloading of the secret._
```
kubectl create secret tls wildcard-apps-example-com-cert \
    -n istio-system \
    --cert='path/to/wildcard-apps.example.com.crt' \
    --key='path/to/wildcard-apps.example.com.key'
```

_Note: This can be either a wildcard cert or a cert for a single domain._

2. Add a new entry under `spec.servers` to the [Istio
   `Gateway`](https://istio.io/docs/reference/config/networking/gateway/) object
   in the. This is `cf-workloads` by default. So, to add configuration for
   `'*.apps.example.com'` you would need to add this entry to the `spec.servers`
   array. Here is an example [`ytt`](https://get-ytt.io/) overlay that adds it:

```yaml
#@ load("@ytt:overlay", "overlay")

#@overlay/match by=overlay.and_op(overlay.subset({"metadata":{"name":"istio-ingressgateway"}}), overlay.subset({"kind": "Gateway"}))
---
spec:
  servers:
  #@overlay/append
  - hosts:
    - '*.apps.example.com'
    port:
      name: https-apps-example-com
      number: 443
      protocol: HTTPS
    tls:
      credentialName: wildcard-apps-example-com-cert
      mode: SIMPLE
```

This overlay can be applied during installation of
[cf-for-k8s](https://github.com/cloudfoundry/cf-for-k8s).

### Rotation:

- Istio uses [SDS](https://istio.io/docs/tasks/traffic-management/ingress/secure-ingress-sds/) for the Ingress Gateway. SDS allows the Ingress Gateway to reload certificates without restarting Envoy. When the secret is updated, Istio will reload the certificate automatically.
