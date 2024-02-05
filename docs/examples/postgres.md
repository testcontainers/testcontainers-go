# Reusable Postgres Database

## Introduction

This example shows how to use the postgres module with templates to give each test a clean database without having
to recreate the database container on every test or run heavy scripts to clean your database. This makes the individual
tests very modular, since they always run on a brand-new database.

<!--codeinclude-->
[Creating the Postgres container](../../examples/postgres/postgres.go) inside_block:runContainer
<!--/codeinclude-->

<!--codeinclude-->
[Snapshot and reset the database](../../examples/postgres/postgres.go) inside_block:snapshotAndReset
<!--/codeinclude-->

<!--codeinclude-->
[Test the reusable Postgres container](../../examples/postgres/postgres_test.go)
<!--/codeinclude-->


