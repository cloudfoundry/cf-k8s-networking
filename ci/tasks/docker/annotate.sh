#!/bin/bash

set -euo pipefail

deplab --image-tar input-image/image.tar \
  --git repository \
  --output-tar output-image/image.tar

