#!/usr/bin/env bash

sleep_pid=0

shutdown() {
    echo "Shutting down..."
    kill $sleep_pid

    wait $sleep_pid
    exit 143 # 128 + 15 -- SIGTERM
}

trap 'shutdown' TERM

echo "hello world" > /tmp/hello.txt
echo "done"

sleep 10 &
sleep_pid=$!

wait $sleep_pid
