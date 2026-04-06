package model

import (
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func prepareSubscriptionConsumeTables(t *testing.T) {
	t.Helper()
	require.NoError(t, DB.AutoMigrate(&UserSubscription{}, &SubscriptionPreConsumeRecord{}, &SubscriptionCode{}))
	require.NoError(t, DB.Exec(`
		CREATE TABLE IF NOT EXISTS subscription_plans (
			id INTEGER PRIMARY KEY,
			title TEXT NOT NULL DEFAULT '',
			subtitle TEXT NOT NULL DEFAULT '',
			price_amount REAL NOT NULL DEFAULT 0,
			currency TEXT NOT NULL DEFAULT 'USD',
			duration_unit TEXT NOT NULL DEFAULT 'month',
			duration_value INTEGER NOT NULL DEFAULT 1,
			custom_seconds BIGINT NOT NULL DEFAULT 0,
			enabled NUMERIC DEFAULT 1,
			sort_order INTEGER DEFAULT 0,
			stripe_price_id TEXT DEFAULT '',
			creem_product_id TEXT DEFAULT '',
			max_purchase_per_user INTEGER DEFAULT 0,
			upgrade_group TEXT DEFAULT '',
			available_group TEXT DEFAULT '',
			total_amount BIGINT NOT NULL DEFAULT 0,
			quota_reset_period TEXT DEFAULT 'never',
			quota_reset_custom_seconds BIGINT DEFAULT 0,
			created_at BIGINT,
			updated_at BIGINT
		)
	`).Error)
	t.Cleanup(func() {
		DB.Exec("DELETE FROM subscription_pre_consume_records")
		DB.Exec("DELETE FROM user_subscriptions")
		DB.Exec("DELETE FROM subscription_codes")
		DB.Exec("DELETE FROM subscription_plans")
	})
}

func TestPreConsumeUserSubscription_AllowsCodeSubscription(t *testing.T) {
	prepareSubscriptionConsumeTables(t)

	sub := &UserSubscription{
		Id:          1001,
		UserId:      4242,
		PlanId:      0,
		AmountTotal: 1000,
		AmountUsed:  0,
		StartTime:   time.Now().Unix(),
		EndTime:     time.Now().Add(24 * time.Hour).Unix(),
		Status:      "active",
		Source:      "code",
	}
	require.NoError(t, DB.Create(sub).Error)

	res, err := PreConsumeUserSubscription("req-code-sub-1", sub.UserId, "gpt-4o-mini", 0, 200)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, sub.Id, res.UserSubscriptionId)
	assert.EqualValues(t, 200, res.PreConsumed)
	assert.EqualValues(t, sub.AmountTotal, res.AmountTotal)
	assert.EqualValues(t, 0, res.AmountUsedBefore)
	assert.EqualValues(t, 200, res.AmountUsedAfter)

	var reloaded UserSubscription
	require.NoError(t, DB.First(&reloaded, sub.Id).Error)
	assert.EqualValues(t, 200, reloaded.AmountUsed)
}

func TestGetSubscriptionPlanInfoByUserSubscriptionId_CodeSubscriptionReturnsNil(t *testing.T) {
	prepareSubscriptionConsumeTables(t)

	sub := &UserSubscription{
		Id:          1002,
		UserId:      4343,
		PlanId:      0,
		AmountTotal: 500,
		AmountUsed:  0,
		StartTime:   time.Now().Unix(),
		EndTime:     time.Now().Add(24 * time.Hour).Unix(),
		Status:      "active",
		Source:      "code",
	}
	require.NoError(t, DB.Create(sub).Error)

	info, err := GetSubscriptionPlanInfoByUserSubscriptionId(sub.Id)
	require.NoError(t, err)
	assert.Nil(t, info)
}

