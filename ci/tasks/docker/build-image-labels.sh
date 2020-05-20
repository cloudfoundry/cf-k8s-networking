#!/bin/bash

set -euo pipefail

deplab \
  --image-tar initial-image/image \
  --git cf-k8s-networking \
  --metadata-file metadata.json

echo -n '{"io.pivotal.metadata": ' > labels/labels.json
cat metadata.json | jq '. | tostring' >> labels/labels.json
echo '}' >> labels/labels.json
