## Redeploying the Acceptance Environment

0. Let the deploy finish
1. In this directory, run `./fetch-acceptance-values.sh`
1. Check out the commit SHA of cf-for-k8s-master in your local cf-for-k8s:
   - Click `get: cf-for-k8s-master` and copy the value for `commit`
   - `cd ~/workspace/cf-for-k8s/`
   - `git checkout <sha>`
1. Make any changes desired to `/tmp/good-acceptance/cf-values.yml`
1. Run `./create-and-deploy.sh good-acceptance`

You're done!
