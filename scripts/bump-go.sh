#!/usr/bin/env bash

# Bumps the minimal Go version (major.minor) across the project:
# - go.mod files: "go" and "toolchain" directives.
# - CI docs (docs/system_requirements/ci), workflows matrix, test Dockerfiles, devcontainer, AI.md.
#
# Derived automatically: current version (root go.mod), next minor, latest patch releases
# (Go module proxy; override with GO_PATCH / GO_NEXT_PATCH) and the golang:X.Y-alpine digest
# (docker; override with GOLANG_IMAGE_DIGEST).
#
# Dry-run by default, printing the commands:   ./scripts/bump-go.sh "1.26"
# Apply the changes:                            DRY_RUN="false" ./scripts/bump-go.sh "1.26"
#
# Afterwards: run "make tidy-all", check golangci-lint supports the new Go version
# (commons-test.mk, .github/workflows/ci-lint-go.yml) and review Go checks in ci-test-go.yml.

set -euo pipefail

readonly CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly DRY_RUN="${DRY_RUN:-true}"
readonly ROOT_DIR="$(dirname "$CURRENT_DIR")"
readonly GO_MOD_FILE="${ROOT_DIR}/go.mod"
readonly DEVCONTAINER_IMAGE_PREFIX="go:"

function main() {
  local goVersion="${1:-}"
  if [[ ! "${goVersion}" =~ ^[0-9]+\.[0-9]+$ ]]; then
    echo "Usage: $0 <major.minor>  (e.g. $0 1.26)" >&2
    exit 1
  fi

  echo "Updating Go version:"

  local currentGoVersion="$(extractCurrentVersion)"
  local currentMajorMinor="$(echo "${currentGoVersion}" | grep -oE '^[0-9]+\.[0-9]+')"
  echo " - Current: ${currentGoVersion}"

  local nextGoVersion="${goVersion%.*}.$(( ${goVersion#*.} + 1 ))"
  echo " - New: ${goVersion}"
  echo " - Next: ${nextGoVersion}"

  local goPatch="${GO_PATCH:-$(latestPatch "${goVersion}")}"
  local goNextPatch="${GO_NEXT_PATCH:-$(latestPatch "${nextGoVersion}")}"
  echo " - Latest patch releases: ${goPatch}, ${goNextPatch}"

  local imageDigest="${GOLANG_IMAGE_DIGEST:-$(golangImageDigest "${goVersion}")}"
  echo " - golang:${goVersion}-alpine digest: ${imageDigest:-<not resolved, keeping the existing one>}"

  # all the files but go.mod reference the major.minor version, so the patterns use it
  local escapedCurrentGoVersion="$(escape "${currentMajorMinor}")"
  local escapedGoVersion="$(escape "${goVersion}")"

  # bump mod files in all the modules
  for modFile in $(find "${ROOT_DIR}" -name "go.mod" -not -path "${ROOT_DIR}/vendor/*" -not -path "${ROOT_DIR}/.git/*"); do
    bumpModFile "${modFile}" "${escapedCurrentGoVersion}" "${goVersion}" "${goPatch}"
  done

  # bump markdown files
  for f in $(find "${ROOT_DIR}" -name "*.md" -not -path "${ROOT_DIR}/.git/*"); do
    bumpGolangDockerImages "${f}" "${escapedCurrentGoVersion}" "${goVersion}" "${goPatch}" "${goNextPatch}"
  done

  # bump github action workflows
  for f in $(find "${ROOT_DIR}/.github/workflows" -name "*.yml"); do
    bumpCIMatrix "${f}" "${escapedCurrentGoVersion}" "${escapedGoVersion}" "${goVersion}" "${nextGoVersion}"
  done

  # bump Dockerfiles used in tests
  for f in $(find "${ROOT_DIR}" \( -name "Dockerfile" -o -name "*.Dockerfile" \) -not -path "${ROOT_DIR}/.git/*"); do
    bumpTestDockerfile "${f}" "${goVersion}" "${imageDigest}"
  done

  # bump devcontainer file
  bumpDevcontainer "${ROOT_DIR}/.devcontainer/devcontainer.json" "${escapedCurrentGoVersion}" "${goVersion}"

  # bump the contributors guide for AI agents
  bumpAIGuide "${ROOT_DIR}/AI.md" "${escapedCurrentGoVersion}" "${goPatch}"

  echo ""
  echo "Done. Please run 'make tidy-all', and verify the golangci-lint version supports Go ${goVersion}."
}

# escapes the dots of a version so it can be used in a sed regular expression.
function escape() {
  echo "${1}" | sed 's/\./\\./g'
}

# runs the given sed expression in-place on the given file, or prints it in dry-run mode.
function replace() {
  local file="${1}"
  shift

  if [[ ! -f "${file}" ]]; then
    return
  fi

  if [[ "${DRY_RUN}" == "true" ]]; then
    local expr
    for expr in "$@"; do
      echo "sed -E \"${expr}\" ${file} > ${file}.tmp && mv ${file}.tmp ${file}"
    done
  else
    local expr
    for expr in "$@"; do
      sed -E "${expr}" "${file}" > "${file}.tmp"
      mv "${file}.tmp" "${file}"
    done
  fi
}

