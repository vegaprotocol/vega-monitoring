-- +goose Up

CREATE VIEW metrics.validator_nodes AS (
  SELECT n.*, rs.status AS validator_status, rs.previous_status AS validator_previous_status, rs.stake_score, rs.performance_score, rs.ranking_score, rs.voting_power
  FROM nodes n
  JOIN ranking_scores rs ON ( rs.node_id = n.id AND rs.epoch_seq = (SELECT MAX(id) FROM epochs))
  WHERE
      rs.status <> 'VALIDATOR_NODE_STATUS_PENDING'
    OR
      encode(n.vega_pub_key, 'hex') IN (
        SELECT DISTINCT submitter FROM comet_txs
        WHERE
            vega_time >= NOW() - INTERVAL '2 days'
          AND
            command = 'Validator Heartbeat'
      )
);

-- +goose Down

DROP VIEW IF EXISTS metrics.validator_nodes;
