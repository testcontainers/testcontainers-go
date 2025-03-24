#!/usr/bin/env bash

# This script is used to prepare a release for a new version of the Testcontainers for Go library.
# By default, it will be run in dry-run mode, which will print the commands that would be executed, without actually
# executing them.
#
# Usage: ./scripts/pre-release.sh
#
# It's possible to run the script without dry-run mode actually executing the commands.
#
# Usage: DRY_RUN="false" ./scripts/pre-release.sh

readonly DRY_RUN="${DRY_RUN:-true}"
readonly CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly ROOT_DIR="$(dirname "$CURRENT_DIR")"
readonly MKDOCS_FILE="${ROOT_DIR}/mkdocs.yml"
readonly VERSION_FILE="${ROOT_DIR}/internal/version.go"

readonly REPOSITORY="github.com/testcontainers/testcontainers-go"
readonly DIRECTORIES=(examples modules)

function main() {
  readonly version="$(extractCurrentVersion)"
  readonly vVersion="v${version}"

  gitFn checkout main
  bumpVersion "${version}"
}

# This function is used to bump the version in those files that refer to the project version.
function bumpVersion() {
  local versionToBumpWithoutV="${1}"
  local versionToBump="v${versionToBumpWithoutV}"

  # Bump version in the mkdocs descriptor file
  if [[ "${DRY_RUN}" == "true" ]]; then
    echo "sed \"s/latest_version: .*/latest_version: ${versionToBump}/g\" ${MKDOCS_FILE} > ${MKDOCS_FILE}.tmp"
    echo "mv ${MKDOCS_FILE}.tmp ${MKDOCS_FILE}"
  else
    sed "s/latest_version: .*/latest_version: ${versionToBump}/g" ${MKDOCS_FILE} > ${MKDOCS_FILE}.tmp
    mv ${MKDOCS_FILE}.tmp ${MKDOCS_FILE}
  fi

  # Bump version across all modules, in their go.mod files
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

  cd "${ROOT_DIR}/docs"

  versionEscapingDots="${versionToBumpWithoutV/./\.}"
  NON_RELEASED_STRING='Not available until the next release of testcontainers-go <a href=\"https:\/\/github.com\/testcontainers\/testcontainers-go\"><span class=\"tc-version\">:material-tag: main<\/span><\/a>'
  RELEASED_STRING="Since testcontainers-go <a href=\\\"https:\/\/github.com\/testcontainers\/testcontainers-go\/releases\/tag\/v${versionEscapingDots}\\\"><span class=\\\"tc-version\\\">:material-tag: v${versionEscapingDots}<\/span><\/a>"

  # find all markdown files, and for each of them, replace the release string
  find . -name "*.md" | while read -r module_file; do
    if [[ "${DRY_RUN}" == "true" ]]; then
      echo "sed \"s/${NON_RELEASED_STRING}/${RELEASED_STRING}/g\" ${module_file} > ${module_file}.tmp"
      echo "mv ${module_file}.tmp ${module_file}"
    else
      sed "s/${NON_RELEASED_STRING}/${RELEASED_STRING}/g" ${module_file} > ${module_file}.tmp
      mv ${module_file}.tmp ${module_file}
    fi
  done
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

main "$@"
