## Understanding Access Logs

### Duration fields

The access log contains the following fields for duration:

- `upstream_service_time`: time in milliseconds spent by the workload processing the request. 
- `duration`: total duration in milliseconds of the request from the start time to the last byte out.
- `response_duration`: total duration in milliseconds of the request from the start time to the first byte read from the workload.
- `response_tx_duration`: total duration in milliseconds of the request from the first byte read from the workload to the last byte sent downstream.

If you want to determine workload time compared to network and router latency you can compare `upstream_service_time` which 
is time spent in the workload with `response_duration - upstream_service_time` which will contain router and network latency.
Latencies above 100ms can indicate problems with the network. An alert value on this metric should be tuned to the specifics of the deployment and its underlying network considerations; a suggested starting point is 100ms.

![](assets/duration-flamegraph.jpg)
