#!/usr/bin/env bash
set -Eeo pipefail

# Hardcoded user id from pg11-alpine image
chown 70:70 /tmp/data/ca_cert.pem
chown 70:70 /tmp/data/server.cert
chown 70:70 /tmp/data/server.key

ls -lah /tmp/data
/usr/local/bin/docker-entrypoint.sh "$@"