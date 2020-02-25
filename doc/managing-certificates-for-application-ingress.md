# Managing certificates for application ingress

For each domain in CF:

0. [Make sure SDS is enabled for Istio](https://istio.io/docs/tasks/traffic-management/ingress/secure-ingress-sds/#configure-a-tls-ingress-gateway-using-sds)

1. Create a secret with the certificate and the key in `istio-system` namespace:
```
kubectl create secret tls wildcard-apps-example-com-cert \
    -n istio-system \
    --cert='path/to/wildcard-apps.example.com.crt' \
    --key='path/to/wildcard-apps.example.com.key'
```

_Note: This can be either a wildcard cert or a cert for a single domain._

2. Add new server in the edge Istio gateway object in cf-workloads, e.g.
```
- hosts:
  # the host from your certificate
  - '*.apps.example.com'
  port:
    name: https-apps-example-com
    number: 443
    protocol: HTTPS
  tls:
    # secret created in first step
    credentialName: wildcard-apps-example-com-cert
    mode: SIMPLE
```

### Rotation:

- SDS will swap certs automatically when the secret is updated.
