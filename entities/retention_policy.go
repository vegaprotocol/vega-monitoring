package entities

import "fmt"

const InfiniteInterval = "inf"

type RetentionPolicy struct {
	TableName string
	Interval  string
}

func (rp RetentionPolicy) AsString() string {
	return fmt.Sprintf("[Table: %s, Interval: %s]", rp.TableName, rp.Interval)
}

type RetentionPolicies []RetentionPolicy
