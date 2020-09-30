#!/usr/bin/env bash
set -euo pipefail

# ENV
: "${GITHUB_KEY:?}"
: "${GITHUB_TITLE:?}"
: "${GITHUB_BODY:?}"
: "${BRANCH:?}"

# Create PR
pull=$(curl \
  -sS \
  --fail \
  -X POST \
  -H "Accept: application/vnd.github.v3+json" \
  -H "Authorization: token ${GITHUB_KEY}" \
  https://api.github.com/repos/cloudfoundry/cf-for-k8s/pulls \
  -d '{
    "head":"'"$BRANCH"'",
    "base":"develop",
    "maintainer_can_modify": true,
    "title": "'"$GITHUB_TITLE"'",
    "body": "'"$GITHUB_BODY"'"
  }')

pull_number="$(echo "${pull}" | jq -r '.number')"

# Add networking label to run our CI job for tests
curl \
  --fail \
  -X POST \
  -H "Accept: application/vnd.github.v3+json" \
  -H "Authorization: token ${GITHUB_KEY}" \
  "https://api.github.com/repos/cloudfoundry/cf-for-k8s/issues/${pull_number}/labels" \
  -d '{"labels":["networking"]}'

