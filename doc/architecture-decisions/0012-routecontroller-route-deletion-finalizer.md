# 12. Routecontroller Route Deletion Finalizer

Date: 2020-05-11

## Status

Accepted

## Context
A [finalizer](https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#finalizers) allows you to write asynchronous pre-delete logic, such as deleting associated resources. Here's an [example](https://book.kubebuilder.io/reference/using-finalizers.html) of finalizers can be used with Kubebuilder.

Without a finalizer, we could rely on cascading deletion of child objects when Routes are deleted and `RequeueAfter` to rebuild the virtual services and services  from its routes every `ResyncInterval` seconds. `ResyncInterval` is currently set to 30 seconds.

However, this doesn’t meet our SLO to handle changes within 10 seconds. [Cascading deletes don’t work within 10 seconds](https://github.com/kubernetes/kubernetes/blob/af67408c172630d59996207a2f3587ea88c96572/test/integration/garbagecollector/garbage_collector_test.go#L385-L392), so meeting the SLO would require `ResyncInterval` to be less than 10 seconds, which seems unreasonable.

Cascading deletion alone also doesn’t handle the case of a virtual service being owned by >1 Route. This means a cascading delete cannot update the virtual service’s contents to not include the paths related to that deleted route. So we would need to rely on `RequeueAfter` on a different route with the same FQDN for those updates, which would be slow, and a strange behavior to support.


## Decision

In order to handle all of the cases: deleting services, deleting virtual services owned by only that route and updating virtual services owned by many routes, we rely on a finalizer, so we can have a “fast path” to all of these cases. 

Finalizers do a “soft delete” to keep the route in the K8s API while handling deletion/updates to the route’s child objects.  

Using finalizers allows us to implement all of the cases in our route deletion logic. This helps us meet our SLO of having 95% of route changes being reflected within 10 seconds.`RequeueAfter` serves as a “sync” to handle disaster recovery scenarios when unexpected operations outside of normal controller reconciliation happen (i.e child resources are deleted in etcd).

## Consequences

* We can meet our SLO of having 95% of route changes being reflected within 10 seconds when routes are deleted
* Routes will fail to be deleted if there is no routecontroller available to resolve the finalizer
  * As a result, cf-for-k8s will now delete the workload namespaces first when deleting a CF deployment
