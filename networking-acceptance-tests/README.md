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

* JSON configuration file of the following format (subset of [CATS config file](https://github.com/cloudfoundry/cf-acceptance-tests#test-configuration)):

  ```json
  {
    "api": "URL for CF API",
    "admin_user": "CF admin user username",
    "admin_password": "CF admin user password"
  }
  ```
  
* `diego_docker` feature flag enabled in your CF deployment:

  ```bash
  cf enable-feature-flag diego_docker
  ```

## Run

### Without bbl-state

```bash
# make sure you targeted your cluster before executing this
cd networking-acceptance-tests
./bin/test_local <path to config.json> [path to kube config]
```

## With bbl-state

```bash
cd networking-acceptance-tests
./bin/test_with_bbl_state <path to config.json> <path to bbl state> [path to kube config] 
```
