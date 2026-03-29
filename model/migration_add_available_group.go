package model

import (
	"github.com/QuantumNous/new-api/common"
)

// MigrateAddAvailableGroupToSubscriptionPlan adds available_group column to subscription_plans table
func MigrateAddAvailableGroupToSubscriptionPlan() error {
	if common.UsingPostgreSQL {
		// PostgreSQL
		if err := DB.Exec(`ALTER TABLE subscription_plans ADD COLUMN IF NOT EXISTS available_group VARCHAR(64) DEFAULT ''`).Error; err != nil {
			return err
		}
	} else if common.UsingSQLite {
		// SQLite - check if column exists first
		var count int64
		DB.Raw("SELECT COUNT(*) FROM pragma_table_info('subscription_plans') WHERE name='available_group'").Scan(&count)
		if count == 0 {
			if err := DB.Exec(`ALTER TABLE subscription_plans ADD COLUMN available_group VARCHAR(64) DEFAULT ''`).Error; err != nil {
				return err
			}
		}
	} else {
		// MySQL
		if err := DB.Exec(`ALTER TABLE subscription_plans ADD COLUMN IF NOT EXISTS available_group VARCHAR(64) DEFAULT ''`).Error; err != nil {
			return err
		}
	}
	return nil
}

// MigrateAddAvailableGroupToUserSubscription adds available_group column to user_subscriptions table
func MigrateAddAvailableGroupToUserSubscription() error {
	if common.UsingPostgreSQL {
		// PostgreSQL
		if err := DB.Exec(`ALTER TABLE user_subscriptions ADD COLUMN IF NOT EXISTS available_group VARCHAR(64) DEFAULT ''`).Error; err != nil {
			return err
		}
	} else if common.UsingSQLite {
		// SQLite - check if column exists first
		var count int64
		DB.Raw("SELECT COUNT(*) FROM pragma_table_info('user_subscriptions') WHERE name='available_group'").Scan(&count)
		if count == 0 {
			if err := DB.Exec(`ALTER TABLE user_subscriptions ADD COLUMN available_group VARCHAR(64) DEFAULT ''`).Error; err != nil {
				return err
			}
		}
	} else {
		// MySQL
		if err := DB.Exec(`ALTER TABLE user_subscriptions ADD COLUMN IF NOT EXISTS available_group VARCHAR(64) DEFAULT ''`).Error; err != nil {
			return err
		}
	}
	return nil
}