func TestResolveSubscriptionPlanAndGroupTx_UsesSnapshotAvailableGroup(t *testing.T) {
	prepareSubscriptionConsumeTables(t)

	plan := &SubscriptionPlan{
		Id:             2001,
		Title:          "Snapshot Plan",
		AvailableGroup: "legacy-group",
		DurationUnit:   SubscriptionDurationMonth,
		DurationValue:  1,
		TotalAmount:    1000,
		Enabled:        true,
	}
	require.NoError(t, DB.Create(plan).Error)

	sub := &UserSubscription{
		Id:             1004,
		UserId:         5151,
		PlanId:         plan.Id,
		AmountTotal:    plan.TotalAmount,
		AmountUsed:     0,
		StartTime:      time.Now().Unix(),
		EndTime:        time.Now().Add(24 * time.Hour).Unix(),
		Status:         "active",
		Source:         "order",
		AvailableGroup: plan.AvailableGroup,
	}
	require.NoError(t, DB.Create(sub).Error)
	require.Equal(t, "legacy-group", sub.AvailableGroup)

	require.NoError(t, DB.Model(&SubscriptionPlan{}).
		Where("id = ?", plan.Id).
		Update("available_group", "new-group").Error)
	InvalidateSubscriptionPlanCache(plan.Id)

	resolvedPlan, availableGroup, err := resolveSubscriptionPlanAndGroupTx(nil, sub)
	require.NoError(t, err)
	require.NotNil(t, resolvedPlan)
	assert.Equal(t, plan.Id, resolvedPlan.Id)
	assert.Equal(t, "legacy-group", availableGroup)
}

func TestGetAllUserSubscriptions_IncludesDisabledPlanSummary(t *testing.T) {
	prepareSubscriptionConsumeTables(t)

	plan := &SubscriptionPlan{
		Id:            2002,
		Title:         "Hidden Plan",
		Subtitle:      "legacy subscribers keep this title",
		DurationUnit:  SubscriptionDurationMonth,
		DurationValue: 1,
		Enabled:       false,
	}
	require.NoError(t, DB.Create(plan).Error)

	sub := &UserSubscription{
		Id:          1003,
		UserId:      6262,
		PlanId:      plan.Id,
		AmountTotal: 500,
		StartTime:   time.Now().Unix(),
		EndTime:     time.Now().Add(24 * time.Hour).Unix(),
		Status:      "active",
		Source:      "order",
	}
	require.NoError(t, DB.Create(sub).Error)

	summaries, err := GetAllUserSubscriptions(sub.UserId)
	require.NoError(t, err)
	require.Len(t, summaries, 1)
	require.NotNil(t, summaries[0].Plan)
	assert.Equal(t, plan.Id, summaries[0].Plan.Id)
	assert.Equal(t, plan.Title, summaries[0].Plan.Title)
	assert.Equal(t, plan.Subtitle, summaries[0].Plan.Subtitle)
}

func TestCreateUserSubscriptionFromCodeTx_UsesCurrentTimeForSubscriptionWindow(t *testing.T) {
	prepareSubscriptionConsumeTables(t)

	userID := 7373
	createdAt := common.GetTimestamp() - 7*24*3600
	code := &SubscriptionCode{
		Id:             3001,
		UserId:         1,
		Key:            "redeemtimewindowtestcode00000001",
		Status:         common.RedemptionCodeStatusEnabled,
		Name:           "Day Card",
		Quota:          1000,
		DurationUnit:   SubscriptionDurationDay,
		DurationValue:  1,
		CustomSeconds:  0,
		AvailableGroup: "default",
		CreatedTime:    createdAt,
		ExpiredTime:    0,
	}
	require.NoError(t, DB.Create(code).Error)

	createBefore := common.GetTimestamp()
	sub, err := createUserSubscriptionFromCodeTx(DB, userID, code)
	createAfter := common.GetTimestamp()
	require.NoError(t, err)
	require.NotNil(t, sub)

	assert.Greater(t, sub.StartTime, createdAt)
	assert.GreaterOrEqual(t, sub.StartTime, createBefore-1)
	assert.LessOrEqual(t, sub.StartTime, createAfter+1)

	expectedMinEnd := createBefore + 24*3600 - 2
	expectedMaxEnd := createAfter + 24*3600 + 2
	assert.GreaterOrEqual(t, sub.EndTime, expectedMinEnd)
	assert.LessOrEqual(t, sub.EndTime, expectedMaxEnd)
}
