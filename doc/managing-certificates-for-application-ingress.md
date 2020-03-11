# Managing Certificates for Application Ingress

For each domain in CF:

1. Create a secret with the certificate and the key in `istio-system` namespace:
```
kubectl create secret tls wildcard-apps-example-com-cert \
    -n istio-system \
    --cert='path/to/wildcard-apps.example.com.crt' \
    --key='path/to/wildcard-apps.example.com.key'
```

_Note: This can be either a wildcard cert or a cert for a single domain._

2. Add a new entry under `spec.servers` to the [Istio `Gateway`](https://istio.io/docs/reference/config/networking/gateway/) object in the namespace that your apps are running in. This is `cf-workloads` by default. So, to add configuration for `'*.apps.example.com'` you would need to add this entry to the `spec.servers` array:
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
