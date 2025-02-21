#! /bin/bash

# You should use it like this: bash ./scripts/change-version.sh 1.9.8 1.10 1.10.09.9

find . -type f | while read -r file; do
  # ignore files or directories starting with a dot
  [[ $file == *"/."* ]] && continue

  # ignore vendor directory
  [[ $file == *"/vendor/"* ]] && continue

  # ignore CHANGELOG file
  [[ $file == *"/CHANGELOG.md" ]] && continue

  echo $file
  sed -i '' "s/$1/$2/g" "$file"
done
