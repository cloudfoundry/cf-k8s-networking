## Simple CP Latency Test

This folder contains scripts for a simple control plane latency test.

To run it:

1. Create a huge cluster by running `./hack/cf4k8s/create-huge.sh`. This will
   create a gke cluster of the appropriate size. (100 n-standard-8 nodes with ip
   aliasing and network policy enabled).
2. Run `setup.sh` in this directory to push 1,000 apps.
3. Run `./cfscale.sh 100` to run a 100 datapoint test. It'll output all the
   latencies it sees in ascending order, so you can see the 95th percentile by
   looking at the 5th one from the bottom :)

Keep in mind this is a very rudimentary test, for serious scale tests you should
use significantly more than 100 datapoints.
