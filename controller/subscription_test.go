package controller

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type subscriptionSelfResponse struct {
	BillingPreference string                      `json:"billing_preference"`
	Subscriptions     []model.SubscriptionSummary `json:"subscriptions"`
	AllSubscriptions  []model.SubscriptionSummary `json:"all_subscriptions"`
}

func setupSubscriptionControllerTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	gin.SetMode(gin.TestMode)
	common.UsingSQLite = true
	common.UsingMySQL = false
	common.UsingPostgreSQL = false
	common.RedisEnabled = false
	common.BatchUpdateEnabled = false
	model.InitDBColumnNames()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	model.DB = db
	model.LOG_DB = db

	if err := db.AutoMigrate(
		&model.User{},
		&model.Log{},
		&model.SubscriptionPlan{},
		&model.UserSubscription{},
		&model.SubscriptionCode{},
	); err != nil {
		t.Fatalf("failed to migrate subscription controller tables: %v", err)
	}

	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	return db
}

func seedSubscriptionTestUser(t *testing.T, db *gorm.DB, userID int, setting dto.UserSetting) *model.User {
	t.Helper()

	user := &model.User{
		Id:       userID,
		Username: fmt.Sprintf("user_%d", userID),
		Password: "password123",
		Status:   common.UserStatusEnabled,
		Role:     common.RoleCommonUser,
		Group:    "default",
	}
	user.SetSetting(setting)
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	return user
}

func seedSubscriptionPlan(t *testing.T, db *gorm.DB, id int, title string, enabled bool) *model.SubscriptionPlan {
	t.Helper()

	plan := &model.SubscriptionPlan{
		Id:            id,
		Title:         title,
		Subtitle:      title + " subtitle",
		PriceAmount:   9.9,
		Currency:      "USD",
		DurationUnit:  model.SubscriptionDurationMonth,
		DurationValue: 1,
		Enabled:       enabled,
		SortOrder:     id,
	}
	if err := db.Create(plan).Error; err != nil {
		t.Fatalf("failed to create subscription plan: %v", err)
	}
	return plan
}

func seedUserSubscription(t *testing.T, db *gorm.DB, sub *model.UserSubscription) *model.UserSubscription {
	t.Helper()

	if err := db.Create(sub).Error; err != nil {
		t.Fatalf("failed to create user subscription: %v", err)
	}
	return sub
}

func seedSubscriptionCode(t *testing.T, db *gorm.DB, code *model.SubscriptionCode) *model.SubscriptionCode {
	t.Helper()

	if err := db.Create(code).Error; err != nil {
		t.Fatalf("failed to create subscription code: %v", err)
	}
	return code
}

func TestGetSubscriptionSelf_IncludesDisabledPlanSummaryAndActiveSplit(t *testing.T) {
	db := setupSubscriptionControllerTestDB(t)
	user := seedSubscriptionTestUser(t, db, 101, dto.UserSetting{
		BillingPreference: "subscription_only",
	})
	activePlan := seedSubscriptionPlan(t, db, 201, "Pro Plan", true)
	disabledPlan := seedSubscriptionPlan(t, db, 202, "Legacy Plan", false)

	now := common.GetTimestamp()
	seedUserSubscription(t, db, &model.UserSubscription{
		Id:             301,
		UserId:         user.Id,
		PlanId:         activePlan.Id,
		AmountTotal:    5000,
		AmountUsed:     100,
		StartTime:      now - 3600,
		EndTime:        now + 86400,
		Status:         "active",
		Source:         "purchase",
		AvailableGroup: "vip",
		CreatedAt:      now - 3600,
		UpdatedAt:      now - 3600,
	})
	seedUserSubscription(t, db, &model.UserSubscription{
		Id:             302,
		UserId:         user.Id,
		PlanId:         disabledPlan.Id,
		AmountTotal:    1000,
		AmountUsed:     1000,
		StartTime:      now - 86400*3,
		EndTime:        now - 60,
		Status:         "expired",
		Source:         "purchase",
		AvailableGroup: "legacy",
		CreatedAt:      now - 86400*3,
		UpdatedAt:      now - 60,
	})

	ctx, recorder := newAuthenticatedContext(t, http.MethodGet, "/api/subscription/self", nil, user.Id)
	GetSubscriptionSelf(ctx)

	response := decodeAPIResponse(t, recorder)
	if !response.Success {
		t.Fatalf("expected success response, got message: %s", response.Message)
	}

	var payload subscriptionSelfResponse
	if err := common.Unmarshal(response.Data, &payload); err != nil {
		t.Fatalf("failed to decode subscription self response: %v", err)
	}

	if payload.BillingPreference != "subscription_only" {
		t.Fatalf("expected billing preference subscription_only, got %q", payload.BillingPreference)
	}
	if len(payload.Subscriptions) != 1 {
		t.Fatalf("expected 1 active subscription, got %d", len(payload.Subscriptions))
	}
	if len(payload.AllSubscriptions) != 2 {
		t.Fatalf("expected 2 total subscriptions, got %d", len(payload.AllSubscriptions))
	}
	if payload.Subscriptions[0].Plan == nil || payload.Subscriptions[0].Plan.Title != activePlan.Title {
		t.Fatalf("expected active subscription to include plan %q", activePlan.Title)
	}

	foundDisabledPlan := false
	for _, item := range payload.AllSubscriptions {
		if item.Subscription != nil && item.Subscription.PlanId == disabledPlan.Id {
			if item.Plan == nil {
				t.Fatalf("expected disabled plan subscription summary to include plan metadata")
			}
			if item.Plan.Title != disabledPlan.Title {
				t.Fatalf("expected disabled plan title %q, got %q", disabledPlan.Title, item.Plan.Title)
			}
			foundDisabledPlan = true
		}
	}
	if !foundDisabledPlan {
		t.Fatalf("expected all_subscriptions to include disabled plan subscription")
	}
}

