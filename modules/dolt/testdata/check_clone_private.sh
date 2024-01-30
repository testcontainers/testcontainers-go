#!/bin/bash

if [ -z "$DOLT_CREDS_PUB_KEY" ]; then
  echo "failed: DOLT_CREDS_PUB_KEY was unset"
  exit 1;
fi

if [ -z "$DOLT_REMOTE_CLONE_URL" ]; then
    echo "failed: DOLT_REMOTE_CLONE_URL was unset"
    exit 1;
fi

if [ -z "$DOLT_REMOTE_CLONE_URL" ]; then
    echo "failed: DOLT_REMOTE_CLONE_URL was unset"
    exit 1;
fi

if [ -z "$(ls -A /root/.dolt/creds)" ]; then
   echo "failed: no creds found"
   exit 1;
fi
