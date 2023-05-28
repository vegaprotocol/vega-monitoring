-- +goose Up

create extension if not exists timescaledb;

create type signer_role_type as enum('ROLE_PROPOSER', 'ROLE_SIGNER');

create table m_block_signers
(
  vega_time           TIMESTAMP WITH TIME ZONE NOT NULL,
  height              BIGINT                   NOT NULL,
  role                signer_role_type         NOT NULL,
  tendermint_address  TEXT                     NOT NULL,
  tendermint_pub_key  BYTEA,
  PRIMARY KEY(vega_time, role, tendermint_address)
);
select create_hypertable('m_block_signers', 'vega_time', chunk_time_interval => INTERVAL '1 day');
create index on m_block_signers (tendermint_address, role);

-- +goose Down

DROP TABLE IF EXISTS m_block_signers;
DROP TYPE IF EXISTS signer_role_type;
