# 14. Ingress Gateway as a daemon set instead of a deployment

Date: 2020-06-26

## Status

Accepted

## Context

By default the Istio Ingress Gateway is deployed as a Kubernetes Deployment.
Along with a Kubernetes Load Balancer Service. This is fine for clusters that
support Load Balancer Services. For clusters that do not, it takes more effort
to configure the Istio Ingress Gateway in a way that is accessible from outside
the cluster while also using the well-known http/https ports (80/443).

## Decision

CF-K8s-Networking changes the Istio Ingress Gateway to be deployed as a Daemon
Set to make it easier for users that can't use a Kubernetes Load Balancer
Service on their clusters to try cf-for-k8s. By deploying a Daemon Set we can
bind port 80 and 443 on each Node to the Istio Ingress Gateway directly. This
allows a user to send traffic to each node on port 80 and 443 without
needing a Kubernetes Service.

## Consequences

- Easier for an operator to get started with cf-for-k8s.
- Less control over the number of Istio Ingress Gateways for the cluster. There
  is performance concerns with having a large number of gateways on a large
  cluster.
