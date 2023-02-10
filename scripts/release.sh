#!/usr/bin/env bash

# This script is used to release a new version of the Testcontainers for Go library.
# It creates a tag for the root module and for each module in the modules directory,
# and then triggers the Go proxy to fetch the module.
#
# Usage: ./scripts/release.sh "v0.0.1"
# where v0.0.1 is the tag to be created.
#
# It's possible to run the script in dry-run mode, which will print the commands that
# would be executed, without actually executing them.
#
# Usage: DRY_RUN="true" ./scripts/release.sh "v0.0.1" 

readonly DRY_RUN="${DRY_RUN:-false}"
readonly CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly ROOT_DIR="$(dirname "$CURRENT_DIR")"

readonly REPOSITORY="github.com/testcontainers/testcontainers-go"
readonly TAG="${1}"

function main() {
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
    cd "${ROOT_DIR}/${directory}"

    ls -d */ | grep -v "_template" | while read -r module; do
      module="${module%?}" # remove trailing slash
      module_path="${REPOSITORY}/${directory}/${module}"
      curlGolangProxy "${module_path}" # e.g. github.com/testcontainers/testcontainers-go/modules/mongodb/@v/v0.0.1
    done
  done
}

# This function is used to trigger the Go proxy to fetch the module.
function curlGolangProxy() {
  local module_path="${1}"

  if [[ "${DRY_RUN}" == "true" ]]; then
    echo "curl -X POST https://proxy.golang.org/${module_path}/@v/${TAG}"
    return
  fi

  # e.g.:
  #   github.com/testcontainers/testcontainers-go/v0.0.1
  #   github.com/testcontainers/testcontainers-go/modules/mongodb/v0.0.1
  curl "https://proxy.golang.org/${module_path}/@v/${TAG}"
}

# This function is used to push the tags to the remote repository.
function gitPushTags() {
  if [[ "${DRY_RUN}" == "true" ]]; then
    echo "git push --tags"
    return
  fi

  git push --tags
}

# This function is used to create a tag for the module.
function tagModule() {
  local module_tag="${1}"

  if [[ "${DRY_RUN}" == "true" ]]; then
    echo "git tag -d ${module_tag} | true"
    echo "git tag ${module_tag}"
    return
  fi

  git tag -d "${module_tag}" | true # do not fail if tag does not exist
  git tag "${module_tag}"
}

main "$@"
