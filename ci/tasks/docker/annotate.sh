#!/bin/bash

set -euo pipefail

deplab --image-tar input-image/image.tar \
  --git respository \
  --output-tar output-image/image.tar

