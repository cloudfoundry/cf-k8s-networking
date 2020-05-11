# 11. Routecontroller Route Deletion Finalizer

Date: 2020-05-11

## Status

Accepted

## Context
Without a finalizer, we could rely on cascading deletion of child objects when Routes are deleted and `RequeueAfter` to rebuild the virtual services and services  from its routes every `ResyncInterval` seconds. `ResyncInterval` is currently set to 30 seconds.

However, this doesn’t meet our SLO to handle changes within 10 seconds. Cascading deletes don’t work within 10 seconds, so meeting the SLO would require `ResyncInterval` to be less than 10 seconds, which seems unreasonable. 

Cascading deletion alone also doesn’t handle the case of a virtual service being owned by >1 Route. This means a cascading delete cannot update the virtual service’s contents to not include the paths related to that deleted route. So we would need to rely on `RequeueAfter` for those updates, which would be slow.


## Decision

In order to handle all of the cases: deleting services, deleting virtual services owned by only that route and updating virtual services owned by many routes, we rely on a finalizer, so we can have a “fast path” to all of these cases. 

Finalizers do a “soft delete” to keep the route in the K8s API while handling deletion/updates to the route’s child objects.  

Using finalizers as a “fast path” and `RequeueAfter` as a “sync” to handle any failures can help us meet our SLO of having 95% of route changes being reflected within 10 seconds.

## Consequences

* We can meet our SLO of having 95% of route changes being reflected within 10 seconds when routes are deleted
