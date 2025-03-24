#!/usr/bin/env bash

# This script is used to bump the version of Go in Testcontainers for Go,
# modifying the following files:
# - go.mod in the core module and submodules.
# - Markdown files explaining how to use Testcontainers for Go in the different CI systems.
# - Github action workflows using a test matrix to test Testcontainers for Go in different versions of Go.
# - Templates for the module generator.
# - Devcontainer file for VSCode.
#
# By default, it will be run in dry-run mode, which will print the commands that would be executed, without actually
# executing them.
#
# Usage: ./scripts/bump-go.sh "1.20"
#
# It's possible to run the script without dry-run mode actually executing the commands.
#
# Usage: DRY_RUN="false" ./scripts/bump-go.sh "1.20"

readonly CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly DRY_RUN="${DRY_RUN:-true}"
readonly ROOT_DIR="$(dirname "$CURRENT_DIR")"
readonly GO_MOD_FILE="${ROOT_DIR}/go.mod"
readonly DEVCONTAINER_IMAGE_PREFIX="go:"

function main() {
  echo "Updating Go version:"

  local currentGoVersion="$(extractCurrentVersion)"
  echo " - Current: ${currentGoVersion}"
  local escapedCurrentGoVersion="$(echo "${currentGoVersion}" | sed 's/\./\\./g')"

  local goVersion="${1}"
  local escapedGoVersion="$(echo "${goVersion}" | sed 's/\./\\./g')"
  echo " - New: ${goVersion}"

  # bump mod files in all the modules
  for modFile in $(find "${ROOT_DIR}" -name "go.mod" -not -path "${ROOT_DIR}/vendor/*" -not -path "${ROOT_DIR}/.git/*"); do
    bumpModFile "${modFile}" "${escapedCurrentGoVersion}" "${escapedGoVersion}"
  done

  # bump markdown files
  for f in $(find "${ROOT_DIR}" -name "*.md"); do
    bumpGolangDockerImages "${f}" "${escapedCurrentGoVersion}" "${escapedGoVersion}"
  done

  # bump github action workflows
  for f in $(find "${ROOT_DIR}/.github/workflows" -name "*.yml"); do
    bumpCIMatrix "${f}" "${escapedCurrentGoVersion}" "${escapedGoVersion}"
  done

  # bump devcontainer file
  bumpDevcontainer "${ROOT_DIR}/.devcontainer/devcontainer.json" "${escapedCurrentGoVersion}" "${escapedGoVersion}"
}

# it will replace the 'go-version: [${oldGoVersion}, 1.x]' with 'go-version: [${newGoVersion}, 1.x]' in the given file
function bumpCIMatrix() {
  local file="${1}"
  local oldGoVersion="${2}"
  local newGoVersion="${3}"

  if [[ "${DRY_RUN}" == "true" ]]; then
    echo "sed \"s/go-version: \[${oldGoVersion}/go-version: \[${newGoVersion}/g\" ${file} > ${file}.tmp"
    echo "sed \"s/go-version: \\"${oldGoVersion}.x\"/go-version: \\"${newGoVersion}.x\"/g\" ${file}.tmp > ${file}.tmp"
    echo "sed \"s/go-version == '\"${oldGoVersion}\.x\"'/go-version == '\"${newGoVersion}\.x\"'/g\" ${file}.tmp > ${file}.tmp"
    echo "mv ${file}.tmp ${file}"
  else
    sed "s/go-version: \[${oldGoVersion}/go-version: \[${newGoVersion}/g" ${file} > ${file}.tmp
    mv ${file}.tmp ${file}
    sed "s/go-version: \"${oldGoVersion}\.x\"/go-version: \"${newGoVersion}\.x\"/g" ${file} > ${file}.tmp
    mv ${file}.tmp ${file}
    sed "s/go-version == '\"${oldGoVersion}\.x\"'/go-version == '\"${newGoVersion}\.x\"'/g" ${file} > ${file}.tmp
    mv ${file}.tmp ${file}
  fi
}

# it will replace the 'go:0-${oldGoVersion}-bullseye' with 'go:0-${newGoVersion}-bullseye' in the given file
function bumpDevcontainer() {
  local file="${1}"
  local oldGoVersion="${2}"
  local newGoVersion="${3}"

  if [[ "${DRY_RUN}" == "true" ]]; then
    echo "sed \"s/${DEVCONTAINER_IMAGE_PREFIX}${oldGoVersion}/${DEVCONTAINER_IMAGE_PREFIX}${newGoVersion}/g\" ${file} > ${file}.tmp"
    echo "mv ${file}.tmp ${file}"
  else
    sed "s/${DEVCONTAINER_IMAGE_PREFIX}${oldGoVersion}/${DEVCONTAINER_IMAGE_PREFIX}${newGoVersion}/g" ${file} > ${file}.tmp
    mv ${file}.tmp ${file}
  fi
}

# it will replace the 'golang:${oldGoVersion}' with 'golang:${newGoVersion}' in the given file
function bumpGolangDockerImages() {
  local file="${1}"
  local oldGoVersion="${2}"
  local newGoVersion="${3}"

  if [[ "${DRY_RUN}" == "true" ]]; then
    echo "sed \"s/golang:${oldGoVersion}/golang:${newGoVersion}/g\" ${file} > ${file}.tmp"
    echo "sed \"s/golang: ${oldGoVersion}/golang: ${newGoVersion}/g\" ${file}.tmp > ${file}.tmp"
    echo "sed \"s/- \\"${oldGoVersion}\\"/- \\"${newGoVersion}\\"/g\" ${file}.tmp > ${file}.tmp"
    echo "mv ${file}.tmp ${file}"
  else
    sed "s/golang:${oldGoVersion}/golang:${newGoVersion}/g" ${file} > ${file}.tmp
    mv ${file}.tmp ${file}
    sed "s/golang: ${oldGoVersion}/golang: ${newGoVersion}/g" ${file} > ${file}.tmp
    mv ${file}.tmp ${file}
    sed "s/- \"${oldGoVersion}\"/- \"${newGoVersion}\"/g" ${file} > ${file}.tmp
    mv ${file}.tmp ${file}
  fi
}

# it will replace the 'go ${oldGoVersion}' with 'go ${newGoVersion}' in the given go.mod file
function bumpModFile() {
  local goModFile="${1}"
  local oldGoVersion="${2}"
  local newGoVersion="${3}"

  if [[ "${DRY_RUN}" == "true" ]]; then
    echo "sed \"s/^go ${oldGoVersion}/go ${newGoVersion}/g\" ${goModFile} > ${goModFile}.tmp"
    echo "mv ${goModFile}.tmp ${goModFile}"
  else
    sed "s/^go ${oldGoVersion}/go ${newGoVersion}/g" ${goModFile} > ${goModFile}.tmp
    mv ${goModFile}.tmp ${goModFile}
  fi
}

# This function reads the reaper.go file and extracts the current version.
function extractCurrentVersion() {
  cat "${GO_MOD_FILE}" | grep '^go .*' | sed 's/^go //g' | head -n 1
}

main "$@"
