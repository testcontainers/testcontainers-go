#!/bin/bash
set -e

cqlsh -e "CREATE KEYSPACE IF NOT EXISTS init_sh_keyspace WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1};" && \
cqlsh -e "CREATE TABLE IF NOT EXISTS init_sh_keyspace.test_table (id bigint,name text,primary key (id));" && \
cqlsh -e "INSERT INTO init_sh_keyspace.test_table (id, name) VALUES (1, 'NAME');"