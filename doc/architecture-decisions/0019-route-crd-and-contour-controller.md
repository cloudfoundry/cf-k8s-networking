# 19. Allow Alternative Ingress Solution Provider in RouteController

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

#### Why not make a separate controller altogether?

We are opting to extend routecontroller instead of make a separate one because
we believe it to be simpler. Best practice is to have one controller reconciling
objects of one type. Because routecontroller only watches for route CRs, it
doesn't break that best practice. Whether or Virtual Services or HTTPProxies are
created as a result does not matter.

It seems overbearing to maintain a separate controller and all the boilerplate
around it when all we really need to a separate resource builder.

We plan to make routecontroller only create one type of resource or another,
never both. This will prevent the confusing situation of istio resources
existing in the cluster when contour is the selected ingress solution provider,
or vice versa.

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

