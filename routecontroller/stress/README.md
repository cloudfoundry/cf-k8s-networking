# Route Controller Stress Tests

## Proposed Metrics

- Create 1000 routes, then deploy RouteController and measure how long it takes
  for all the VirtualServices and Services to be created?
- Once there are 1000 routes (and their corresponding VirtualServices and
  Services), if we add 100 more routes quickly, how long does it take to create
  the VirtualService and Service?
- Given I have 1100 routes, how long does it take to modify 100 of routes as
  individual updates?
- Given I have 1100 routes, how long does it take to modify 1000 in bulk?
- Given I have 1100 routes, how long does it take to remove 100 routes ?
- Given I have 1000 routes, how long does it take to remove all the routes?

For each of these 6 metrics, we measure the time that each change takes to fully
propagate by observing modifications to VirtualServices and Services, then the
avg propagation of 95% of the new routes is under a baseline. If we fall outside
of some tolerance range of the baseline, the test fails.

## Architectural Details
- runs on KIND
- only deploy the Route CRD and the VirtualService CRD from Istio
- No Istio
- No deployment of cf-for-k8s
- No AIs
- deploys RouteController as part of each batch of tests

## Run the tests
```
cd cf-k8s-networking/routecontroller/stress
ginkgo .
```
