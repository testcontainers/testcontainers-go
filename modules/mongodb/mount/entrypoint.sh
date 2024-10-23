#!/bin/bash

set -Eeuo pipefail

# detect mongo user and group
function get_user_group() {
  user_group=$(cut -d: -f1,5 /etc/passwd | grep mongo)
  echo "${user_group}"
}

# detect the entrypoint
function get_entrypoint() {
  entrypoint=$(find /usr/local/bin -name 'docker-entrypoint.*')
  if [[ "${entrypoint}" == *.py ]]; then
    entrypoint="python3 ${entrypoint}"
  else
    entrypoint="exec ${entrypoint}"
  fi
  echo "${entrypoint}"
}

ENTRYPOINT=$(get_entrypoint)
MONGO_USER_GROUP=$(get_user_group)

# Create the keyfile
openssl rand -base64 756 > "${MONGO_KEYFILE}"

# Set the permissions and ownership of the keyfile
chown "${MONGO_USER_GROUP}" "${MONGO_KEYFILE}"
chmod 400 "${MONGO_KEYFILE}"

${ENTRYPOINT} "$@"
