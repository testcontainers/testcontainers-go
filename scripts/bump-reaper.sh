#!/usr/bin/env bash

# This script is used to bump the version of Ryuk in Testcontainers for Go,
# modifying the "reaper.go" file. By default, it will be run in
# dry-run mode, which will print the commands that would be executed, without actually
# executing them.
#
# Usage: ./scripts/bump-reaper.sh "docker.io/testcontainers/ryuk:1.2.3"
#
# It's possible to run the script without dry-run mode actually executing the commands.
#
# Usage: DRY_RUN="false" ./scripts/bump-reaper.sh "docker.io/testcontainers/ryuk:1.2.3"

readonly CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly DRY_RUN="${DRY_RUN:-true}"
readonly ROOT_DIR="$(dirname "$CURRENT_DIR")"
readonly REAPER_FILE="${ROOT_DIR}/internal/config/config.go"

function main() {
  echo "Updating Ryuk version:"

  local currentVersion="$(extractCurrentVersion)"
  echo " - Current: ${currentVersion}"

  local ryukVersion="${1}"
  local escapedRyukVersion="${ryukVersion//\//\\/}"
  echo " - New: ${ryukVersion}"

  # Bump the version in the config.go file
  if [[ "${DRY_RUN}" == "true" ]]; then
    echo "sed \"s/ReaperDefaultImage = \".*\"/ReaperDefaultImage = \"${escapedRyukVersion}\"/g\" ${REAPER_FILE} > ${REAPER_FILE}.tmp"
    echo "mv ${REAPER_FILE}.tmp ${REAPER_FILE}"
  else
    # replace using sed the version in the config.go file
    sed "s/ReaperDefaultImage = \".*\"/ReaperDefaultImage = \"${escapedRyukVersion}\"/g" ${REAPER_FILE} > ${REAPER_FILE}.tmp
    mv ${REAPER_FILE}.tmp ${REAPER_FILE}
  fi
}

# This function reads the reaper.go file and extracts the current version.
function extractCurrentVersion() {
  cat "${REAPER_FILE}" | grep 'ReaperDefaultImage = ' | cut -d '=' -f 2 | cut -d '"' -f 1
}

main "$@"
