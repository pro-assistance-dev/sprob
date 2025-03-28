#!/bin/bash

last_tag=$(git tag | sort -V | tail -n 1)
echo "$last_tag"

readarray -td. versions_part <<<"$last_tag"

minor="${versions_part[2]}"

minor=$((minor + 1))

new_tag="${versions_part[0]}.${versions_part[1]}.${minor}"
echo "$new_tag"

git add .
git commit -m "$new_tag"
git push
git tag -a $new_tag -m "$new_tag"
git push --tags
