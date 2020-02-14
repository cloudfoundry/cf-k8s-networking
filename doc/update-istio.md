## Updating Istio

### Generating a new Istio installation

To generate the YAML to install Istio you will need to have the [`istioctl`](https://github.com/istio/istio/releases) CLI for the version of Istio you want to generate YAML for.

This script will generate the file [config/istio-generated/xxx-generated-istio.yaml](config/istio-generated/xxx-generated-istio.yaml).

```bash
config/istio/build.sh
```

### Install Istio

To install Istio, you can apply the YAML in [config/istio-generated](config/istio-generated):

```bash
kubectl apply -f config/istio-generated/
```

