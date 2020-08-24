#!/bin/bash

set -eux

echo "Retagging istio images from distroless to cf:tiny"

istio_tag=$(cat ./istio-release/tag)

docker pull proxy
deplab proxy
tag new images

