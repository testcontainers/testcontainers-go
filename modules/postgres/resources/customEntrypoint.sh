#!/usr/bin/env bash
set -Eeo pipefail


pUID=$(id -u postgres)
pGID=$(id -g postgres)

if [ -z "$pUID" ]
then
    echo "Unable to find postgres user id, required in order to chown key material"
    exit 1
fi

if [ -z "$pGID" ]
then
    echo "Unable to find postgres group id, required in order to chown key material"
    exit 1
fi

chown "$pUID":"$pGID" \
    /tmp/testcontainers-go/postgres/ca_cert.pem \
    /tmp/testcontainers-go/postgres/server.cert \
    /tmp/testcontainers-go/postgres/server.key

/usr/local/bin/docker-entrypoint.sh "$@"
