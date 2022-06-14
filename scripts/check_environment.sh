#!/usr/bin/env bash

docker_out="/tmp/dout.txt"
containers=$(docker ps --format "{{.ID}}" | wc -l)

if [ "$containers" -eq "0" ]; then
   exit 0
fi

docker ps
echo "Number of containers are still running:" "$containers"
exit 255;