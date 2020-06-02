# 13. Stopped App Routing

Date: 2020-05-29

## Status

Proposal

## Context

In CF-for-VMs, when a route is mapped to multiple apps, some started and some
stopped, gorouter only directs requests to the apps that are started.

In CF-for-k8s _currently_, we balance traffic equally between the started and
stopped apps. This results in 503 errors for the user, and is not feature
pairity.

This is because the components that create the Istio configuration do not
currently know about the status of apps. They assume all destinations on a Route
object are up and able to receive traffic.

See this issue: https://github.com/cloudfoundry/cf-for-k8s/issues/158

## Proposal

We have two proposals:

1. Cloud Controller only adds destinations to Routes that are running. This way
   Route objects remain the single source of truth for routing in the platform.
   * _Note:_ Routes with _no_ destinations are explicitly supported already.

2. Cloud Controller writes an App CR (this is a new CRD) that contains the
   application's status, so that the routecontroller can use that to decide how
   to configure Istio.

## Alternatives Considered

1. Bind a reconciler to Pod or StatefulSet create/update/delete events. This is
   problematic because delete events will require a full reconciliation of all
   services and virtual services. A full reconciliation is required because on a
   delete there isn't a way to retrieve the App GUID on a non-existent object.
   It will also create a lot of no-op reconcile events because pods can cause
   create/delete events for any number of reasons outside of just
   starting/stopping an app.
2. Attach a finalizer to every Pod or StatefulSet in addition to binding a
   reconciler. This eliminates the full reconciliation of above, but still
   causes a lot of no-op reconcile events for the same reason as above. It also
   doesn't feel good to attach finalizers to every pod.
3. Configure envoy to perform retries with circuit breaking to cull stopped
   backends. This has many problems including: hiding real errors, increasing
   latency when backends come back up, etc.

After discussion, we feel that these alternatives constitute a workaround for
invalid configuration (Routes referring to destinations that do not exist).

Proposal 2 is also a type of workaround for invalid configuration, but we feel
it's more acceptable because there is still an App that exists in some form that
can be referred to by the Route.

## Decision

TBD

## Consequences

TBD
