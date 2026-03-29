package model

import (
	"github.com/QuantumNous/new-api/common"
)

// MigrateSubscriptionCodeAddDurationFields adds duration and available_group fields
func MigrateSubscriptionCodeAddDurationFields() error {
	if !DB.Migrator().HasTable("subscription_codes") {
		return nil
	}

	// Add duration_unit column
	if !DB.Migrator().HasColumn(&SubscriptionCode{}, "duration_unit") {
		if err := DB.Migrator().AddColumn(&SubscriptionCode{}, "duration_unit"); err != nil {
			return err
		}
		// Set default value for existing rows
		if err := DB.Exec("UPDATE subscription_codes SET duration_unit = 'month' WHERE duration_unit = '' OR duration_unit IS NULL").Error; err != nil {
			return err
		}
	}

	// Add duration_value column
	if !DB.Migrator().HasColumn(&SubscriptionCode{}, "duration_value") {
		if err := DB.Migrator().AddColumn(&SubscriptionCode{}, "duration_value"); err != nil {
			return err
		}
		// Set default value for existing rows
		if err := DB.Exec("UPDATE subscription_codes SET duration_value = 1 WHERE duration_value = 0 OR duration_value IS NULL").Error; err != nil {
			return err
		}
	}

	// Add custom_seconds column
	if !DB.Migrator().HasColumn(&SubscriptionCode{}, "custom_seconds") {
		if err := DB.Migrator().AddColumn(&SubscriptionCode{}, "custom_seconds"); err != nil {
			return err
		}
	}

	// Add available_group column
	if !DB.Migrator().HasColumn(&SubscriptionCode{}, "available_group") {
		if err := DB.Migrator().AddColumn(&SubscriptionCode{}, "available_group"); err != nil {
			return err
		}
	}

	return nil
}

// MigrateSubscriptionCodeDaysToQuota migrates days column to quota column
func MigrateSubscriptionCodeDaysToQuota() error {
	// Check if table exists
	if !DB.Migrator().HasTable("subscription_codes") {
		return nil
	}

	// Check if days column exists (old schema)
	if DB.Migrator().HasColumn(&SubscriptionCode{}, "days") {
		if common.UsingPostgreSQL {
			// PostgreSQL: Add quota column, copy data, drop days
			if err := DB.Exec(`
				ALTER TABLE subscription_codes ADD COLUMN IF NOT EXISTS quota BIGINT NOT NULL DEFAULT 0;
				UPDATE subscription_codes SET quota = days * 1000000 WHERE quota = 0;
				ALTER TABLE subscription_codes DROP COLUMN IF EXISTS days;
			`).Error; err != nil {
				return err
			}
		} else if common.UsingSQLite {
			// SQLite doesn't support DROP COLUMN easily, use ALTER TABLE ADD
			if !DB.Migrator().HasColumn(&SubscriptionCode{}, "quota") {
				if err := DB.Exec(`ALTER TABLE subscription_codes ADD COLUMN quota BIGINT NOT NULL DEFAULT 0`).Error; err != nil {
					return err
				}
				// Copy data: convert days to quota (assuming 1 day = 1,000,000 quota)
				if err := DB.Exec(`UPDATE subscription_codes SET quota = days * 1000000 WHERE quota = 0`).Error; err != nil {
					return err
				}
			}
		} else {
			// MySQL: Add quota column, copy data, drop days
			if err := DB.Exec(`
				ALTER TABLE subscription_codes ADD COLUMN IF NOT EXISTS quota BIGINT NOT NULL DEFAULT 0;`).Error; err != nil {
				return err
			}
			if err := DB.Exec(`UPDATE subscription_codes SET quota = days * 1000000 WHERE quota = 0`).Error; err != nil {
				return err
			}
			if err := DB.Exec(`ALTER TABLE subscription_codes DROP COLUMN days`).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// MigrateSubscriptionCodeTable creates subscription_codes table
func MigrateSubscriptionCodeTable() error {
	if common.UsingPostgreSQL {
		// PostgreSQL
		if err := DB.Exec(`
			CREATE TABLE IF NOT EXISTS subscription_codes (
				id SERIAL PRIMARY KEY,
				user_id INTEGER NOT NULL DEFAULT 0,
				key CHAR(32) NOT NULL UNIQUE,
				status INTEGER NOT NULL DEFAULT 1,
				name VARCHAR(255) NOT NULL DEFAULT '',
				days INTEGER NOT NULL DEFAULT 0,
				created_time BIGINT NOT NULL DEFAULT 0,
				redeemed_time BIGINT NOT NULL DEFAULT 0,
				used_user_id INTEGER NOT NULL DEFAULT 0,
				deleted_at TIMESTAMP,
				expired_time BIGINT NOT NULL DEFAULT 0
			);
			CREATE INDEX IF NOT EXISTS idx_subscription_codes_name ON subscription_codes(name);
			CREATE INDEX IF NOT EXISTS idx_subscription_codes_deleted_at ON subscription_codes(deleted_at);
		`).Error; err != nil {
			return err
		}
	} else if common.UsingSQLite {
		// SQLite
		if err := DB.Exec(`
			CREATE TABLE IF NOT EXISTS subscription_codes (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL DEFAULT 0,
				key CHAR(32) NOT NULL UNIQUE,
				status INTEGER NOT NULL DEFAULT 1,
				name VARCHAR(255) NOT NULL DEFAULT '',
				days INTEGER NOT NULL DEFAULT 0,
				created_time BIGINT NOT NULL DEFAULT 0,
				redeemed_time BIGINT NOT NULL DEFAULT 0,
				used_user_id INTEGER NOT NULL DEFAULT 0,
				deleted_at DATETIME,
				expired_time BIGINT NOT NULL DEFAULT 0
			);
			CREATE INDEX IF NOT EXISTS idx_subscription_codes_name ON subscription_codes(name);
			CREATE INDEX IF NOT EXISTS idx_subscription_codes_deleted_at ON subscription_codes(deleted_at);
		`).Error; err != nil {
			return err
		}
	} else {
		// MySQL
		if err := DB.Exec(`
			CREATE TABLE IF NOT EXISTS subscription_codes (
				id INT AUTO_INCREMENT PRIMARY KEY,
				user_id INT NOT NULL DEFAULT 0,
				` + "`key`" + ` CHAR(32) NOT NULL UNIQUE,
				status INT NOT NULL DEFAULT 1,
				name VARCHAR(255) NOT NULL DEFAULT '',
				days INT NOT NULL DEFAULT 0,
				created_time BIGINT NOT NULL DEFAULT 0,
				redeemed_time BIGINT NOT NULL DEFAULT 0,
				used_user_id INT NOT NULL DEFAULT 0,
				deleted_at DATETIME,
				expired_time BIGINT NOT NULL DEFAULT 0,
				INDEX idx_subscription_codes_name (name),
				INDEX idx_subscription_codes_deleted_at (deleted_at)
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
		`).Error; err != nil {
			return err
		}
	}
	return nil
}