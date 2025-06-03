#!/usr/bin/env bash

# This script is used to release a new version of the Testcontainers for Go library.
# It creates a tag for the root module and for each module in the modules directory,
# and then triggers the Go proxy to fetch the module. BY default, it will be run in
# dry-run mode, which will print the commands that would be executed, without actually
# executing them.
#
# Usage: ./scripts/release.sh
#
# It's possible to run the script without dry-run mode actually executing the commands.
#
# Usage: DRY_RUN="false" ./scripts/release.sh

readonly BUMP_TYPE="${BUMP_TYPE:-minor}"
readonly DOCKER_IMAGE_SEMVER="mdelapenya/semver-tool:3.4.0"
readonly DRY_RUN="${DRY_RUN:-true}"
readonly CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly ROOT_DIR="$(dirname "$CURRENT_DIR")"
readonly MKDOCS_FILE="${ROOT_DIR}/mkdocs.yml"
readonly VERSION_FILE="${ROOT_DIR}/internal/version.go"

readonly REPOSITORY="github.com/testcontainers/testcontainers-go"
readonly DIRECTORIES=(examples modules)

function main() {
  local version="$(extractCurrentVersion)"
  local vVersion="v${version}"
  echo "Current version: ${vVersion}"

  # Get the version to bump to from the semver-tool and the bump type
  local newVersion=$(docker run --rm --platform=linux/amd64 -i "${DOCKER_IMAGE_SEMVER}" bump "${BUMP_TYPE}" "${vVersion}")
  if [[ "${newVersion}" == "" ]]; then
    echo "Failed to bump the version. Please check the semver-tool image and the bump type."
    exit 1
  fi

  # Commit the project in the current state
  gitCommitVersion "${vVersion}"

  tagModule "${vVersion}"

  for directory in "${DIRECTORIES[@]}"
  do
    cd "${ROOT_DIR}/${directory}"

    ls -d */ | grep -v "_template" | while read -r module; do
      module="${module%?}" # remove trailing slash
      module_tag="${directory}/${module}/${vVersion}" # e.g. modules/mongodb/v0.0.1
      tagModule "${module_tag}"
    done
  done

  echo "Producing a ${BUMP_TYPE} bump of the version, from ${version} to ${newVersion}"

  # Bump the version in the version.go file
  if [[ "${DRY_RUN}" == "true" ]]; then
    echo "sed \"s/const Version = \".*\"/const Version = \"${newVersion}\"/g\" ${VERSION_FILE} > ${VERSION_FILE}.tmp"
    echo "mv ${VERSION_FILE}.tmp ${VERSION_FILE}"
  else
    sed "s/const Version = \".*\"/const Version = \"${newVersion}\"/g" ${VERSION_FILE} > ${VERSION_FILE}.tmp
    mv ${VERSION_FILE}.tmp ${VERSION_FILE}
  fi

  # Commit the version.go file in the next development version
  gitNextDevelopment "${newVersion}"

  # Update the remote repository with the new tags
  gitPushTags

  # Trigger the Go proxy to fetch the core module
  curlGolangProxy "${REPOSITORY}" "${vVersion}" # e.g. github.com/testcontainers/testcontainers-go/@v/v0.0.1.info

  # Trigger the Go proxy to fetch the modules
  for directory in "${DIRECTORIES[@]}"
  do
    cd "${ROOT_DIR}/${directory}"

    ls -d */ | grep -v "_template" | while read -r module; do
      module="${module%?}" # remove trailing slash
      module_path="${REPOSITORY}/${directory}/${module}"
      curlGolangProxy "${module_path}" "${vVersion}" # e.g. github.com/testcontainers/testcontainers-go/modules/mongodb/@v/v0.0.1.info
    done
  done
}

# This function is used to trigger the Go proxy to fetch the module.
# See https://pkg.go.dev/about#adding-a-package for more details.
function curlGolangProxy() {
  local module_path="${1}"
  local module_version="${2}"

  if [[ "${DRY_RUN}" == "true" ]]; then
    echo "curl https://proxy.golang.org/${module_path}/@v/${module_version}.info"
    return
  fi

  # e.g.:
  #   github.com/testcontainers/testcontainers-go/v0.0.1.info
  #   github.com/testcontainers/testcontainers-go/modules/mongodb/v0.0.1.info
  curl "https://proxy.golang.org/${module_path}/@v/${module_version}.info"
}

# This function reads the version.go file and extracts the current version.
function extractCurrentVersion() {
  cat "${VERSION_FILE}" | grep 'const Version = ' | cut -d '"' -f 2
}

# This function is used to run git commands
function gitFn() {
  args=("$@")
  if [[ "${DRY_RUN}" == "true" ]]; then
    echo "git ${args[@]}"
    return
  fi

  git "${args[@]}"
}

# This function is used to commit the version.go file, mkdocs, examples and modules.
function gitCommitVersion() {
  local version="${1}" 

  cd "${ROOT_DIR}"

  gitFn add "${VERSION_FILE}"
  gitFn add "${MKDOCS_FILE}"
  gitFn add "docs/**/*.md"
  gitFn add "examples/**/go.*"
  gitFn add "modules/**/go.*"
  gitFn commit -m "chore: use new version (${version}) in modules and examples"
}

# This function is used to commit the version.go file.
function gitNextDevelopment() {
  local newVersion="${1}" 

  cd "${ROOT_DIR}"

  gitFn add "${VERSION_FILE}"
  gitFn commit -m "chore: prepare for next ${BUMP_TYPE} development cycle (${newVersion})"
}

# This function is used to push the tags to the remote repository.
function gitPushTags() {
  gitFn push origin main --tags
}

# This function is used to create a tag for the module.
function tagModule() {
  local module_tag="${1}"

  gitFn tag -d "${module_tag}" | true # do not fail if tag does not exist
  gitFn tag "${module_tag}"
}

function validate() {
  # if bump_type is not major, minor or patch, the script will fail
  if [[ "${BUMP_TYPE}" != "major" ]] && [[ "${BUMP_TYPE}" != "minor" ]] && [[ "${BUMP_TYPE}" != "patch" ]]; then
    echo "BUMP_TYPE must be major, minor or patch. Current: ${BUMP_TYPE}"
    exit 1
  fi
}

validate
main "$@"
