#!/usr/bin/env bash


containers=$(docker ps --format "{{.ID}}" | wc -l)

if [ "$containers" -eq "0" ]; then
   exit 0
fi

echo "Number of containers are still running:" "$containers"

exit 255;