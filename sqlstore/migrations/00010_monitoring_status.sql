-- +goose Up

ALTER TYPE metrics.monitoring_service_type ADD VALUE IF NOT EXISTS 'DATA_NODE' AFTER 'PROMETHEUS_METAMONITORING';

-- +goose Down

-- Nothing to do because We only added fields