package sqlstore

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	vega_sqlstore "code.vegaprotocol.io/vega/datanode/sqlstore"
	"code.vegaprotocol.io/vega/logging"
	"github.com/vegaprotocol/vega-monitoring/entities"
	"go.uber.org/zap"
)

var StandardRetentionPolicy = entities.RetentionPolicies{
	entities.RetentionPolicy{
		TableName: "metrics.block_signers",
		Interval:  "4 months",
	},
	entities.RetentionPolicy{
		TableName: "metrics.network_history_segments",
		Interval:  "4 months",
	},
	entities.RetentionPolicy{
		TableName: "metrics.comet_txs",
		Interval:  "4 months",
	},
	entities.RetentionPolicy{
		TableName: "metrics.network_balances",
		Interval:  "4 months",
	},
	entities.RetentionPolicy{
		TableName: "metrics.asset_prices",
		Interval:  "4 months",
	},
	entities.RetentionPolicy{
		TableName: "metrics.monitoring_status",
		Interval:  "7 days",
	},
}

var LiteRetentionPolicy = entities.RetentionPolicies{
	entities.RetentionPolicy{
		TableName: "metrics.block_signers",
		Interval:  "7 days",
	},
	entities.RetentionPolicy{
		TableName: "metrics.network_history_segments",
		Interval:  "7 days",
	},
	entities.RetentionPolicy{
		TableName: "metrics.comet_txs",
		Interval:  "7 days",
	},
	entities.RetentionPolicy{
		TableName: "metrics.network_balances",
		Interval:  "7 days",
	},
	entities.RetentionPolicy{
		TableName: "metrics.asset_prices",
		Interval:  "7 days",
	},
	entities.RetentionPolicy{
		TableName: "metrics.monitoring_status",
		Interval:  "7 days",
	},
}

var ArchivalRetentionPolicy = entities.RetentionPolicies{
	{
		TableName: "metrics.block_signers",
		Interval:  entities.InfiniteInterval,
	},
	{
		TableName: "metrics.network_history_segments",
		Interval:  entities.InfiniteInterval,
	},
	{
		TableName: "metrics.comet_txs",
		Interval:  entities.InfiniteInterval,
	},
	{
		TableName: "metrics.network_balances",
		Interval:  entities.InfiniteInterval,
	},
	{
		TableName: "metrics.asset_prices",
		Interval:  entities.InfiniteInterval,
	},
	{
		TableName: "metrics.monitoring_status",
		Interval:  entities.InfiniteInterval,
	},
}

func RetentionPoliciesFromConfig(basePolicy string, overrides entities.RetentionPolicies) (entities.RetentionPolicies, error) {
	var basePolicyEntries entities.RetentionPolicies

	if err := validatePolicies(overrides); err != nil {
		return nil, fmt.Errorf("failed to validate overrides for retention policies: %w", err)
	}

	switch basePolicy {
	case entities.RetentionPolicyArchival:
		basePolicyEntries = ArchivalRetentionPolicy
	case entities.RetentionPolicyStandard:
		basePolicyEntries = StandardRetentionPolicy
	case entities.RetentionPolicyLite:
		basePolicyEntries = LiteRetentionPolicy
	default:
		return nil, fmt.Errorf(
			"unknown base retention policy: expected one of %s, %s, %s, got %s",
			entities.RetentionPolicyArchival,
			entities.RetentionPolicyStandard,
			entities.RetentionPolicyLite,
			basePolicy,
		)
	}

	for _, policy := range overrides {
		for idx, basePolicyEntry := range basePolicyEntries {
			if basePolicyEntry.TableName == policy.TableName {
				basePolicyEntries[idx].Interval = policy.Interval
				break
			}
		}
	}

	return basePolicyEntries, nil
}

func validatePolicies(policies entities.RetentionPolicies) error {
	validIntervalPlural := regexp.MustCompile(`^\d+ (hours|days|months|years)$`)
	validIntervalSingular := regexp.MustCompile(`1 (hour|day|month|year)`)

	invalidPolicies := []string{}
	for _, policy := range policies {
		if policy.Interval != entities.InfiniteInterval &&
			!validIntervalPlural.MatchString(policy.Interval) &&
			!validIntervalSingular.MatchString(policy.Interval) {
			invalidPolicies = append(invalidPolicies, policy.AsString())
		}
	}

	if len(invalidPolicies) > 0 {
		return fmt.Errorf("invalid policies: %s", strings.Join(invalidPolicies, ", "))
	}

	return nil
}

func setRetentionPolicy(db *sql.DB, entity string, policy string) error {
	if policy == "" {
		return nil
	}
	if _, err := db.Exec(fmt.Sprintf("SELECT remove_retention_policy('%s', true);", entity)); err != nil {
		return fmt.Errorf("failed removing retention policy from %s: %w", entity, err)
	}

	if policy == entities.InfiniteInterval {
		return nil
	}

	if _, err := db.Exec(fmt.Sprintf("SELECT add_retention_policy('%s', INTERVAL '%s');", entity, policy)); err != nil {
		return fmt.Errorf("failed adding retention policy to %s: %w", entity, err)
	}

	return nil
}

func SetRetentionPolicies(connConfig vega_sqlstore.ConnectionConfig, policies entities.RetentionPolicies, logger *logging.Logger) error {
	if err := validatePolicies(policies); err != nil {
		return fmt.Errorf("failed to validate retention policies: %w", err)
	}

	db, err := DBFromConnectionConfig(logger, connConfig)
	if err != nil {
		return fmt.Errorf("failed to create DB object in migrate schema: %w", err)
	}

	defer db.Close()

	for _, policy := range policies {
		logger.Info("Setting retention policy", zap.String("entity", policy.TableName), zap.String("policy", policy.Interval))
		if err := setRetentionPolicy(db, policy.TableName, policy.Interval); err != nil {
			return fmt.Errorf("failed to set retention policy: %w", err)
		}
	}

	return nil
}
