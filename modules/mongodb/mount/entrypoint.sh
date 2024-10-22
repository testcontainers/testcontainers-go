#!/bin/bash

set -Eeuo pipefail

openssl rand -base64 756 > "${MONGO_KEYFILE}"
chown "${MONGO_USER_GROUP}" "${MONGO_KEYFILE}"
chmod 400 "${MONGO_KEYFILE}"

exec /usr/local/bin/docker-entrypoint.sh "$@"
