# 11. Use KIND Clusters for Routecontroller Integration

Date: 2020-05-05

## Status

Accepted

## Context

We are working to get up and running quickly with our new routecontroller
refactor. Much of our work using kubebuilder is informed by what the
ingress-router team learned during their work with it.

The cfroutesync integration tests used GKE and were very fast, however they were
unwieldy and difficult to reason about, as they'd been written to test a very
specific set of circumstances. Since we were not happy with the way those
integration tests worked, we had an opportunity to rethink our tests, and there
was a model available from ingress-router, KIND seemed like the best option.

## Decision

We decided to use KIND as it is a full and lightweight Kubernetes environment
that creates clusters within docker containers.

## Consequences

* Our tests run a bit slower because we're creating a KIND cluster before each
* Our tests cannot pollute each other because a new cluster is used each time
* Test setup is far less complicated, no more provisioning a GKE cluster before
  you can run them
* Tests are easier to reason about and write because there are no surprise
  resources on it like with the cfroutesync tests
* You can run `ginkgo .` and run all the tests, locally, without any setup, in
  under 10 minutes
