#!/usr/bin/env bash

readonly CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly ROOT_DIR="$(dirname "$CURRENT_DIR")"

readonly tag="${1}"

git tag -d "${tag}" | true # do not fail if tag does not exist
git tag "${tag}"

directories=(examples modules)

for directory in "${directories[@]}"
do
  cd "${ROOT_DIR}/${directory}"

  ls -d */ | grep -v "_template" | while read -r module; do
    module=${module%?} # remove trailing slash
    module_tag="${directory}/${module}/${tag}" # e.g. modules/mongodb/0.0.1
    git tag -d "${module_tag}" | true # do not fail if tag does not exist
    git tag "${module_tag}"
  done
done
