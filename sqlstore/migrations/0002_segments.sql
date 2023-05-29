-- +goose Up

CREATE TABLE metrics.network_history_segments
(
  vega_time           TIMESTAMP WITH TIME ZONE NOT NULL,
  height              BIGINT                   NOT NULL,
  data_node           TEXT                     NOT NULL,
  segment_id          TEXT                     NOT NULL,
  PRIMARY KEY(vega_time, data_node)
);
SELECT create_hypertable('metrics.network_history_segments', 'vega_time', chunk_time_interval => INTERVAL '1 day');

-- +goose Down

DROP TABLE IF EXISTS metrics.network_history_segments;
