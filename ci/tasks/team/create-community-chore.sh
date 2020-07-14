#!/usr/bin/env bash

set -euo pipefail

# ENV
: "${TRACKER_TOKEN:?}"

project_id=2407973

echo "Fetching Template Story..."
template_story=`curl -s -X GET -H "X-TrackerToken: $TRACKER_TOKEN" "https://www.pivotaltracker.com/services/v5/projects/$project_id/stories/173205458" | sed -E "s|YYYY/MM/DD|$(date '+%Y/%m/%d')|"`
template_tasks=`curl -s -X GET -H "X-TrackerToken: $TRACKER_TOKEN" "https://www.pivotaltracker.com/services/v5/projects/$project_id/stories/173205458/tasks"`

echo "Creating story..."
story_id=`curl -s -X POST -H "X-TrackerToken: $TRACKER_TOKEN" -H "Content-Type: application/json" -d "$template_story" "https://www.pivotaltracker.com/services/v5/projects/$project_id/stories" | jq .id`
echo $template_tasks | jq -c '(.[])' | xargs -L1 -d$'\n' -I{} curl -s -X POST -H "X-TrackerToken: $TRACKER_TOKEN" -H "Content-Type: application/json" -d '{}' "https://www.pivotaltracker.com/services/v5/projects/$project_id/stories/$story_id/tasks" > /dev/null

echo "Created Story id $story_id"
