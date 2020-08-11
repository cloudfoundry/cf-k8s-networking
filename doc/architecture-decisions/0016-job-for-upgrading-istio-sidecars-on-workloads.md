# 16. Job for Upgrading Istio Sidecars on Workloads

Date: 2020-08-11

## Status

Accepted

## Context

Istio's service mesh capabilites are facilitated via sidecars injected into
workload pods. These sidecars run an Istio-patched version of Envoy that is tied
to the version of Istio that injects them.

Typically when new versions of Istio are released, new versions of the sidecars
are released as well. Istio has been good so far about supporting older versions
of sidecars that were deployed before Istio was upgraded, but it is still
[documented best practice](https://istio.io/latest/docs/setup/upgrade/) to roll
all the pods after an Istio upgrade.

As an additional constraint, the operators of cf-for-k8s clusters expect to be
able to perform upgrades in one `kapp deploy`, with no post-install hooks or
other bash scripts. This limits our options considerably. See this [Slack
thread](https://cloudfoundry.slack.com/archives/CH9LF6V1P/p1592521879117400) on
that constraint.

## Decision

We will use the kubernetes
[Job](https://kubernetes.io/docs/concepts/workloads/controllers/job/) resource
to run the kubectl command needed to roll workload pods, after waiting for the
new Istio control plane to be up and healthy.

To that end, we will add the necessary minimal `ServiceAccounts` and `Roles`
needed to list resources in the `istio-system` namespace, and restart resources
in the configured workload namespace. We will also build and maintain a
container image that contains the Job's logic.

All istio components will be tagged with their Istio version so that the job can
positively determine that the correct version of control plane components are
alive and healthy. We will also name the job according to it's Istio version, so
that we can take advantage of `Jobs` inherent immutability in cases where a
cf-for-k8s upgrade does not contain a new Istio version (pushing the same job
again will not cause it to rerun, preventing workloads from rolling
unnecessarily). Subsequent jobs will clean up previous ones.

## Consequences

* Apps will always have the current version of Istio's side car
* Apps deployed with a single instance will experience downtime during upgrades
    * This may break some uptime testing that other teams are doing, but
      deploying 2 instances should fix them without requiring significant
      additional resources
* A completed job will hang around in the configured workload namespace, but
  only platform operators will see that
