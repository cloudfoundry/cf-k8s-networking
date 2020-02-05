# 5. Networking Acceptance Tests

Date: 2020-02-04

## Status

Proposal

## Context

We wrote a set of [networking-acceptance-tests](../../networking-acceptance-tests) that require a CF for Kubernetes 
environment set up correctly in order to test networking behavior in an integrated environment. 

## Decision

These tests are different from [integration tests](../../cfroutesync/integration) since they require an integrated environment setup.
These tests are different from CATs in that they test specialized networking setup in CF for Kubernetes.

We've decided to keep these acceptance tests in this repository because it is simple
and they are run in CI and rely on this [utils script](../../ci/tasks/k8s/utils.sh).

Tests should be included in [networking-acceptance-tests](../../networking-acceptance-tests) if they require a CF for 
Kubernetes environment and test the setup of networking.

