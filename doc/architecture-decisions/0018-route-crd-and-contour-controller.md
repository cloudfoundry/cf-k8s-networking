# 18. Allow Alternative Ingress Solution Provider in RouteController

Date: 2020-10-26

## Status

Proposal

## Context
In our efforts to allow more optionality in ingress solutions for CF-for-K8s, we
want to allow the use of Contour as a potential alternative to Istio.

### Proposed Design

1. Adding a configuration option `ingress_solution_provider` to CF-for-K8s, with
   potential values `istio` or `contour`.
2. Extend routecontroller to respect the value configured by `ingress_solution_provider: contour`. It will create the appropriate resources based on the chosen provider.

### Open Questions
1. What do we do about the config? Does contour config live in cf-k8s-networking
   and eventually it moves to cf-for-k8s?
   * The config will live in cf-for-k8s.
2. What happens if an operator wants to change their ingress solution provider,
   will they have to redeploy CF-for-K8s? Is that a big deal?
   * Yes, they will.
3. Looks like Contour deploys Envoys as a daemonset. Is that gonna be a problem?
   * This is just part of the quickstart.yaml for learning contour, we can
     change it to Deployment if we need to.

## Decision
Waiting on Review

## Consequences
* RouteController will only be able to create resources for 1 type of ingress
  solution at a time. When an operator makes a decision by
  configuring `ingress_solution_provider`, only resources related to the
  specified provider will be created.
* The available networking abilities are limited to that of the selected ingress
  solution provider. Outcomes achievable only with Istio will not be available
  if Contour is selected.

#### Using Kubebuilder
* Provides Community buy-in; the `kubebuilder` framework is the encouraged way to engineer a CRD
* Provides built-in best practices for writing a controller, including: shared caching, retries, back-offs, leader election for high availability deployments, etc...

#### For Reference
The proposal and discussion for the Route CRD and design can be found [here](https://docs.google.com/document/d/1DF7eTBut1I74w_sVaQ4eeF74iQes1nG3iUv7iJ7E35U/edit?usp=sharing).

