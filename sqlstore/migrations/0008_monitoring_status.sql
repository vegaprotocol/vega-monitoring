-- +goose Up



CREATE TYPE IF NOT EXISTS metrics.monitoring_service_type AS ENUM (
  'BLOCK_SIGNERS', 
  'SEGMENTS', 
  'COMET_TXS',
  'NETWORK_BALANCES',
  'ASSET_PRICES',
);

CREATE TABLE IF NOT EXISTS metrics.monitoring_status (
  status_time           TIMESTAMP WITH TIME ZONE NOT NULL,
  is_healthy          BOOLEAN NOT NULL,
  monitoring_service  metrics.monitoring_service_type NOT NULL,
  unhealthy_reason    SMALLINT DEFAULT 0,
);

SELECT create_hypertable('metrics.monitoring_status', 'status_time', chunk_time_interval => INTERVAL '1 day');
CREATE INDEX ON metrics.monitoring_status (monitoring_service);

-- +goose Down

DROP TABLE IF EXISTS metrics.monitoring_status;
DROP TYPE IF EXISTS metrics.monitoring_service_type;