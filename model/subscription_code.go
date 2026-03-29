package model

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"

	"gorm.io/gorm"
)

// escapeWildcards escapes SQL LIKE wildcard characters to prevent injection
func escapeWildcards(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "%", "\\%")
	s = strings.ReplaceAll(s, "_", "\\_")
	return s
}

// SubscriptionCode represents an activation code that creates a subscription
type SubscriptionCode struct {
	Id              int            `json:"id"`
	UserId          int            `json:"user_id"`
	Key             string         `json:"key" gorm:"type:char(32);uniqueIndex"`
	Status          int            `json:"status" gorm:"default:1"`
	Name            string         `json:"name" gorm:"index"`
	
	// Subscription configuration
	Quota           int64  `json:"quota" gorm:"type:bigint;default:0"`           // Total quota amount
	DurationUnit    string `json:"duration_unit" gorm:"type:varchar(16);default:'month'"` // year/month/day/hour/custom
	DurationValue   int    `json:"duration_value" gorm:"type:int;default:1"`     // Duration value
	CustomSeconds   int64  `json:"custom_seconds" gorm:"type:bigint;default:0"`  // Custom duration in seconds
	AvailableGroup  string `json:"available_group" gorm:"type:varchar(64);default:''"` // Restrict to specific group
	
	CreatedTime     int64          `json:"created_time" gorm:"bigint"`
	RedeemedTime    int64          `json:"redeemed_time" gorm:"bigint"`
	Count           int            `json:"count" gorm:"-:all"`
	UsedUserId      int            `json:"used_user_id"`
	DeletedAt       gorm.DeletedAt `gorm:"index"`
	ExpiredTime     int64          `json:"expired_time" gorm:"bigint"` // Code expiry (not subscription expiry)
}

func GetAllSubscriptionCodes(startIdx int, num int) (codes []*SubscriptionCode, total int64, err error) {
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	err = tx.Model(&SubscriptionCode{}).Count(&total).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	err = tx.Order("id desc").Limit(num).Offset(startIdx).Find(&codes).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	if err = tx.Commit().Error; err != nil {
		return nil, 0, err
	}

	return codes, total, nil
}

func SearchSubscriptionCodes(keyword string, startIdx int, num int) (codes []*SubscriptionCode, total int64, err error) {
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	query := tx.Model(&SubscriptionCode{})

	// Sanitize keyword to prevent SQL injection in LIKE clause
	// Escape special characters: %, _, \
	sanitizedKeyword := keyword
	sanitizedKeyword = escapeWildcards(sanitizedKeyword)

	if id, err := strconv.Atoi(keyword); err == nil {
		query = query.Where("id = ? OR name LIKE ?", id, sanitizedKeyword+"%")
	} else {
		query = query.Where("name LIKE ?", sanitizedKeyword+"%")
	}

	err = query.Count(&total).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	err = query.Order("id desc").Limit(num).Offset(startIdx).Find(&codes).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	if err = tx.Commit().Error; err != nil {
		return nil, 0, err
	}

	return codes, total, nil
}

func GetSubscriptionCodeById(id int) (*SubscriptionCode, error) {
	if id == 0 {
		return nil, errors.New("id 为空！")
	}
	code := SubscriptionCode{Id: id}
	err := DB.First(&code, "id = ?", id).Error
	return &code, err
}

func RedeemSubscriptionCode(key string, userId int) (quota int64, err error) {
	if key == "" {
		return 0, errors.New("未提供激活码")
	}
	if userId == 0 {
		return 0, errors.New("无效的 user id")
	}
	code := &SubscriptionCode{}

	keyCol := "`key`"
	if common.UsingPostgreSQL {
		keyCol = `"key"`
	}
	common.RandomSleep()
	err = DB.Transaction(func(tx *gorm.DB) error {
		err := tx.Set("gorm:query_option", "FOR UPDATE").Where(keyCol+" = ?", key).First(code).Error
		if err != nil {
			return errors.New("无效的激活码")
		}
		if code.Status != common.RedemptionCodeStatusEnabled {
			return errors.New("该激活码已被使用")
		}
		if code.ExpiredTime != 0 && code.ExpiredTime < common.GetTimestamp() {
			return errors.New("该激活码已过期")
		}

		// Create subscription from code configuration
		_, err = createUserSubscriptionFromCodeTx(tx, userId, code)
		if err != nil {
			return err
		}

		code.RedeemedTime = common.GetTimestamp()
		code.Status = common.RedemptionCodeStatusUsed
		code.UsedUserId = userId
		err = tx.Save(code).Error
		return err
	})
	if err != nil {
		common.SysError("subscription code redemption failed: " + err.Error())
		return 0, err
	}
	RecordLog(userId, LogTypeSystem, fmt.Sprintf("通过激活码兑换订阅，激活码ID %d，订阅额度 %d", code.Id, code.Quota))
	return code.Quota, nil
}

