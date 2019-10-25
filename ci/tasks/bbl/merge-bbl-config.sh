#!/bin/bash

set -e -x -u -o pipefail

cp -R source1/${SOURCE1_DIR:-""}/* merged-bbl-config/

if [ -d source2 ]; then
  cp -R source2/${SOURCE2_DIR:-""}/* merged-bbl-config/
fi

if [ -d source3 ]; then
  cp -R source3/${SOURCE3_DIR:-""}/* merged-bbl-config/
fi

if [ -d source4 ]; then
  cp -R source4/${SOURCE4_DIR:-""}/* merged-bbl-config/
fi

if [ -d source5 ]; then
  cp -R source5/${SOURCE5_DIR:-""}/* merged-bbl-config/
fi

ls -R merged-bbl-config
