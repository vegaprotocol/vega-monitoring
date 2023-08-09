-- +goose Up

CREATE VIEW metrics.metamonitoring_status AS (
  (
    SELECT
      'asset_prices' AS check_name,
      CASE WHEN COUNT(*) > 0 THEN 1 ELSE 0 END AS is_healthy,
      last(price_time, price_time) AS last_update
    FROM metrics.asset_prices
    WHERE
      price_time > NOW() - INTERVAL '20 min'
  ) UNION ALL (
    SELECT
      'block_signers' AS check_name,
      CASE WHEN COUNT(*) > 0 THEN 1 ELSE 0 END AS is_healthy,
      last(vega_time, vega_time) AS last_update
    FROM metrics.block_signers
    WHERE
      vega_time > NOW() - INTERVAL '3 min'
  ) UNION ALL (
    SELECT
      'comet_txs' AS check_name,
      CASE WHEN COUNT(*) > 0 THEN 1 ELSE 0 END AS is_healthy,
      last(vega_time, vega_time) AS last_update
    FROM metrics.comet_txs
    WHERE
      vega_time > NOW() - INTERVAL '3 min'
  ) UNION ALL (
    SELECT
      'network_balances' AS check_name,
      CASE WHEN COUNT(*) > 0 THEN 1 ELSE 0 END AS is_healthy,
      last(balance_time, balance_time) AS last_update
    FROM metrics.network_balances
    WHERE
      balance_time > NOW() - INTERVAL '3 min'
  ) UNION ALL (
    SELECT
      'network_history_segments' AS check_name,
      CASE WHEN COUNT(*) > 0 THEN 1 ELSE 0 END AS is_healthy,
      last(vega_time, vega_time) AS last_update
    FROM metrics.network_history_segments
    WHERE
      vega_time > NOW() - INTERVAL '20 min'
  ) UNION ALL (
    SELECT
      'data_node' AS check_name,
      CASE WHEN COUNT(*) > 0 THEN 1 ELSE 0 END AS is_healthy,
      last(vega_time, vega_time) AS last_update
    FROM last_block
    WHERE
      vega_time > NOW() - INTERVAL '3 min'
  )
  ORDER BY check_name
);

-- +goose Down

DROP VIEW IF EXISTS metrics.metamonitoring_status;
