package sqlstore_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vegaprotocol/vega-monitoring/entities"
	"github.com/vegaprotocol/vega-monitoring/sqlstore"
)

func TestRetentionPoliciesFromConfig(t *testing.T) {
	testScenarios := []struct {
		name           string
		basePolicyName string
		overrides      entities.RetentionPolicies
		result         entities.RetentionPolicies
		errorMsg       string
	}{
		{
			name:           "unknown base policy",
			basePolicyName: "unknown",
			overrides:      nil,
			errorMsg:       "expected one of archival, standard, lite, got unknown",
		},
		{
			name:           "invalid retention policy",
			basePolicyName: entities.RetentionPolicyArchival,
			overrides: entities.RetentionPolicies{
				{
					TableName: "metrics.network_balances",
					Interval:  "7 units",
				},
			},
			errorMsg: "invalid policies: [Table: metrics.network_balances, Interval: 7 units]",
		},
		{
			name:           "standard policy, no overrides",
			basePolicyName: entities.RetentionPolicyStandard,
			overrides:      nil,
			result:         sqlstore.StandardRetentionPolicy,
		},
		{
			name:           "archival policy, override one",
			basePolicyName: entities.RetentionPolicyArchival,
			overrides: entities.RetentionPolicies{
				{
					TableName: "metrics.network_balances",
					Interval:  "7 days",
				},
			},
			result: entities.RetentionPolicies{
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
					Interval:  "7 days",
				},
				{
					TableName: "metrics.asset_prices",
					Interval:  entities.InfiniteInterval,
				},
				{
					TableName: "metrics.monitoring_status",
					Interval:  entities.InfiniteInterval,
				},
			},
		},
		{
			name:           "archival policy, add one table, override one",
			basePolicyName: entities.RetentionPolicyArchival,
			overrides: entities.RetentionPolicies{
				{
					TableName: "metrics.network_balances",
					Interval:  "7 days",
				},
				{
					TableName: "custom_table",
					Interval:  "14 days",
				},
			},
			result: entities.RetentionPolicies{
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
					Interval:  "7 days",
				},
				{
					TableName: "metrics.asset_prices",
					Interval:  entities.InfiniteInterval,
				},
				{
					TableName: "metrics.monitoring_status",
					Interval:  entities.InfiniteInterval,
				},
			},
		},
	}

	for _, scenario := range testScenarios {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()
			result, err := sqlstore.RetentionPoliciesFromConfig(scenario.basePolicyName, scenario.overrides)
			if len(scenario.errorMsg) < 1 {
				assert.NotNil(t, result)
				assert.Equal(t, scenario.result, result)
				assert.Nil(t, err)
				return
			}

			assert.Nil(t, result)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), scenario.errorMsg)
		})
	}
}
