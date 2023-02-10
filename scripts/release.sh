#!/usr/bin/env bash

readonly CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly ROOT_DIR="$(dirname "$CURRENT_DIR")"

readonly REPOSITORY="github.com/testcontainers/testcontainers-go"
readonly TAG="${1}"

tagModule "${TAG}"

readonly DIRECTORIES=(examples modules)

for directory in "${DIRECTORIES[@]}"
do
  cd "${ROOT_DIR}/${directory}"

  ls -d */ | grep -v "_template" | while read -r module; do
    module="${module%?}" # remove trailing slash
    module_tag="${directory}/${module}/${TAG}" # e.g. modules/mongodb/v0.0.1
    tagModule "${module_tag}"
  done
done

git push --tags

curl "https://proxy.golang.org/${REPOSITORY}/@v/${TAG}" # e.g. github.com/testcontainers/testcontainers-go/v0.0.1

for directory in "${DIRECTORIES[@]}"
do
  module="${module%?}" # remove trailing slash
  module_path="${REPOSITORY}/${directory}/${module}/"
  curl "https://proxy.golang.org/${module_path}/@v/${TAG}" # e.g. # e.g. github.com/testcontainers/testcontainers-go/modules/mongodb/v0.0.1
done

function tagModule() {
  local module_tag="${1}"

  git tag -d "${module_tag}" | true # do not fail if tag does not exist
  git tag "${module_tag}"
}