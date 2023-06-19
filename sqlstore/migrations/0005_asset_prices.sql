-- +goose Up

CREATE TABLE metrics.asset_prices
(
  price_time          TIMESTAMP WITH TIME ZONE    NOT NULL,
  asset_id            BYTEA                       NOT NULL,
  price               NUMERIC(16, 8)              NOT NULL,
  PRIMARY KEY(price_time, asset_id)
);
SELECT create_hypertable('metrics.asset_prices', 'price_time', chunk_time_interval => INTERVAL '1 day');

CREATE VIEW asset_prices_current AS (
  SELECT DISTINCT ON (asset_id) * FROM asset_prices ORDER BY asset_id, price_time DESC
);

-- +goose Down

DROP VIEW IF EXISTS asset_prices_current;

DROP TABLE IF EXISTS metrics.asset_prices;
