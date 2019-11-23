# example: routing to system components

make CA and wildcard cert
```
./regenerate.sh
```

make the wildcard cert available to the istio ingress gateway
```
kubectl create -n istio-system secret tls istio-ingressgateway-certs --cert sys-wildcard.crt --key sys-wildcard.key
```

install your system component
```
kubectl apply -f example-sys-component.yaml
```

test it out
```
curl -v --cacert sys-ca.crt https://some-cf-component.sys.eirini-dev-1.routing.lol
```
