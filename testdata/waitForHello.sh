#!/bin/bash

file=/scripts/hello.sh

until [ -s "$file" ]
do
    sleep 0.1
done

sh $file