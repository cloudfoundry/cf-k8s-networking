#!/usr/bin/env bash

pushd k8s-deploy-image > /dev/null
    digest="$(cat digest)"
popd

pushd cf-k8s-networking
    sed -i "s/cf-k8s-networking\/cfroutesync:.*/cf-k8s-networking\/cfroutesync:$digest/" config/cfroutesync/values.yaml

    git config user.name "${GIT_COMMIT_USERNAME}"
    git config user.email "${GIT_COMMIT_EMAIL}"

    if [[ -n $(git status --porcelain) ]]; then
        echo "changes detected, will commit..."
        git add config/cfroutesync/values.yaml
        git commit -m "Update cfroutesync image digest to ${digest}"

        git log -1 --color | cat
    else
        echo "no changes in repo, no commit necessary"
    fi
popd

cp -r cf-k8s-networking cf-k8s-networking-modified