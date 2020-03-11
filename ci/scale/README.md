## Simple CP Latency Test

This folder contains scripts for a simple control plane latency test.

To run it, create cf-for-k8s foundation that can handle 1000 httpbins
(recommended 100 standard-8 nodes). Then run `setup.sh` to add all the apps
you'll be testing again.

Then run `./cfscale.sh 100` to run a 100 datapoint test. It'll output all the
latencies it sees in ascending order, so you can see the 95th percentile by
looking at the 5th one from the bottom :)

Keep in mind this is a very rudimentary test, for serious scale tests you should
use significantly more than 100 datapoints.
