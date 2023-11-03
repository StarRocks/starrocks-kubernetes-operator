#!/bin/bash

# export your Personal Access Token (PAT) with name GH_TOKEN
# run this script with the PR number as the first argument
# e.g. ./scripts/add-function-label-to-pr.sh 123

# Set the owner, repo, and PR number
OWNER="StarRocks"
REPO="starrocks-kubernetes-operator"
PR_NUMBER=$1

# Get the PR's commits
commits=$(gh api -X GET /repos/$OWNER/$REPO/pulls/$PR_NUMBER/commits --paginate -H "Authorization: Bearer $GH_TOKEN" | jq -r '.[].commit.message')

labels=()

# Check the commit messages for keywords and add labels accordingly
for commit in $commits; do
  title=$(echo $commit | head -n1 | tr '[:upper:]' '[:lower:]')
  if [[ $title == *"feature"* ]]; then
    labels+=('feature')
  fi
  if [[ $title == *"enhancement"* ]]; then
    labels+=('enhancement')
  fi
  if [[ $title == *"bugfix"* ]]; then
    labels+=('bugfix')
  fi
  if [[ $title == *"chore"* ]]; then
    labels+=('chore')
  fi
  if [[ $title == *"documentation"* ]]; then
    labels+=('documentation')
  fi
done

# Remove duplicate labels
labels=($(printf "%s\n" "${labels[@]}" | sort -u))

# Add the labels to the PR
for label in "${labels[@]}"; do
  done=$(gh api -X POST /repos/$OWNER/$REPO/issues/$PR_NUMBER/labels -H "Authorization: Bearer $GH_TOKEN" -f labels[]=$label)
  echo $done
done
