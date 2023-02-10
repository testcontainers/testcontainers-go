#!/usr/bin/env bash

readonly CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly ROOT_DIR="$(dirname "$CURRENT_DIR")"

readonly repository="github.com/testcontainers/testcontainers-go"
readonly tag="${1}"

git tag -d "${tag}" | true # do not fail if tag does not exist
git tag "${tag}"

directories=(examples modules)

for directory in "${directories[@]}"
do
  cd "${ROOT_DIR}/${directory}"

  ls -d */ | grep -v "_template" | while read -r module; do
    module="${module%?}" # remove trailing slash
    module_tag="${directory}/${module}/${tag}" # e.g. modules/mongodb/v0.0.1
    git tag -d "${module_tag}" | true # do not fail if tag does not exist
    git tag "${module_tag}"
  done
done

git push --tags

curl "https://proxy.golang.org/${repository}/@v/${tag}" # e.g. github.com/testcontainers/testcontainers-go/v0.0.1

for directory in "${directories[@]}"
do
  module="${module%?}" # remove trailing slash
  module_path="${repository}/${directory}/${module}/"
  curl "https://proxy.golang.org/${module_path}/@v/${tag}" # e.g. # e.g. github.com/testcontainers/testcontainers-go/modules/mongodb/v0.0.1
done
