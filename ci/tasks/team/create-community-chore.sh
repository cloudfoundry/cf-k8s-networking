#!/usr/bin/env bash

set -euo pipefail

# ENV
: "${TRACKER_TOKEN:?}"

project_id=2407973

echo "Fetching Template Story..."
template_story=`curl -s -f -X GET -H "X-TrackerToken: $TRACKER_TOKEN" "https://www.pivotaltracker.com/services/v5/projects/$project_id/stories/173205458"`
template_tasks=`curl -s -f -X GET -H "X-TrackerToken: $TRACKER_TOKEN" "https://www.pivotaltracker.com/services/v5/projects/$project_id/stories/173205458/tasks"`

top_of_backlog_story_id=`curl -s -f -X GET -H "X-TrackerToken: $TRACKER_TOKEN" "https://www.pivotaltracker.com/services/v5/projects/$project_id/stories?with_state=unstarted&limit=1" | jq '.[0].id'`

echo "Creating story..."
new_story=`echo $template_story | sed -E "s#YYYY/MM/DD#$(date '+%Y/%m/%d')#" | jq ". + {\"before_id\": $top_of_backlog_story_id }"`
echo $new_story
story_id=`curl -f -X POST -H "X-TrackerToken: $TRACKER_TOKEN" -H "Content-Type: application/json" -d "$new_story" "https://www.pivotaltracker.com/services/v5/projects/$project_id/stories" | jq .id`
echo $story_id
echo $template_tasks | jq -c '(.[])' | xargs -n1 -I{} curl -s -f -X POST -H "X-TrackerToken: $TRACKER_TOKEN" -H "Content-Type: application/json" -d '{}' "https://www.pivotaltracker.com/services/v5/projects/$project_id/stories/$story_id/tasks" > /dev/null

echo "Created Story id $story_id"