func TestRedeemSubscriptionCode_AcceptsJSONAndCreatesSubscription(t *testing.T) {
	db := setupSubscriptionControllerTestDB(t)
	user := seedSubscriptionTestUser(t, db, 102, dto.UserSetting{})
	now := common.GetTimestamp()
	code := seedSubscriptionCode(t, db, &model.SubscriptionCode{
		Id:             401,
		UserId:         1,
		Key:            "1234567890abcdef1234567890abcdef",
		Status:         common.RedemptionCodeStatusEnabled,
		Name:           "Day Pass",
		Quota:          2048,
		DurationUnit:   model.SubscriptionDurationDay,
		DurationValue:  1,
		AvailableGroup: "vip",
		CreatedTime:    now - 30,
	})

	before := time.Now().Unix()
	ctx, recorder := newAuthenticatedContext(t, http.MethodPost, "/api/subscription_code/redeem", map[string]any{
		"code": code.Key,
	}, user.Id)
	RedeemSubscriptionCode(ctx)
	after := time.Now().Unix()

	response := decodeAPIResponse(t, recorder)
	if !response.Success {
		t.Fatalf("expected redeem response to succeed, got message: %s", response.Message)
	}

	var quota int64
	if err := common.Unmarshal(response.Data, &quota); err != nil {
		t.Fatalf("failed to decode redeem quota: %v", err)
	}
	if quota != code.Quota {
		t.Fatalf("expected redeemed quota %d, got %d", code.Quota, quota)
	}

	var storedCode model.SubscriptionCode
	if err := db.First(&storedCode, code.Id).Error; err != nil {
		t.Fatalf("failed to reload subscription code: %v", err)
	}
	if storedCode.Status != common.RedemptionCodeStatusUsed {
		t.Fatalf("expected code status %d, got %d", common.RedemptionCodeStatusUsed, storedCode.Status)
	}
	if storedCode.UsedUserId != user.Id {
		t.Fatalf("expected code used_user_id %d, got %d", user.Id, storedCode.UsedUserId)
	}

	var sub model.UserSubscription
	if err := db.Where("user_id = ? AND source = ?", user.Id, "code").First(&sub).Error; err != nil {
		t.Fatalf("expected code redemption to create subscription: %v", err)
	}
	if sub.PlanId != 0 {
		t.Fatalf("expected code subscription plan_id 0, got %d", sub.PlanId)
	}
	if sub.AmountTotal != code.Quota {
		t.Fatalf("expected subscription total %d, got %d", code.Quota, sub.AmountTotal)
	}
	if sub.AvailableGroup != code.AvailableGroup {
		t.Fatalf("expected available group %q, got %q", code.AvailableGroup, sub.AvailableGroup)
	}
	if sub.StartTime < before || sub.StartTime > after+1 {
		t.Fatalf("expected start_time to be redeem-time based, got %d (window %d-%d)", sub.StartTime, before, after+1)
	}
	if sub.EndTime <= sub.StartTime {
		t.Fatalf("expected end_time %d to be after start_time %d", sub.EndTime, sub.StartTime)
	}
}

