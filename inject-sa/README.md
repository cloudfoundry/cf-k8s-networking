# Route Integrity Implementation

Refer to [proposal
doc](https://docs.google.com/document/d/182wjnQRyjIc7oBffvaeP37JlivAXMmZQT3kTK5bZWYE/edit#)
for more information.

## Implementation

It's a simple mutating admission controller which watches for pods in
`cf-workloads` namespace and modifies the `serviceAccountName` field and create
the service account if it doesn't exist.

## Links

* [A Guide to Kubernetes Admission
  Controllers](https://kubernetes.io/blog/2019/03/21/a-guide-to-kubernetes-admission-controllers/)
* [kubernetes/client-go simple
    example](https://github.com/kubernetes/client-go/tree/master/examples/in-cluster-client-configuration)
* [Route Integrity
  proposal](https://docs.google.com/document/d/182wjnQRyjIc7oBffvaeP37JlivAXMmZQT3kTK5bZWYE/edit#)
