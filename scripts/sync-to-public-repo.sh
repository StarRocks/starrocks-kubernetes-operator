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

# List of directories to sync
DIRS=("doc" "examples" "helm-charts" "deploy", "scripts")

echo "Starting synchronization from $HOME_PATH to $TARGET_PATH..."

for dir in "${DIRS[@]}"; do
  if [ -d "$HOME_PATH/$dir" ]; then
    echo "Processing directory: $dir"
    
    # Remove the old directory in target to ensure a clean copy
    rm -rf "$TARGET_PATH/$dir"
    
    # Copy the directory recursively
    cp -R "$HOME_PATH/$dir" "$TARGET_PATH/"
    
    echo "Successfully synced $dir"
  else
    echo "Warning: Source directory $HOME_PATH/$dir does not exist, skipping."
  fi
done

echo "---------------------------------------------------"
echo "Synchronization completed successfully!"
echo "Target: $TARGET_PATH"
