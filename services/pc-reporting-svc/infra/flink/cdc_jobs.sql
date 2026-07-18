-- Flink SQL CDC Job: Kafka -> ClickHouse for pc-reporting-svc (Tier-0)

-- 1. Create the source table connected to Kafka (reading Avro events from Policy Context)
CREATE TABLE kafka_policy_events (
    `event_id` STRING,
    `tenant_id` STRING,
    `policy_id` STRING,
    `policy_number` STRING,
    `lob` STRING,
    `product_code` STRING,
    `branch_code` STRING,
    `status` STRING,
    `effective_date` TIMESTAMP(3),
    `gross_written_premium` DECIMAL(15, 2),
    `net_written_premium` DECIMAL(15, 2),
    `occurred_at` TIMESTAMP(3),
    WATERMARK FOR `occurred_at` AS `occurred_at` - INTERVAL '5' SECOND
) WITH (
    'connector' = 'kafka',
    'topic' = 'platform.policy.lifecycle.v1',
    'properties.bootstrap.servers' = 'kafka:9092',
    'properties.group.id' = 'flink-reporting-projector',
    'scan.startup.mode' = 'earliest-offset',
    'format' = 'avro-confluent',
    'avro-confluent.schema-registry.url' = 'http://schema-registry:8081',
    'value.format.ignore-parse-errors' = 'true', -- Prevents job failure on bad avro messages
    'properties.dead.letter.queue.topic' = 'platform.dlq.v1' -- Custom logic/connector config for DLQ
);

-- 2. Create the sink table connected to ClickHouse (OLAP Read Model)
CREATE TABLE clickhouse_policy_fact (
    `policy_id` STRING,
    `policy_number` STRING,
    `tenant_id` STRING,
    `lob` STRING,
    `product_code` STRING,
    `branch_code` STRING,
    `status` STRING,
    `effective_date` TIMESTAMP(3),
    `gross_written_premium` DECIMAL(15, 2),
    `net_written_premium` DECIMAL(15, 2),
    `last_event_timestamp` TIMESTAMP(3),
    `sign` TINYINT
) WITH (
    'connector' = 'clickhouse',
    'url' = 'clickhouse://clickhouse:8123',
    'database-name' = 'reporting',
    'table-name' = 'rm_policies_fact',
    'username' = 'default',
    'password' = '',
    'sink.batch-size' = '500',
    'sink.flush-interval' = '1000',
    'sink.max-retries' = '3'
);

-- 3. Execute the Streaming Projection (UPSERT logic via ClickHouse CollapsingMergeTree logic)
-- We insert a sign of 1 for the new state. In a full CDC pipeline handling updates/deletes, 
-- Flink would emit retract streams (-1 sign) before emitting the new state (+1 sign).
INSERT INTO clickhouse_policy_fact
SELECT
    policy_id,
    policy_number,
    tenant_id,
    lob,
    product_code,
    branch_code,
    status,
    effective_date,
    gross_written_premium,
    net_written_premium,
    occurred_at AS last_event_timestamp,
    CAST(1 AS TINYINT) AS sign
FROM kafka_policy_events;
