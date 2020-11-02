# 18. Create a Policy Server to Manage Network Policy

Date: 2020-10-27

## Status

Accepted

## Context

Well implemented service oriented apps typically include backend services who
never serve requests from users or clients outside the foundation.

Currently, for an app to reach a backend service, the backend service must
expose itself through the ingress gateway, and the app must hairpin through the
ingress gateway to reach it. This is a security concern, backend apps should not
be accessible outside the foundation.

CF for VMs provides an API mechanism for configuring which apps are permitted to
communicate with which other apps, called [Network
Policy](https://docs.cloudfoundry.org/devguide/deploy-apps/cf-networking.html#create-policies).
The job that provides this API is called policy-server, it has its own database
and API endpoint that the CLI communicates with.

Kubernetes provides the [NetworkPolicy Resource](https://kubernetes.io/docs/concepts/services-networking/network-policies/)
which serves similar outcomes.

Our objective is to create Kubernetes NetworkPolicy from CF Network Policy.

For more background, see the [exploration
document](https://docs.google.com/document/d/1qAYy737uB7orT8St56wg5MbnlrbuA9NoJib8dv4mNK0/edit#)

## Decision

We will write a new stateless component which implements the existing CF Network
Policy API endpoint. It will read and write Kubernetes NetworkPolicy directly to
the API server and use it as its store, rather than maintaining its own
database.

For specific implementation details, example commands, and example resources,
see [the exploration
document](https://docs.google.com/document/d/1qAYy737uB7orT8St56wg5MbnlrbuA9NoJib8dv4mNK0/edit#heading=h.pb0he04m6fbf).

## Consequences

- The API Server will be the source of truth for CF Network Policy, effectively
  making CF Network Policy and Kubernetes NetworkPolicy one and the same.
- As this uses entirely built-in Kubernetes resources, we do not add any
  external dependencies.
- This component may also need to write or modify Sidecar resources on
  foundations using Istio sidecars, as sidecars are not currently configured
  for app to app communication.
- There will be no path to import a cf-for-vms policy-server database into
  cf-for-k8s
- Our new component will need to conform to the new observability initiative
  (incl the distributed tracing work) that CAPI is doing so that this configuration
  can be traced as well.
