## Prerequisites
* gcloud installed and authenticated
* kubectl
* kapp
* ytt
* bosh cli
* cf cli
* yq (the one from pip, not the one from brew)

## Create a Development CF-for-K8s environment
You can use the `create-and-deploy.sh` script to create a cluster on GKE and deploy cf-for-k8s.

The required parameters are `$CLUSTER_NAME` and `$CF_DOMAIN`. You can optionally provide `$SHARED_DNS_ZONE_NAME`, but it defaults to `routing-lol`

For example, to deploy an environment called `cf-for-k8s-dev-1`, run:

```bash
CLUSTER_NAME=cf-for-k8s-dev-1 CF_DOMAIN=cf-for-k8s-dev-1.routing.lol ~/workspace/cf-k8s-networking/hack/cf4k8s/create-and-deploy.sh
```

## Cleanup
To destroy the cluster and cf-for-k8s deployment, run:

```bash
CLUSTER_NAME=cf-for-k8s-dev-1 CF_DOMAIN=cf-for-k8s-dev-1.routing.lol ~/workspace/cf-k8s-networking/hack/cf4k8s/destroy.sh
```
