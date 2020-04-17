# 9. Kubebuilder Controllers: Use the Controller Runtime Dynamic Client over Generated Clients

Date: 2020-04-17

## Status

Accepted

## Context

### Kubebuilder uses dynamic controller-runtime clients by default
Kubebuilder uses the
[controller-runtime](https://github.com/kubernetes-sigs/controller-runtime)
library. Controller Runtime has a dynamic
[client](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client) that
is used by Kubebuilder controllers by default, rather than a generated
[client](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/generating-clientset.md).
When you `kubebuilder create api`, it creates the api types and in order to
interact with these api types the controller is supplied a controller-runtime
client in it's controller scaffolding so you can CRUD that api type in the
Kubernetes API.

### Problems with generated clients

Using third party generated clients can also be problamatic because of the
transitive dependency on the Kubernetes
[client-go](https://github.com/kubernetes/client-go) library and our own
depdendency on client-go. When our controllers want to use a newer version of
the client-go library, this can cause problems for our third party generated
clients because they will use a different version of client-go. This doesn't
cause problems if client-go keeps the same interface, but we have seen newer
versions of client-go break its public interface causing compilation issues.

If the third party libraries ensured they updated their libraries to use the
same version of client-go in a timely manner, this could be less of a
problem. However, this puts a dependency on these third party libraries to
keep their client-go libraries up-to-date.

## Decision

We will only use the controller-runtime client to interact with Kubernetes
API objects instead of generated clients. This limits our dependency on third
party libraries that can cause conflicts with the client-go library.

## Consequences

- To interact with Istio objects we won't use the istio/client-go library and
instead use the controller-runtime client with the istio/api library
directly. This does require us to wrap the istio/api objects in our own
Kubernetes specific API structs.
- Updating our version of client-go won't require us to bump a plethora of
third party libraries that also use client-go.