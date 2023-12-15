<?xml version="1.0"?>
<clickhouse>
    <zookeeper>
        <node index="1">
            <host>{{.Host}}</host>
            <port>{{.Port}}</port>
        </node>
    </zookeeper>

    <remote_servers>
        <default>
            <shard>
                <replica>
                    <host>localhost</host>
                    <port>9000</port>
                </replica>
            </shard>
        </default>
    </remote_servers>
    <macros>
        <cluster>default</cluster>
        <shard>shard</shard>
        <replica>replica</replica>
    </macros>

    <distributed_ddl>
        <path>/clickhouse/task_queue/ddl</path>
    </distributed_ddl>

    <format_schema_path>/var/lib/clickhouse/format_schemas/</format_schema_path>
</clickhouse>
