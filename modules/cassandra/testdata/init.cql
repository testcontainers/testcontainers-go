CREATE KEYSPACE IF NOT EXISTS test_keyspace WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };
CREATE TABLE IF NOT EXISTS test_keyspace.test_table (id bigint,name text,primary key (id));
INSERT INTO test_keyspace.test_table (id, name) VALUES (1, 'NAME');