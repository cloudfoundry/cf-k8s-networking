# 13. Stopped App Routing

Date: 2020-05-29
Updated: 2020-06-02

## Status

Proposal

## Context

In CF-for-VMs, when a route is mapped to multiple apps, some started and some
stopped, gorouter only directs requests to the apps that are started. In the
case of crashed apps the gorouter also only directs requests to the apps that
are started and healthy. If **ALL** destinations are stopped or crashed then a
503 response will occur.

In CF-for-k8s _currently_, we balance traffic equally between the started and
stopped apps. This results in 503 errors for the user, and is not feature
parity. Intermittent 503 responses similarly occur for crashed apps if there are
multiple apps mapped to a single route. However, in the case of multiple
instances of a single app, routing is handled correctly when some instances
crash and some are healthy i.e it will not 503. This is because multiple apps
mapped to a single route result in multiple destinations in a Virtual Service,
but a single app with multiple instances results in a single destination and can
take advantage of a Kubernetes Service watching the Readiness Probe of a pod.

This is because the components that create the Istio configuration do not
currently know about the status of apps. They assume all destinations on a Route
object are up and able to receive traffic. More concretely the Services that are
created use the app guid and process type as selectors. When an application is
stopped the StatefulSet and Pods backing that application are deleted. This
means the Selector can't select any pods and results in a 503.

This ADR is specifically focused on handling routing to stopped applications.

See this issue: https://github.com/cloudfoundry/cf-for-k8s/issues/158
Additional discussion here: https://github.com/cloudfoundry/cf-k8s-networking/pull/46

## Proposal

We have two proposals:

1. Cloud Controller only adds destinations to Routes that are running. This does
   go against our ideal state of the Route CR being our high-level description
   of desired intent and not just an intermediate representation of the current
   state from the Cloud Controller. However, we do see this as a step that could
   later evolve into something like proposal two, where the Route CR can return
   to being a high-level description of desired intent.

   In addition, this proposal does mean that a route needs to be updated
   whenever process state changes for an app. This does increase the surface
   area of API calls that need to be considered to allow them to update the
   Route CR when process state changes e.g `cf start`, `cf stop`, `cf push`, `cf
   restart` etc. This might call for revisiting the decision to defer the
   CCDB->k8s Route bulk sync work because the Route CR will be updated more
   often and there is an increased chance of consistency issues.
   * _Note:_ Routes with _no_ destinations are explicitly supported already.

2. Cloud Controller writes an App CR (this is a new CRD) that contains the
   application's status, so that the routecontroller can use that to decide how
   to configure Istio. This retains the idealized version of the Route CR, a
   high-level description of desired intent.

Our current recommendation is to first implement proposal one to hit our GA
targets as we see this as the more straightforward one to implement. Afterwards,
we will transition to implementing proposal two so the Route CR can return to
being a description of desired intent and not an intermediate representation of
the current state.

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
4. Use [Kubernetes Readiness
   Probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-readiness-probes)
   on the StatefulSet/Pods. This doesn't work because when an appliction is
   stopped the StatefulSet and Pods are deleted so there is no longer anywhere
   to add a readiness probe.

After discussion, we feel that these alternatives constitute a workaround for
invalid configuration (Routes referring to destinations that do not exist).

Proposal 2 is also a type of workaround for invalid configuration, but we feel
it's more acceptable because there is still an App that exists in some form that
can be referred to by the Route.

## Decision

TBD

## Consequences

TBD