func TestUpdateSubscriptionPreference_NormalizesInvalidValue(t *testing.T) {
	db := setupSubscriptionControllerTestDB(t)
	user := seedSubscriptionTestUser(t, db, 103, dto.UserSetting{
		BillingPreference: "wallet_only",
	})

	ctx, recorder := newAuthenticatedContext(t, http.MethodPut, "/api/subscription/self/preference", map[string]any{
		"billing_preference": "not-a-real-preference",
	}, user.Id)
	UpdateSubscriptionPreference(ctx)

	response := decodeAPIResponse(t, recorder)
	if !response.Success {
		t.Fatalf("expected update preference response to succeed, got message: %s", response.Message)
	}

	var payload map[string]string
	if err := common.Unmarshal(response.Data, &payload); err != nil {
		t.Fatalf("failed to decode update preference response: %v", err)
	}
	if payload["billing_preference"] != "subscription_first" {
		t.Fatalf("expected normalized preference subscription_first, got %q", payload["billing_preference"])
	}

	var stored model.User
	if err := db.First(&stored, user.Id).Error; err != nil {
		t.Fatalf("failed to reload user: %v", err)
	}
	if stored.GetSetting().BillingPreference != "subscription_first" {
		t.Fatalf("expected stored normalized preference subscription_first, got %q", stored.GetSetting().BillingPreference)
	}
}

func TestUpdateSubscriptionCode_RejectsInvalidAvailableGroup(t *testing.T) {
	db := setupSubscriptionControllerTestDB(t)
	code := seedSubscriptionCode(t, db, &model.SubscriptionCode{
		Id:             501,
		UserId:         1,
		Key:            "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Status:         common.RedemptionCodeStatusEnabled,
		Name:           "Starter Code",
		Quota:          1024,
		DurationUnit:   model.SubscriptionDurationDay,
		DurationValue:  1,
		AvailableGroup: "vip",
		CreatedTime:    common.GetTimestamp(),
	})

	ctx, recorder := newAuthenticatedContext(t, http.MethodPut, "/api/subscription_code/", map[string]any{
		"id":              code.Id,
		"name":            "Starter Code",
		"quota":           code.Quota,
		"duration_unit":   model.SubscriptionDurationDay,
		"duration_value":  1,
		"custom_seconds":  0,
		"available_group": "group-not-exists",
		"expired_time":    0,
	}, 1)
	UpdateSubscriptionCode(ctx)

	response := decodeAPIResponse(t, recorder)
	if response.Success {
		t.Fatalf("expected invalid available_group update to fail")
	}

	var stored model.SubscriptionCode
	if err := db.First(&stored, code.Id).Error; err != nil {
		t.Fatalf("failed to reload subscription code: %v", err)
	}
	if stored.AvailableGroup != code.AvailableGroup {
		t.Fatalf("expected available_group to remain %q, got %q", code.AvailableGroup, stored.AvailableGroup)
	}
}

func TestUpdateSubscriptionCode_RejectsInvalidCustomDuration(t *testing.T) {
	db := setupSubscriptionControllerTestDB(t)
	code := seedSubscriptionCode(t, db, &model.SubscriptionCode{
		Id:             502,
		UserId:         1,
		Key:            "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		Status:         common.RedemptionCodeStatusEnabled,
		Name:           "Custom Code",
		Quota:          2048,
		DurationUnit:   model.SubscriptionDurationHour,
		DurationValue:  3,
		AvailableGroup: "vip",
		CreatedTime:    common.GetTimestamp(),
	})

	ctx, recorder := newAuthenticatedContext(t, http.MethodPut, "/api/subscription_code/", map[string]any{
		"id":              code.Id,
		"name":            "Custom Code",
		"quota":           code.Quota,
		"duration_unit":   model.SubscriptionDurationCustom,
		"duration_value":  0,
		"custom_seconds":  0,
		"available_group": "vip",
		"expired_time":    0,
	}, 1)
	UpdateSubscriptionCode(ctx)

	response := decodeAPIResponse(t, recorder)
	if response.Success {
		t.Fatalf("expected invalid custom duration update to fail")
	}

	var stored model.SubscriptionCode
	if err := db.First(&stored, code.Id).Error; err != nil {
		t.Fatalf("failed to reload subscription code: %v", err)
	}
	if stored.DurationUnit != code.DurationUnit {
		t.Fatalf("expected duration_unit to remain %q, got %q", code.DurationUnit, stored.DurationUnit)
	}
	if stored.CustomSeconds != code.CustomSeconds {
		t.Fatalf("expected custom_seconds to remain %d, got %d", code.CustomSeconds, stored.CustomSeconds)
	}
}
