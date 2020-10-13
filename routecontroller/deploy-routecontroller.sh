#! /bin/bash

## Build and push a routecontroller image with latest changes to public gcr
docker build . -t gcr.io/cf-routing/scratch/routecontroller:latest
docker push gcr.io/cf-routing/scratch/routecontroller:latest

## Render config to deploy routecontroller standalone
ytt -f ../config/routecontroller -f ../config/values.yaml -v systemNamespace=default > routecontroller.yaml

## Apply the route custom resource and routcontroller config
kubectl apply -f "config/crd/bases/networking.cloudfoundry.org_routes.yaml"
kubectl apply -f "routecontroller.yaml"

