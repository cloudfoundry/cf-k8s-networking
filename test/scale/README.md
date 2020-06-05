## Control Plane Latency Test

### Prerequisites
#### Cluster Requirements

These scale tests are meant to be run against a large kubernetes cluster. We run
these tests in CI on GKE clusters with 100 `n1-standard-8` nodes with [ip
aliasing](https://cloud.google.com/sdk/gcloud/reference/beta/container/clusters/create#--enable-ip-alias)
and [network policy
enabled](https://cloud.google.com/sdk/gcloud/reference/beta/container/clusters/create#--enable-network-policy).

#### Environment Setup

After deploying an appropriate size GKE cluster, CI will deploy cf-for-k8s and
push 2000 app instances (1000 apps, 2 instances per app) with 1000 routes (1 route
per app).

### Tests
#### Steady State Test

The steady state test runs once the environment has been set up with 2000 app
instances and 1000 routes. This test is "steady state" because it keeps the
number of routes constant at 1000, by deleting one route every time it maps a
new route.  We chose to use a steady state test because we want to keep the
number of routes constant to measure the control plane latency under the desired
load.

For each route the test maps, the test measures the latency from when the route
is mapped until when that route is reachable. The test asserts that this latency
is under 10 seconds for 95% of the `map-route` requests.

_Note_: This test currently does not pass using the cfroutesync component of
cf-k8s-networking.

### CI
Currently we run scale tests as defined in the [scaling
pipeline](../scaling.yml).