# it will replace the 'go-version: [${oldGoVersion}.x, ${newGoVersion}.x]' with
# 'go-version: [${newGoVersion}.x, ${nextGoVersion}.x]' in the given file,
# and the references to the lowest version in the matrix, used to decide if Sonar must be run.
function bumpCIMatrix() {
  local file="${1}"
  local oldGoVersion="${2}"
  local escapedNewGoVersion="${3}"
  local newGoVersion="${4}"
  local nextGoVersion="${5}"

  replace "${file}" \
    "s/go-version: \[${oldGoVersion}\.x, ${escapedNewGoVersion}\.x\]/go-version: [${newGoVersion}.x, ${nextGoVersion}.x]/g" \
    "s/go-version: \[${oldGoVersion}\.x, 1\.x\]/go-version: [${newGoVersion}.x, 1.x]/g" \
    "s/go-version: \"${oldGoVersion}\.x\"/go-version: \"${newGoVersion}.x\"/g" \
    "s/\"${oldGoVersion}\.x\" ==/\"${newGoVersion}.x\" ==/g" \
    "s/go-version == '\"${oldGoVersion}\.x\"'/go-version == '\"${newGoVersion}.x\"'/g"
}

# it will replace the 'go:${oldGoVersion}-trixie' with 'go:${newGoVersion}-trixie' in the given file
function bumpDevcontainer() {
  local file="${1}"
  local oldGoVersion="${2}"
  local newGoVersion="${3}"

  replace "${file}" "s/${DEVCONTAINER_IMAGE_PREFIX}${oldGoVersion}-/${DEVCONTAINER_IMAGE_PREFIX}${newGoVersion}-/g"
}

# it will replace the 'golang:${oldGoVersion}' with 'golang:${newGoVersion}' in the given file,
# plus the YAML variants used by the different CI systems, and the explicit patch versions used by CircleCI.
function bumpGolangDockerImages() {
  local file="${1}"
  local oldGoVersion="${2}"
  local newGoVersion="${3}"
  local goPatch="${4}"
  local goNextPatch="${5}"

  replace "${file}" \
    "s/golang:${oldGoVersion}([^0-9.]|$)/golang:${newGoVersion}\1/g" \
    "s/golang: ${oldGoVersion}([^0-9.]|$)/golang: ${newGoVersion}\1/g" \
    "s/- \"${oldGoVersion}\"/- \"${newGoVersion}\"/g" \
    "s/(minimal version of Go required to use this module is )\*\*[0-9]+\.[0-9]+\*\*/\1**${newGoVersion}**/g" \
    "s/\[\"${oldGoVersion}\.[0-9]+\", \"[0-9]+\.[0-9]+\.[0-9]+\"\]/[\"${goPatch}\", \"${goNextPatch}\"]/g"
}

# it will replace the 'golang:X.Y-alpine@sha256:...' with 'golang:${newGoVersion}-alpine@sha256:${digest}' in the given file.
# If the digest could not be resolved, only the tag is updated and the digest must be refreshed manually.
function bumpTestDockerfile() {
  local file="${1}"
  local newGoVersion="${2}"
  local digest="${3}"

  if ! grep -qE "^FROM golang:[0-9]+\.[0-9]+-alpine" "${file}"; then
    return
  fi

  if [[ -n "${digest}" ]]; then
    replace "${file}" "s/^FROM golang:[0-9]+\.[0-9]+-alpine@sha256:[0-9a-f]+/FROM golang:${newGoVersion}-alpine@${digest}/"
  else
    echo "WARN: could not resolve the digest for golang:${newGoVersion}-alpine, please update it manually in ${file}" >&2
    replace "${file}" "s/^FROM golang:[0-9]+\.[0-9]+-alpine/FROM golang:${newGoVersion}-alpine/"
  fi
}

# it will replace the 'go ${oldGoVersion}[.N]' with 'go ${newGoVersion}.0' and the 'toolchain go${oldGoVersion}.N'
# with 'toolchain go${goPatch}' in the given go.mod file, where oldGoVersion is major.minor. Submodules pinned to
# a patch release of the old version (e.g. 'go 1.25.7') are bumped as well.
function bumpModFile() {
  local goModFile="${1}"
  local oldGoVersion="${2}"
  local newGoVersion="${3}"
  local goPatch="${4}"

  replace "${goModFile}" \
    "s/^go ${oldGoVersion}(\.[0-9]+)?$/go ${newGoVersion}.0/" \
    "s/^toolchain go${oldGoVersion}\.[0-9]+$/toolchain go${goPatch}/"
}

# it will replace the Go version used by the contributors in the AI guide, in the form 'Go 1.25.9' or 'gvm 1.25.9'.
function bumpAIGuide() {
  local file="${1}"
  local oldGoVersion="${2}"
  local goPatch="${3}"

  replace "${file}" "s/(Go|gvm) ${oldGoVersion}\.[0-9]+/\1 ${goPatch}/g"
}

# This function reads the root go.mod file and extracts the current version.
function extractCurrentVersion() {
  grep '^go .*' "${GO_MOD_FILE}" | sed 's/^go //g' | head -n 1
}

# This function returns the latest patch release of the given major.minor Go version, using the Go module proxy.
function latestPatch() {
  local majorMinor="${1}"

  local latest
  latest="$(go list -m -versions go 2>/dev/null | tr ' ' '\n' | grep -E "^${majorMinor//./\\.}\.[0-9]+$" | tail -n 1 || true)"
  if [[ -z "${latest}" ]]; then
    echo "WARN: could not resolve the latest patch release of Go ${majorMinor}, using ${majorMinor}.0" >&2
    latest="${majorMinor}.0"
  fi

  echo "${latest}"
}

# This function returns the digest of the golang:X.Y-alpine multi-platform image, using docker.
function golangImageDigest() {
  local majorMinor="${1}"

  docker buildx imagetools inspect "golang:${majorMinor}-alpine" 2>/dev/null | grep -m1 '^Digest:' | awk '{print $2}' || true
}

main "$@"
