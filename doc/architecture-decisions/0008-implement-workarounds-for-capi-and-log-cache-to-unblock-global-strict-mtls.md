# 8. Implement Workarounds for CAPI and Log-Cache to Unblock Global STRICT mTLS

Date: 2020-02-24

## Status

Accepted

## Context

We need to turn on STRICT mTLS for all components on the mesh. However, some
components are currently incompatible with this mode.

CAPI is incompatible because it uses an init container to run migrations. This
init container comes up before the sidecar, so it is unable to establish an mTLS
connection with the capi database. This causes the init container to fail and
prevents capi from coming up. See [this
issue](https://github.com/cloudfoundry/capi-k8s-release/issues/12) in capi.

Log-cache is incompatible because it is configured to establish its own tls
connection, which is incompatible with the mTLS the sidecars are attempting to
establish.

## Decision

We have provided configuration workarounds in the form of Policies, that were
placed in the cf-for-k8s repo to be owned by the respective teams that manage
the troublesome components.

[Pull Request](https://github.com/cloudfoundry/cf-for-k8s/pull/35)


## Consequences

These components will accept plain text communication. We don't consider this to
be a significant issue because both already implement encryption in some form on
their own. It would be best in the long run if they could stop being exceptions
though.

The log-cache and capi teams now have to care about istio configuration, and
will eventually need to make changes to their components to eliminate these
workarounds. However, our work is no longer blocked on their changes, so we
consider this an absolute win.

This is the way. 🗞🙃
