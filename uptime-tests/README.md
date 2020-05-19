# Uptime Test

## Global Configuration

UPGRADE_DISCOVERY_TIMEOUT is the amount of time given to discover an upgrade is
happening using `kapp app-change ls -a cf`.

CF_APP_DOMAIN is the app domain. This is currently used to map new routes to
test control plane uptime.

## Data Plane Service Level Measurements

Our SLO is defined as:

X% of GET Requests to a Route succeed in less than Y milliseconds.

The SLI used to measure our SLO is request latency.

Request Latency is measured by the following:

Given an app 'A' deployed on the platform with route 'r', the SLI times how long it
takes to make an HTTP Get Request to 'r' for 'A' and receive a response.

### Configuration

X: DATA_PLANE_SLO_PERCENTAGE
Y: DATA_PLANE_SLO_MAX_REQUEST_LATENCY
r: DATA_PLANE_SLI_APP_ROUTE_URL

## Control Plane Service Level Measurements

Our SLO is defined as:

X% of routes that get mapped become available in less than Y seconds.

Available is defined as Z% of GET Requests to a newly mapped Route succeed in
less than W milliseconds after Y seconds and up to U seconds.

The SLI used to measure our SLO is route propagation latency.

Route Propagation latency is measured by the following:

Given an app 'A' deployed on the platform, the SLI times how long it
takes to make a route available. This is done as follows:

1. Over the course of the upgrade, map a route at an interval.
2. After Y seconds, make GET requests to the route for U seconds and record request latency,
   and if it was a successful response (e.g 200 status code).
3. For those requests, the percentage of successful requests whose
   request latency is under W should be greater than Z.

### Configuration

X: CONTROL_PLANE_SLO_PERCENTAGE
Y: CONTROL_PLANE_SLO_MAX_ROUTE_PROPAGATION_TIME
Z: CONTROL_PLANE_SLO_DATA_PLANE_AVAILABILITY_PERCENTAGE
W: CONTROL_PLANE_SLO_DATA_PLANE_MAX_REQUEST_LATENCY
U: CONTROL_PLANE_SLO_SAMPLE_CAPTURE_TIME

CONTROL_PLANE_APP_NAME is the name of the app to map routes to.
