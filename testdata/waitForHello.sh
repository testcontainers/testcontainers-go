#!/bin/bash

file=/scripts/hello.sh
sleep_pid=0

shutdown() {
    echo "Shutting down..."
    kill $sleep_pid

    wait $sleep_pid
    exit 143 # 128 + 15 -- SIGTERM
}

trap 'shutdown' TERM

echo "Waiting for $file..."
until [ -s "$file" ]
do
    sleep 0.1 &
    sleep_pid=$!

    wait $sleep_pid
done

echo "Found $file"
exec $file
