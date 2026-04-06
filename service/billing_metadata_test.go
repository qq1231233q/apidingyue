package service

import (
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateMjOtherInfo_IncludesBillingInfo(t *testing.T) {
	relayInfo := &relaycommon.RelayInfo{
		RequestURLPath:                        "/mj/submit?foo=bar",
		BillingSource:                         BillingSourceSubscription,
		SubscriptionId:                        42,
		SubscriptionPreConsumed:               3,
		SubscriptionPostDelta:                 2,
		SubscriptionPlanId:                    7,
		SubscriptionPlanTitle:                 "Agent Pro",
		SubscriptionAmountTotal:               100,
		SubscriptionAmountUsedAfterPreConsume: 10,
		UserSetting: dto.UserSetting{
			BillingPreference: "subscription_first",
		},
	}

	other := GenerateMjOtherInfo(relayInfo, types.PriceData{
		ModelPrice: 0.25,
		GroupRatioInfo: types.GroupRatioInfo{
			GroupRatio:        1.5,
			HasSpecialRatio:   true,
			GroupSpecialRatio: 1.2,
		},
	})

	assert.Equal(t, "/mj/submit", other["request_path"])
	assert.Equal(t, BillingSourceSubscription, other["billing_source"])
	assert.Equal(t, "subscription_first", other["billing_preference"])
	assert.Equal(t, 42, other["subscription_id"])
	assert.Equal(t, 7, other["subscription_plan_id"])
	assert.Equal(t, "Agent Pro", other["subscription_plan_title"])
	assert.Equal(t, int64(100), other["subscription_total"])
	assert.Equal(t, int64(12), other["subscription_used"])
	assert.Equal(t, int64(88), other["subscription_remain"])
	assert.Equal(t, int64(5), other["subscription_consumed"])
	assert.Equal(t, 0, other["wallet_quota_deducted"])
}

func TestChargeViolationFeeIfNeeded_LogIncludesBillingInfo(t *testing.T) {
	truncate(t)

	const userID, tokenID, subscriptionID, channelID = 1, 1, 1, 1

	seedUser(t, userID, 10000)
	seedToken(t, tokenID, userID, "sk-test-key", 10000)
	seedSubscription(t, subscriptionID, userID, 50000, 100)
	seedChannel(t, channelID)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest("POST", "/v1/chat/completions", nil)
	ctx.Set("token_name", "test_token")

	relayInfo := &relaycommon.RelayInfo{
		UserId:                                userID,
		TokenId:                               tokenID,
		TokenKey:                              "sk-test-key",
		OriginModelName:                       "grok-2",
		RequestURLPath:                        "/v1/chat/completions",
		BillingSource:                         BillingSourceSubscription,
		SubscriptionId:                        subscriptionID,
		StartTime:                             time.Now(),
		SubscriptionPlanId:                    9,
		SubscriptionPlanTitle:                 "Agent Pro",
		SubscriptionAmountTotal:               50000,
		SubscriptionAmountUsedAfterPreConsume: 100,
		PriceData: types.PriceData{
			GroupRatioInfo: types.GroupRatioInfo{
				GroupRatio: 1,
			},
		},
		UserSetting: dto.UserSetting{
			BillingPreference: "subscription_first",
		},
		ChannelMeta: &relaycommon.ChannelMeta{
			ChannelId: channelID,
		},
	}

	apiErr := WrapAsViolationFeeGrokCSAM(types.NewError(errors.New(CSAMViolationMarker), types.ErrorCodeInvalidRequest))

	require.True(t, ChargeViolationFeeIfNeeded(ctx, relayInfo, apiErr))

	log := getLastLog(t)
	require.NotNil(t, log)
	other, err := common.StrToMap(log.Other)
	require.NoError(t, err)
	require.NotNil(t, other)

	assert.Equal(t, BillingSourceSubscription, other["billing_source"])
	assert.Equal(t, float64(subscriptionID), other["subscription_id"])
	assert.Equal(t, "subscription_first", other["billing_preference"])
	assert.Equal(t, "/v1/chat/completions", other["request_path"])
	assert.Equal(t, float64(0), other["wallet_quota_deducted"])

	assert.Greater(t, getSubscriptionUsed(t, subscriptionID), int64(100))
}
