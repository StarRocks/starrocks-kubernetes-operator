#!/bin/bash

# input parameters: a list of PR numbers and a version tag
# e.g. v1.8.6 1 2 3 4 5
VERSION_TAG=$1

# create label by gh
gh label create $VERSION_TAG

shift
PR_NUMBERS=("$@")

# define an array of keywords
KEYWORDS=("feature" "enhancement" "bugfix" "chore" "documentation")

# iterate over the list of PR numbers
for PR_NUMBER in "${PR_NUMBERS[@]}"; do
  # add a version label to each PR
  gh pr edit $PR_NUMBER --add-label "$VERSION_TAG"

  # get the PR title
  PR_TITLE=$(gh pr view $PR_NUMBER --json title -q '.title')

  # convert PR title to lowercase
  PR_TITLE=$(echo "$PR_TITLE" | tr '[:upper:]' '[:lower:]')

  # check if the PR title contains any of the keywords
  for KEYWORD in "${KEYWORDS[@]}"; do
    if [[ $PR_TITLE == *"$KEYWORD"* ]]; then
      # create the label if it doesn't exist
      gh label create $KEYWORD 2>/dev/null
      # add the label to the PR
      gh pr edit $PR_NUMBER --add-label "$KEYWORD"
    fi
  done
done
