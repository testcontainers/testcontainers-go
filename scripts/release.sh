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

gitPushTags

curlGolangProxy "${REPOSITORY}" # e.g. github.com/testcontainers/testcontainers-go/@v/v0.0.1

for directory in "${DIRECTORIES[@]}"
do
  module="${module%?}" # remove trailing slash
  module_path="${REPOSITORY}/${directory}/${module}/"
  curlGolangProxy "${module_path}" # e.g. github.com/testcontainers/testcontainers-go/modules/mongodb/@v/v0.0.1
done

function curlGolangProxy() {
  local module_path="${1}"

  # e.g.:
  #   github.com/testcontainers/testcontainers-go/v0.0.1
  #   github.com/testcontainers/testcontainers-go/modules/mongodb/v0.0.1
  curl "https://proxy.golang.org/${module_path}/@v/${TAG}"
}

function gitPushTags() {
  git push --tags
}

function tagModule() {
  local module_tag="${1}"

  git tag -d "${module_tag}" | true # do not fail if tag does not exist
  git tag "${module_tag}"
}