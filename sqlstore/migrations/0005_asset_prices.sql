-- +goose Up

CREATE TABLE metrics.asset_prices
(
  asset_id            BYTEA                       NOT NULL,
  price_time          TIMESTAMP WITH TIME ZONE    NOT NULL,
  price               NUMERIC(16, 8)              NOT NULL,
  PRIMARY KEY(asset_id)
);

-- +goose Down

DROP TABLE IF EXISTS metrics.asset_prices;
