-- +goose Up

CREATE TYPE metrics.balance_source_type AS enum('ASSET_POOL', 'PARTIES_TOTAL', 'UNREALISED_WITHDRAWALS_TOTAL');

CREATE TABLE metrics.network_balances
(
  balance_time        TIMESTAMP WITH TIME ZONE    NOT NULL,
  asset_id            BYTEA                       NOT NULL,
  balance_source      metrics.balance_source_type NOT NULL,
  balance             HUGEINT                     NOT NULL,
  PRIMARY KEY(balance_time, asset_id, balance_source)
);
SELECT create_hypertable('metrics.network_balances', 'balance_time', chunk_time_interval => INTERVAL '1 day');

-- +goose Down

DROP TABLE IF EXISTS metrics.network_balances;
