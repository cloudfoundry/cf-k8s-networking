# Uptime Test

## Global Configuration

UPGRADE_DISCOVERY_TIMEOUT is the amount of time given to discover an upgrade is
happening using `kapp app-change ls -a cf`.

## Data Plane Service Level Measurements

Our SLO is defined as:

95% (X) of GET Requests to a Route succeed in less than 100 (Y) milliseconds.

The SLI used to measure our SLO is request latency.

Request Latency is measured by the following:

Given an app 'A' deployed on the platform with route 'r', the SLI times how long it
takes to make an HTTP Get Request to 'r' for 'A' and receive a response.

### Description
CF_APP_DOMAIN is the app domain. This is currently used to map new routes to
test control plane uptime.

DATA_PLANE_APP_NAME: Name of the app.


### Configuration

X: DATA_PLANE_SLO_PERCENTAGE
Y: DATA_PLANE_SLO_MAX_REQUEST_LATENCY
r: DATA_PLANE_SLI_APP_ROUTE_URL

## Control Plane Service Level Measurements

Our SLO is defined as:

95% (X) of routes that get mapped become available in less than 15 (Y) seconds.

1. Every 5 seconds, map a route
2. Sleep for propagation time (Y)
3. Send requests to route for 30 seconds (U), record their latency and response
   code
4. If greater than 95% (Z) of those requests had a non-200 response code, or
   exceeded the latency SLO (W), consider that route to be a failure
5. If greater than 95% (X) of routes are failures, fail the test

### Description
CONTROL_PLANE_SLO_PERCENTAGE (X): Percentage of routes that we expect to get mapped
and become available in less than number of seconds defined by
`CONTROL_PLANE_SLO_MAX_ROUTE_PROPAGATION_TIME`. Defaults to 95%.

CONTROL_PLANE_SLO_MAX_ROUTE_PROPAGATION_TIME (Y): Time we wait before seeing if
a route is live, defaults to 15 seconds.

CONTROL_PLANE_SLO_DATA_PLANE_AVAILABILITY_PERCENTAGE (Z): Percentage of routes
that succeed in less than the number of seconds defined by
`CONTROL_PLANE_SLO_DATA_PLANE_AVAILABILITY_PERCENTAGE`. Defaults to 95%.

CONTROL_PLANE_SLO_DATA_PLANE_MAX_REQUEST_LATENCY (W): Max response time from
mapped route, defaults to 200ms.

CONTROL_PLANE_SLO_SAMPLE_CAPTURE_TIME (U): How often we send a request to a
route. Defaults to 10 seconds.

CONTROL_PLANE_APP_NAME: Name of the app to map routes to. Defaults to
`upgrade-control-plane-sli`.

