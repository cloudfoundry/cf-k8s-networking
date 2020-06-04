# Managing Certificates for Application Ingress

For each domain in CF:

1. Create a secret with the certificate and the key in `istio-system` namespace:
   _Note: The secret name should not start with `istio` or `prometheus` for
   dynamic reloading of the secret._

```
kubectl create secret tls wildcard-apps-example-com-cert \
    -n istio-system \
    --cert='path/to/wildcard-apps.example.com.crt' \
    --key='path/to/wildcard-apps.example.com.key'
```

_Note: This can be either a wildcard cert or a cert for a single domain._

2. Add a new entry under `spec.servers` to the [Istio
   `Gateway`](https://github.com/cloudfoundry/cf-for-k8s/blob/21209bbfcadf626a81bc19a8320050b98076f25e/config/gateway.lib.yml).
   This is in `cf-system` namespace by default. So, to add configuration for
   `'*.apps.example.com'`, you would need to add this entry to the
   `spec.servers` array. Here is an example [`ytt`](https://get-ytt.io/) overlay
   that adds it:

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

_NOTE: There is a known issue where the System Domain certificates won't
automatically reload in the gateway because the secrets are prefixed with
`istio`. Istio doesn't hot reload certificates from secrets prefixed with
`istio. The gateway can be updated to use a differently named secret for the
system domain.

To update the application domain or system domain certificates for Cloud Foundry
the Kubernetes Secret storing the certificate can be updated to reflect the new
certificate. Istio uses
[SDS](https://istio.io/docs/tasks/traffic-management/ingress/secure-ingress-sds/)
for the Ingress Gateway. SDS allows the Ingress Gateway to reload certificates
without restarting Envoy. When the secret is updated, Istio will reload the
certificate automatically.

For example, in the case of creating your own `wildcard-apps-example-com-cert`
(as shown above), you can rotate the certificate by recreating the secret with
your new certificate and key and the gateway should reload it automatically.

```
kubectl delete secret tls wildcard-apps-example-com-cert

kubectl create secret tls wildcard-apps-example-com-cert \
    -n istio-system \
    --cert='path/to/wildcard-apps.example.com.crt' \
    --key='path/to/wildcard-apps.example.com.key'
```

When using [CF-for-K8s](https://github.com/cloudfoundry/cf-for-k8s) and you want
to rotate the default application domain certificate or system domain
certificate, you can do the following.

In your `cf-values.yaml`, the system domain certificate is called
`system_certificate` and the application domain certificate is called
`workloads_certificate`. These two properties can be independently updated and a
redeploy using `ytt` and `kapp` will cause the secrets to be recreated and
successfully rotated in the Istio Ingress Gateway.
