#!/usr/bin/env bash
set -Eeo pipefail


pUID=$(id -u postgres)
pGID=$(id -g postgres)

if [ -z "$pUID" ]
then
      exit 1
fi

if [ -z "$pGID" ]
then
      exit 1
fi

chown "$pUID":"$pGID" /tmp/data/ca_cert.pem
chown "$pUID":"$pGID" /tmp/data/server.cert
chown "$pUID":"$pGID" /tmp/data/server.key

/usr/local/bin/docker-entrypoint.sh "$@"