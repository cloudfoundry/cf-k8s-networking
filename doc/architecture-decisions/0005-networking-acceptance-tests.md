# 5. Networking Acceptance Tests

Date: 2020-02-04
Updated: 2020-06-05

## Status

Accepted

## Context

We wrote a set of [networking-acceptance-tests](../../test/acceptance) that require a CF for Kubernetes
environment set up correctly in order to test networking behavior in an integrated environment.

## Decision

These tests are different from [integration tests](../../routecontroller/integration) since they require an integrated environment setup.
These tests are different from CATs in that they test specialized networking setup in CF for Kubernetes.

We've decided to keep these acceptance tests in this repository because it is simple
and they are run in CI and rely on this
[script](../../ci/tasks/tests/run-networking-acceptance-gke.sh).

Tests should be included in [networking-acceptance-tests](../../test/acceptance) if they require a CF for
Kubernetes environment and test the setup of networking.

## Addendum
2020-06-19: Updated the "integration tests" link to point to the
`routecontroller` directory and updated the "script" link as per [ADR
010](./0010-route-crd-and-kubebuilder-instead-of-metacontroller.md)
