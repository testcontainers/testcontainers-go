# Reusable Postgres Database

## Introduction

This example shows the usage of the postgres module's Snapshot feature to give each test a clean database without having
to recreate the database container on every test or run heavy scripts to clean your database. This makes the individual
tests very modular, since they always run on a brand-new database.

<!--codeinclude-->
[Test with a reusable Postgres container](../../modules/postgres/postgres_test.go) inside_block:snapshotAndReset
<!--/codeinclude-->


