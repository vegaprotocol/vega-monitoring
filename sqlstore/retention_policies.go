package sqlstore

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

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

func validatePolicies(policies entities.RetentionPolicies) error {
	validIntervalPlural := regexp.MustCompilePOSIX(`^\d+ (hours|days|months|years)$`)
	validIntervalSingular := regexp.MustCompilePOSIX(`1 (hour|day|month|year)`)

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

func SetRetentionPolicies(db *sql.DB, policies entities.RetentionPolicies, logger *zap.Logger) error {
	if err := validatePolicies(policies); err != nil {
		return fmt.Errorf("failed to validate retention policies: %w", err)
	}

	for _, policy := range policies {
		logger.Info("Setting retention policy", zap.String("entity", policy.TableName), zap.String("policy", policy.Interval))
		if err := setRetentionPolicy(db, policy.TableName, policy.Interval); err != nil {
			return fmt.Errorf("failed to set retention policy: %w", err)
		}
	}

	return nil
}
