#!/bin/bash

# use credentials for remote
if [ -n "$DOLT_CREDS_PUB_KEY" ]; then
  dolt creds use "$DOLT_CREDS_PUB_KEY"
fi

# clone
dolt sql -q "CALL DOLT_CLONE('$DOLT_REMOTE_CLONE_URL', '$DOLT_DATABASE');"
