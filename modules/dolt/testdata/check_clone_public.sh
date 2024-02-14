#!/bin/bash

if [ -z "$DOLT_REMOTE_CLONE_URL" ]; then
    echo "failed: DOLT_REMOTE_CLONE_URL was unset"
    exit 1;
fi
