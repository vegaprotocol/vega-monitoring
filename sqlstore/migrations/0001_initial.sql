-- +goose Up

create extension if not exists timescaledb;

create table m_blocks
(
  vega_time     TIMESTAMP WITH TIME ZONE NOT NULL PRIMARY KEY,
  height        BIGINT                   NOT NULL,
  hash          BYTEA                    NOT NULL
);
select create_hypertable('m_blocks', 'vega_time', chunk_time_interval => INTERVAL '1 day');
create index on blocks (height);

-- +goose Down

DROP TABLE IF EXISTS m_blocks;