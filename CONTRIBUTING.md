# Contributing to CF-K8s-Networking

The Cloud Foundry team uses GitHub and accepts code contributions via [pull
requests](https://help.github.com/articles/about-pull-requests/).

## Prerequisites

Before working on a PR to the CF-K8s-Networking code base, please:

  - reach out to us first via a [GitHub issue](https://github.com/cloudfoundry/cf-k8s-networking/issues),

You can always chat with us on our [Slack #networking channel](https://cloudfoundry.slack.com/app_redirect?channel=CFX13JK7B) ([request an invite](http://slack.cloudfoundry.org/)),

After reaching out to the App Connectivity team and the conclusion is to make a PR, please follow these steps:

1. Ensure that you have either:
   * completed our Contributor License Agreement (CLA) for individuals (if you
     haven't done this already the PR will prompt you to)
   * or, are a [public member](https://help.github.com/articles/publicizing-or-hiding-organization-membership/) of an organization
   that has signed the corporate CLA.
1. Fork the project repository.
1. Create a feature branch (e.g. `git checkout -b good_network`) and make changes on this branch
   * Tests are required for any changes.
1. Push to your fork (e.g. `git push origin good_network`) and [submit a pull request](https://help.github.com/articles/creating-a-pull-request)

Note: All contributions must be sent using GitHub Pull Requests.
We prefer a small, focused pull request with a clear message
that conveys the intent of your change.

## Local development for cf-k8s-networking

### Development dependencies

#### Golang
We currently build with Golang `1.13.x` and use `go mod` for dependencies.

* [`go`](https://golang.org/)
* [`ginkgo`](https://github.com/onsi/ginkgo)

#### k14s Tools
Most of our templating and deploy scripts rely on the [k14s Kubernetes Tools](https://k14s.io/). We recommend installing the latest versions.

* `ytt`
* `kapp`
* `vendir`

#### Integration Test Dependencies
Our integration tests spin up temporary
[Kind](https://kubernetes.io/docs/setup/learning-environment/kind/) clusters
using Docker.

* [`docker`](https://docs.docker.com/get-docker/)
* [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
* [`kind`](https://kind.sigs.k8s.io/docs/user/quick-start/)


### Running Tests
When contributing to this project you should, at a minimum, be running the
`routecontroller` unit and integration tests. We also recommend running the Networking Acceptance Tests.

#### RouteController Unit and Integration Tests
1. `cd cf-k8s-networking/routecontroller`
1. `./scripts/test`
1. `./scripts/integration`

#### Networking Acceptance Tests
To run the acceptance tests, you must have a Kubernetes cluster provisioned.
Follow the steps in [./test/acceptance/](./test/acceptance/README.md) to run
these tests.

### Other Tests
We have a few additional test suites that run in our CI.
* [Scaling Tests in
  CI](https://networking.ci.cf-app.com/teams/cf-k8s/pipelines/scaling).
* [Uptime Tests in
  CI](https://networking.ci.cf-app.com/teams/cf-k8s/pipelines/cf-k8s-upgrade)

#### CF-K8s-Networking Upgradeability Uptime Tests
1. For more information, check this [README](test/uptime/README.md)

#### CF-K8s-Networking Scaling Tests
1. For more information, check this [README](test/scale/README.md)

### Deploying your changes
CF-K8s-Networking is a set of components meant to be integrated into a
[cf-for-k8s](https://github.com/cloudfoundry/cf-for-k8s) deployment.

To deploy your local changes to `cf-k8s-networking` with `cf-for-k8s`, you can
follow these steps:

1. Run `./scripts/vendir-sync-local` which will run `vendir sync` in
   `cf-for-k8s` with override to use the local cf-k8s-networking config.
1. Follow docs to install
   [`cf-for-k8s`](https://github.com/cloudfoundry/cf-for-k8s/blob/master/docs/deploy.md)
