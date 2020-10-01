# 15. App Access Logs from IngressGateway

Date: 2020-07-06

## Status

Accepted

## Context

`cf logs` for an app shows logs both emitted by that app, and access logs from
the Gorouter. The access logs from Gorouter look like this in the logstream:

```
2020-06-25T23:42:19.00+0000 [<source_type>/<instance_id>] OUT <log>
```

In cf-for-k8s, an app developer should similarly be able to see access logs as
requests for the app travel through the routing tier. One discussion we had was
whether these access logs should come from the ingressgateway envoy and/or from the
app sidecar envoys.


One important piece of functionality in cf-for-BOSH is that when an app process has been killed, healthcheck requests still
show up in the access log stream. This is because those requests still make it
to the Gorouter, even though they do not make it to the app itself.

This is an example of an access log of a healthcheck request to a killed app.
The 503 is being returned directly from the Gorouter.
```
2020-07-06T10:45:55.83-0700 [RTR/0] OUT dora.maximumpurple.cf-app.com - [2020-07-06T17:45:55.828757970Z] "GET /health HTTP/1.1" 503 0 24 "-" "curl/7.54.0" "35.191.2.88:63168" "10.0.1.11:61002" x_forwarded_for:"76.126.189.35, 34.102.206.8, 35.191.2.88" x_forwarded_proto:"http" vcap_request_id:"0cd79f32-3cde-4eea-5853-9a2ca401be40" response_time:0.004478 gorouter_time:0.000433 app_id:"1e196708-3b2d-4edc-b5b8-bf6b1119d802" app_index:"0" x_b3_traceid:"7f470cc2fcf44cc6" x_b3_spanid:"7f470cc2fcf44cc6" x_b3_parentspanid:"-" b3:"7f470cc2fcf44cc6-7f470cc2fcf44cc6"
```

We've done some previous work exploring access logs on the Istio Envoy
IngressGateway (see related stories section below) and have [documented some of
the fields
here](https://github.com/cloudfoundry/cf-k8s-networking/blob/37dabf7907ffa7b284980cfcb6813ebcd449736c/doc/access-logs.md).

## Decision

We decided to have the access logs come from the ingressgateway to begin with,
as we think those provide the most valuable information.

Imagine a scenario where an app has crashed and the Pod is being rescheduled.
The Envoy on the ingressgateway will still log this failed request. The sidecar,
on the other hand, would be unreachable so it would not be able to log anything.
Having the failed request in the access logs in this scenario could be valuable
information for a developer attempting to debug their app with `cf logs`.

We also decided that the access log format would be JSON with the [following
fields](https://docs.google.com/spreadsheets/d/1CuvoUEkiizVKvSZ2IaLya40sgMbm5at78CqxB8uUe80/edit#gid=0)

The work to enable this was completed in [#173568724](https://www.pivotaltracker.com/story/show/173568724).

## Consequences

- In order to enable this, we added fluent-bit sidecars to our ingressgateways.
  Information on why we decided to add our own fluent-bit images can be found in
  this [draft PR](https://github.com/cloudfoundry/cf-k8s-networking/pull/57).
  The final iteration of this was merged in from [this
  PR](https://github.com/cloudfoundry/cf-k8s-networking/pull/63)
- Will need to do [some extra work](https://www.pivotaltracker.com/story/show/172732552)
to get logs from the ingressgateway pods into the log stream corresponding to the destination app.
See https://github.com/cloudfoundry/cf-k8s-logging/tree/main/examples/forwarder-fluent-bit for more information.
- It is unclear how difficult it would be to add custom formatting to sidecar Envoy logs,
we [know how to do it for the ingressgateway logs](https://www.pivotaltracker.com/story/show/169739120)
- We may need to revisit the sidecar logs later if we want access logs for container-to-container (c2c)
networking (this doesn't exist for c2c in CF for BOSH today)

---

## Related Stories
For additional context, here are some stories our team has worked on in the
past:

- [Emit JSON ingress gateway access logs](https://www.pivotaltracker.com/story/show/169739120)
- [Adding fields into access logs for gorouter parity](https://www.pivotaltracker.com/story/show/169737156)
- [Explore emitting ingress gateway access logs with Fluentd](https://www.pivotaltracker.com/story/show/170119094)
