#!/bin/bash

mkdir -p /tmp/good-acceptance/

gsutil cp gs://cf-k8s-networking/good-acceptance/cf-install-values.yml /tmp/good-acceptance/cf-values.yml

echo "You can now find the values for good-acceptance in /tmp/good-acceptance/cf-values.yml"
