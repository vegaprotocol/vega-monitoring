-- +goose Up

alter table metrics.network_balances
    add column if not exists chain_id VARCHAR NOT NULL default '';


-- +goose StatementBegin
DO
$$
    DECLARE
        primary_chain_id VARCHAR;
    BEGIN
        -- All existing assets come have been
        -- enable on the primary bridge.
        -- So it's safe to update all assets with the chain ID configured in the
        -- network parameters.
        SELECT value::JSONB ->> 'chain_id' as chain_id
        INTO primary_chain_id
        FROM network_parameters_current
        WHERE key = 'blockchains.ethereumConfig';

        UPDATE metrics.network_balances SET chain_id = primary_chain_id;
    END;
$$;
-- +goose StatementEnd


-- +goose Down

alter table metrics.network_balances
    drop column if exists chain_id;
