-- +goose Up

CREATE extension IF NOT EXISTS timescaledb;

CREATE SCHEMA IF NOT EXISTS metrics;

CREATE TYPE metrics.signer_role_type AS enum('ROLE_PROPOSER', 'ROLE_SIGNER');

CREATE TABLE metrics.block_signers
(
  vega_time           TIMESTAMP WITH TIME ZONE  NOT NULL,
  role                metrics.signer_role_type  NOT NULL,
  tendermint_pub_key  BYTEA                     NOT NULL,
  PRIMARY KEY(vega_time, role, tendermint_pub_key)
);
SELECT create_hypertable('metrics.block_signers', 'vega_time', chunk_time_interval => INTERVAL '1 day');
CREATE INDEX ON metrics.block_signers (tendermint_pub_key, role);

-- +goose Down

DROP TABLE IF EXISTS metrics.block_signers;
DROP TYPE IF EXISTS metrics.signer_role_type;
DROP SCHEMA IF EXISTS metrics;
