#!/bin/bash

# input parameters: a list of PR numbers and a version tag
# e.g. v1.8.5 1 2 3 4 5
VERSION_TAG=$1

# create label by gh
gh label create $VERSION_TAG

shift
PR_NUMBERS=("$@")

# iterate over the list of PR numbers
for PR_NUMBER in "${PR_NUMBERS[@]}"; do
  # add a version label to each PR
  gh pr edit $PR_NUMBER --add-label "$VERSION_TAG"
done
