#!/bin/bash

set -euo pipefail

if [[ -d input-image-tar ]]; then
  deplab --image-tar input-image-tar/image.tar \
    --git repository \
    --output-tar output-image/image.tar
elif [[ -d input-image-name ]]; then
  deplab --image "$(cat input-image-name/name.txt)" \
    --git repository \
    --output-tar output-image/image.tar
else
      echo "When using this task, you must specify EITHER input-image-tar OR input-image-name"
fi

