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

readonly DOCKER_IMAGE_SEMVER="docker.io/mdelapenya/semver-tool:3.4.0"
readonly COMMIT="${COMMIT:-false}"
readonly DRY_RUN="${DRY_RUN:-true}"
readonly CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly ROOT_DIR="$(dirname "$CURRENT_DIR")"
readonly MKDOCS_FILE="${ROOT_DIR}/mkdocs.yml"
readonly VERSION_FILE="${ROOT_DIR}/internal/version.go"
readonly BUMP_TYPE="${BUMP_TYPE:-minor}"

readonly REPOSITORY="github.com/testcontainers/testcontainers-go"
readonly DIRECTORIES=(examples modules)

function main() {
  readonly version="$(extractCurrentVersion)"
  readonly vVersion="v${version}"

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

  gitState
  bumpVersion "${version}"
  gitPushTags
  gitUnstash

  curlGolangProxy "${REPOSITORY}" "${vVersion}" # e.g. github.com/testcontainers/testcontainers-go/@v/v0.0.1

  for directory in "${DIRECTORIES[@]}"
  do
    cd "${ROOT_DIR}/${directory}"

    ls -d */ | grep -v "_template" | while read -r module; do
      module="${module%?}" # remove trailing slash
      module_path="${REPOSITORY}/${directory}/${module}"
      curlGolangProxy "${module_path}" "${vVersion}" # e.g. github.com/testcontainers/testcontainers-go/modules/mongodb/@v/v0.0.1
    done
  done
}

# This function is used to bump the version in the version.go file and in the mkdocs.yml file.
function bumpVersion() {
  local versionToBumpWithoutV="${1}"
  local versionToBump="v${versionToBumpWithoutV}"

  local newVersion=$(docker run --rm "${DOCKER_IMAGE_SEMVER}" bump "${BUMP_TYPE}" "${versionToBump}")
  echo "Producing a ${BUMP_TYPE} bump of the version, from ${versionToBump} to ${newVersion}"

  if [[ "${DRY_RUN}" == "true" ]]; then
    echo "sed \"s/const Version = \".*\"/const Version = \"${newVersion}\"/g\" ${VERSION_FILE} > ${VERSION_FILE}.tmp"
    echo "mv ${VERSION_FILE}.tmp ${VERSION_FILE}"
  else
    sed "s/const Version = \".*\"/const Version = \"${newVersion}\"/g" ${VERSION_FILE} > ${VERSION_FILE}.tmp
    mv ${VERSION_FILE}.tmp ${VERSION_FILE}
  fi

  if [[ "${DRY_RUN}" == "true" ]]; then
    echo "sed \"s/latest_version: .*/latest_version: ${versionToBump}/g\" ${MKDOCS_FILE} > ${MKDOCS_FILE}.tmp"
    echo "mv ${MKDOCS_FILE}.tmp ${MKDOCS_FILE}"
  else
    sed "s/latest_version: .*/latest_version: ${versionToBump}/g" ${MKDOCS_FILE} > ${MKDOCS_FILE}.tmp
    mv ${MKDOCS_FILE}.tmp ${MKDOCS_FILE}
  fi

  for directory in "${DIRECTORIES[@]}"
  do
    cd "${ROOT_DIR}/${directory}"

    ls -d */ | grep -v "_template" | while read -r module; do
      module="${module%?}" # remove trailing slash
      module_mod_file="${module}/go.mod" # e.g. modules/mongodb/go.mod
      if [[ "${DRY_RUN}" == "true" ]]; then
        echo "sed \"s/testcontainers-go v.*/testcontainers-go v${versionToBumpWithoutV}/g\" ${module_mod_file} > ${module_mod_file}.tmp"
        echo "mv ${module_mod_file}.tmp ${module_mod_file}"
      else
        sed "s/testcontainers-go v.*/testcontainers-go v${versionToBumpWithoutV}/g" ${module_mod_file} > ${module_mod_file}.tmp
        mv ${module_mod_file}.tmp ${module_mod_file}
      fi
    done

    make "tidy-${directory}"
  done

  gitCommitVersion "${newVersion}"
}

# This function is used to trigger the Go proxy to fetch the module.
# See https://pkg.go.dev/about#adding-a-package for more details.
function curlGolangProxy() {
  local module_path="${1}"
  local module_version="${2}"

  if [[ "${DRY_RUN}" == "true" || "${COMMIT}" == "false" ]]; then
    echo "curl https://proxy.golang.org/${module_path}/@v/${module_version}"
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
  if [[ "${DRY_RUN}" == "true" || "${COMMIT}" == "false" ]]; then
    echo "git ${args[@]}"
    return
  fi

  git "${args[@]}"
}

# This function is used to commit the version.go file.
function gitCommitVersion() {
  local newVersion="${1}" 

  cd "${ROOT_DIR}"

  gitFn add "${VERSION_FILE}"
  gitFn add "${MKDOCS_FILE}"
  gitFn add "examples/**/go.*"
  gitFn add "modules/**/go.*"
  gitFn commit -m "chore: prepare for next ${BUMP_TYPE} development cycle (${newVersion})"
}

# This function is used to push the tags to the remote repository.
function gitPushTags() {
  gitFn push origin main --tags
}

# This function is setting the git state to the next development cycle:
# - Stashing the changes
# - Moving to the main branch
function gitState() {
  gitFn stash
  gitFn checkout main
}

function gitUnstash() {
  gitFn stash pop
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