// createUserSubscriptionFromCodeTx creates a UserSubscription from a SubscriptionCode
func createUserSubscriptionFromCodeTx(tx *gorm.DB, userId int, code *SubscriptionCode) (*UserSubscription, error) {
	if tx == nil || code == nil {
		return nil, errors.New("invalid arguments")
	}
	if userId <= 0 {
		return nil, errors.New("invalid user id")
	}
	if code.Quota <= 0 {
		return nil, errors.New("激活码额度必须大于 0")
	}
	if code.DurationValue <= 0 && code.DurationUnit != SubscriptionDurationCustom {
		return nil, errors.New("激活码时长必须大于 0")
	}

	nowUnix := GetDBTimestamp()
	now := time.Unix(nowUnix, 0)
	
	// Calculate subscription end time
	var endUnix int64
	switch code.DurationUnit {
	case SubscriptionDurationYear:
		endUnix = now.AddDate(code.DurationValue, 0, 0).Unix()
	case SubscriptionDurationMonth:
		endUnix = now.AddDate(0, code.DurationValue, 0).Unix()
	case SubscriptionDurationDay:
		endUnix = now.AddDate(0, 0, code.DurationValue).Unix()
	case SubscriptionDurationHour:
		endUnix = now.Add(time.Duration(code.DurationValue) * time.Hour).Unix()
	case SubscriptionDurationCustom:
		if code.CustomSeconds <= 0 {
			return nil, errors.New("自定义时长必须大于 0")
		}
		endUnix = now.Add(time.Duration(code.CustomSeconds) * time.Second).Unix()
	default:
		return nil, fmt.Errorf("无效的时长单位: %s", code.DurationUnit)
	}

	sub := &UserSubscription{
		UserId:         userId,
		PlanId:         0, // No plan, created from code
		AmountTotal:    code.Quota,
		AmountUsed:     0,
		StartTime:      nowUnix,
		EndTime:        endUnix,
		Status:         "active",
		Source:         "code",
		LastResetTime:  0,
		NextResetTime:  0,
		UpgradeGroup:   "",
		PrevUserGroup:  "",
		AvailableGroup: code.AvailableGroup,
		CreatedAt:      common.GetTimestamp(),
		UpdatedAt:      common.GetTimestamp(),
	}
	
	if err := tx.Create(sub).Error; err != nil {
		return nil, err
	}
	return sub, nil
}

func (code *SubscriptionCode) Insert() error {
	return DB.Create(code).Error
}

func (code *SubscriptionCode) SelectUpdate() error {
	return DB.Model(code).Select("redeemed_time", "status").Updates(code).Error
}

func (code *SubscriptionCode) Update() error {
	return DB.Model(code).Select("name", "quota", "duration_unit", "duration_value", "custom_seconds", "available_group", "status", "redeemed_time", "expired_time").Updates(code).Error
}

func (code *SubscriptionCode) Delete() error {
	return DB.Delete(code).Error
}

func DeleteSubscriptionCodeById(id int) error {
	if id == 0 {
		return errors.New("id 为空！")
	}
	code := SubscriptionCode{Id: id}
	err := DB.Where(code).First(&code).Error
	if err != nil {
		return err
	}
	return code.Delete()
}

func DeleteInvalidSubscriptionCodes() (int64, error) {
	now := common.GetTimestamp()
	result := DB.Where("status IN ? OR (status = ? AND expired_time != 0 AND expired_time < ?)", []int{common.RedemptionCodeStatusDisabled, common.RedemptionCodeStatusUsed}, common.RedemptionCodeStatusEnabled, now).Delete(&SubscriptionCode{})
	return result.RowsAffected, result.Error
}