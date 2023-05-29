-- +goose Up

CREATE extension IF NOT EXISTS timescaledb;

CREATE SCHEMA metrics;

CREATE TYPE metrics.signer_role_type AS enum('ROLE_PROPOSER', 'ROLE_SIGNER');

create table metrics.block_signers
(
  vega_time           TIMESTAMP WITH TIME ZONE  NOT NULL,
  role                metrics.signer_role_type  NOT NULL,
  tendermint_pub_key  BYTEA                     NOT NULL,
  PRIMARY KEY(vega_time, role, tendermint_pub_key)
);
select create_hypertable('metrics.block_signers', 'vega_time', chunk_time_interval => INTERVAL '1 day');
create index on metrics.block_signers (tendermint_pub_key, role);

-- +goose Down

DROP TABLE IF EXISTS metrics.block_signers;
DROP TYPE IF EXISTS metrics.signer_role_type;
SET search_path TO public;
DROP SCHEMA IF EXISTS metrics;
