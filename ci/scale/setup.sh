#!/bin/bash

echo "This script will create 1,000 apps with 1,000 routes pointing to them. It is recommended you have at least 100 nodes."
read -p "Are you sure you want to continue? [yn] " -n 1 -r
echo    # (optional) move to a new line
if [[ ! $REPLY =~ ^[Yy]$ ]]
then
    exit 1
fi

cf update-quota default -r 3000 -m 3T

for n in {0..999}
do
  echo "bin-$n"
  cf push bin-$n -o cfrouting/httpbin8080
done
