#!/bin/bash

# include common.sh to get project root
source "$(dirname "$0")/common.sh"
HOME_PATH=$(printHomePath)

# Define the target public repository directory
TARGET_PATH="${HOME_PATH}/../celerdata-kubernetes-operator"

# Check if the target directory exists
if [ ! -d "$TARGET_PATH" ]; then
  echo "Error: Target directory $TARGET_PATH does not exist."
  echo "Please make sure you have cloned the public repository in the same parent directory."
  exit 1
fi

# List of items to sync (directories or files)
ITEMS=("doc" "examples" "helm-charts" "deploy" "scripts" "README.md" "README_ZH-CN.md")

echo "Starting synchronization from $HOME_PATH to $TARGET_PATH..."

for item in "${ITEMS[@]}"; do
  if [ -e "$HOME_PATH/$item" ]; then
    echo "Processing: $item"
    
    # Remove the old item in target to ensure a clean copy
    rm -rf "$TARGET_PATH/$item"
    
    # Copy the item recursively
    cp -R "$HOME_PATH/$item" "$TARGET_PATH/"
    
    echo "Successfully synced $item"
  else
    echo "Warning: Source $HOME_PATH/$item does not exist, skipping."
  fi
done

echo "---------------------------------------------------"
echo "Synchronization completed successfully!"
echo "Target: $TARGET_PATH"
