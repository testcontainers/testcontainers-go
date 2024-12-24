#!/usr/bin/env bash

i=0
while [ $i -ne 30 ]
do
    containers=$(docker ps --format "{{.ID}}" | wc -l)
    if [ "$containers" -eq "0" ]; then
      exit 0
    fi
    sleep 2
    i=$((i + 1))
done

docker ps
echo "Number of containers are still running:" "$containers"
exit 255;