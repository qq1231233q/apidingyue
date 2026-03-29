package model

import (
	"github.com/QuantumNous/new-api/common"
)

// MigrateSubscriptionCodeToSubscription migrates subscription_codes table to add subscription fields
func MigrateSubscriptionCodeToSubscription() error {
	if !common.UsingMySQL && !common.UsingSQLite && !common.UsingPostgreSQL {
		return nil
	}

	// Add new columns for subscription configuration
	columns := []struct {
		name    string
		sqlType string
	}{
		{"duration_unit", "VARCHAR(16) DEFAULT 'month'"},
		{"duration_value", "INT DEFAULT 1"},
		{"custom_seconds", "BIGINT DEFAULT 0"},
		{"available_group", "VARCHAR(64) DEFAULT ''"},
	}

	for _, col := range columns {
		var sql string
		if common.UsingPostgreSQL {
			sql = `ALTER TABLE subscription_codes ADD COLUMN IF NOT EXISTS ` + col.name + ` ` + col.sqlType
		} else if common.UsingMySQL {
			sql = `ALTER TABLE subscription_codes ADD COLUMN ` + col.name + ` ` + col.sqlType
		} else {
			// SQLite
			sql = `ALTER TABLE subscription_codes ADD COLUMN ` + col.name + ` ` + col.sqlType
		}

		err := DB.Exec(sql).Error
		if err != nil {
			// Column might already exist, log but don't fail
			common.SysLog("migration warning: " + err.Error())
		}
	}

	common.SysLog("subscription code migration completed")
	return nil
}