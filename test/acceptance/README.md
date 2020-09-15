# Networking Acceptance Tests

## Requirements

To run tests you need to have the following installed:

* [kapp](https://k14s.io/)

  ```bash
  $ wget -O- https://k14s.io/install.sh | bash
  ```

* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

Other requirements:

* You should have `kubectl` config with the access to system and workload namespaces to GET/POST/PUT/PATCH pods, services, service accounts and execute commands in pods.

* JSON-formatted configuration file with the following required fields:

  ```
  {
   "kubeconfig_path": "/Users/user/.kube/config",
   "api": "api.example.com",
   "admin_user": "admin",
   "admin_password": "PASSWORD",
   "apps_domain": "apps.example.com"
  }
  ```

* `diego_docker` feature flag enabled in your CF deployment:

  ```bash
  cf enable-feature-flag diego_docker
  ```

## Run

```bash
# make sure you targeted your cluster before executing this
cd test/acceptance
./bin/test_local <path to config.json> [path to kube config]
```


## Configuration

As was mentioned [configuration file](cfg/cfg.go) is a subset of [CATS config file](https://github.com/cloudfoundry/cf-acceptance-tests#test-configuration) with some additions.

There are few environment variables which can be used to control tests setup:

* `CONFIG_KEEP_CLUSTER=1` to not destroy deployed pods and services after tests, helpful for debugging in CI
* `CONFIG_KEEP_CF=1` to not revert changes in CF after tests, helpful for debugging in CI
