-- +goose Up

CREATE TABLE metrics.comet_txs
(
  vega_time           TIMESTAMP WITH TIME ZONE NOT NULL,
  height_idx          SMALLINT                 NOT NULL,
  height              BIGINT                   NOT NULL,
  code                INT                      NOT NULL,
  submitter           TEXT                     NOT NULL,
  command             TEXT                     NOT NULL,
  attributes          JSONB,
  info                TEXT,
  PRIMARY KEY(vega_time, height_idx)
);
CREATE INDEX ON metrics.comet_txs (vega_time, command, submitter);
CREATE INDEX ON metrics.comet_txs (code);
SELECT create_hypertable('metrics.comet_txs', 'vega_time', chunk_time_interval => INTERVAL '1 day');

-- +goose Down

DROP TABLE IF EXISTS metrics.comet_txs;
