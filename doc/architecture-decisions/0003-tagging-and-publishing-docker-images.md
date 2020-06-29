# 3. Tagging and Publishing Dev/CI Images

Date: 2019-10-31

## Status

Accepted

## Context
(🎃 Happy Halloween 👻)

We need to have a way of creating and deploying container images of our software for development and CI use.

We didn't want to break the `latest` image tag every time we did local development and we need to have CI deploy a consistent image between runs.

## Decision

### Local development
Each development environment pipeline deploys off a dedicated docker tag.  e.g. `eirini-dev-1` environment deploys the
docker image tagged `eirini-dev-1` so when developing locally (i.e. without pushing to Git)
we can tag and push images with that dedicated tag and redeploy easily.

Example Workflow:
```bash
environment_name=eirini-dev-1
docker tag $img gcr.io/cf-routing/cf-k8s-networking/cfroutesync:$environment_name
docker push gcr.io/cf-routing/cf-k8s-networking/cfroutesync:$environment_name
```

### Branch development
A github action will trigger on pushes to all branches and publish a Docker image tagged with the git SHA and the branch name.

#### Develop branch
When we push to the `develop` branch we will tag the image with the git SHA, branch name, and `latest`.

## Consequences

We will likely be producing a significant amount of images with this workflow (one for each push) so eventually we will need to figure out a way of pruning old ones.
For now though our images are pretty small so we feel we can defer this work.

## Addendum
2020-06-26: Replaced "master" branch with "develop" branch as described in this
[ADR](./0013-rename-master-branch.md).
