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

For example, to deploy an environment called `cf-for-k8s-dev-1`, run:

```bash
~/workspace/cf-k8s-networking/hack/cf4k8s/create-and-deploy.sh cf-for-k8s-dev-1
```

If you'd like to customize any of these options, you can use the `$CLUSTER_NAME` and `$CF_DOMAIN` or `$SHARED_DNS_ZONE_NAME` environment variables, but the defaults are probably what you want.

## Cleanup
To destroy the cluster and cf-for-k8s deployment, run:

```bash
~/workspace/cf-k8s-networking/hack/cf4k8s/destroy.sh cf-for-k8s-dev-1
```
