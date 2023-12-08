#! /bin/bash

# set the home path
function printHomePath() {
  HOME_PATH=$(
    cd "$(dirname "$0")"
    cd ..
    pwd
  )
  echo "${HOME_PATH}"
}
