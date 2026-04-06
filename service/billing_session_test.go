package service

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func prepareBillingSessionTables(t *testing.T) {
	t.Helper()
	require.NoError(t, model.DB.AutoMigrate(&model.SubscriptionPreConsumeRecord{}))
}

func makeBillingSessionContext() *gin.Context {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest("POST", "/v1/chat/completions", nil)
	return ctx
}

func makeRelayInfoForBillingTest(userID, tokenID int, tokenKey, pref, requestID string, userQuota int) *relaycommon.RelayInfo {
	return &relaycommon.RelayInfo{
		UserId:          userID,
		TokenId:         tokenID,
		TokenKey:        tokenKey,
		OriginModelName: "gpt-4o-mini",
		RequestId:       requestID,
		UserQuota:       userQuota,
		UserSetting: dto.UserSetting{
			BillingPreference: pref,
		},
		StartTime: time.Now(),
	}
}

func TestNewBillingSession_SubscriptionFirst_NoActiveSubscriptionFallsBackToWallet(t *testing.T) {
	prepareBillingSessionTables(t)
	truncate(t)

	const (
		userID       = 11
		tokenID      = 21
		initialQuota = 5000
		tokenQuota   = 4000
		preConsume   = 300
	)

	seedUser(t, userID, initialQuota)
	seedToken(t, tokenID, userID, "sk-sub-first-wallet", tokenQuota)

	ctx := makeBillingSessionContext()
	relayInfo := makeRelayInfoForBillingTest(
		userID,
		tokenID,
		"sk-sub-first-wallet",
		"subscription_first",
		t.Name(),
		initialQuota,
	)

	session, apiErr := NewBillingSession(ctx, relayInfo, preConsume)
	require.Nil(t, apiErr)
	require.NotNil(t, session)

	assert.Equal(t, BillingSourceWallet, session.funding.Source())
	assert.Equal(t, BillingSourceWallet, relayInfo.BillingSource)
	assert.Equal(t, 0, relayInfo.SubscriptionId)
	assert.Equal(t, preConsume, session.GetPreConsumedQuota())
	assert.Equal(t, initialQuota-preConsume, getUserQuota(t, userID))
	assert.Equal(t, tokenQuota-preConsume, getTokenRemainQuota(t, tokenID))
}

func TestNewBillingSession_WalletFirst_InsufficientWalletFallsBackToSubscription(t *testing.T) {
	prepareBillingSessionTables(t)
	truncate(t)

	const (
		userID       = 12
		tokenID      = 22
		subID        = 32
		initialQuota = 100
		tokenQuota   = 4000
		preConsume   = 300
		subTotal     = 5000
		subUsed      = 50
	)

	seedUser(t, userID, initialQuota)
	seedToken(t, tokenID, userID, "sk-wallet-first-sub", tokenQuota)
	seedSubscription(t, subID, userID, subTotal, subUsed)

	ctx := makeBillingSessionContext()
	relayInfo := makeRelayInfoForBillingTest(
		userID,
		tokenID,
		"sk-wallet-first-sub",
		"wallet_first",
		t.Name(),
		initialQuota,
	)

	session, apiErr := NewBillingSession(ctx, relayInfo, preConsume)
	require.Nil(t, apiErr)
	require.NotNil(t, session)

	assert.Equal(t, BillingSourceSubscription, session.funding.Source())
	assert.Equal(t, BillingSourceSubscription, relayInfo.BillingSource)
	assert.Equal(t, subID, relayInfo.SubscriptionId)
	assert.Equal(t, preConsume, session.GetPreConsumedQuota())
	assert.Equal(t, initialQuota, getUserQuota(t, userID))
	assert.Equal(t, tokenQuota-preConsume, getTokenRemainQuota(t, tokenID))
	assert.Equal(t, subUsed+int64(preConsume), getSubscriptionUsed(t, subID))
}

func TestNewBillingSession_SubscriptionFirst_InsufficientSubscriptionFallsBackToWallet(t *testing.T) {
	prepareBillingSessionTables(t)
	truncate(t)

	const (
		userID       = 13
		tokenID      = 23
		subID        = 33
		initialQuota = 5000
		tokenQuota   = 4000
		preConsume   = 300
		subTotal     = 100
		subUsed      = 90
	)

	seedUser(t, userID, initialQuota)
	seedToken(t, tokenID, userID, "sk-sub-insufficient-wallet", tokenQuota)
	seedSubscription(t, subID, userID, subTotal, subUsed)

	ctx := makeBillingSessionContext()
	relayInfo := makeRelayInfoForBillingTest(
		userID,
		tokenID,
		"sk-sub-insufficient-wallet",
		"subscription_first",
		t.Name(),
		initialQuota,
	)

	session, apiErr := NewBillingSession(ctx, relayInfo, preConsume)
	require.Nil(t, apiErr)
	require.NotNil(t, session)

	assert.Equal(t, BillingSourceWallet, session.funding.Source())
	assert.Equal(t, BillingSourceWallet, relayInfo.BillingSource)
	assert.Equal(t, 0, relayInfo.SubscriptionId)
	assert.Equal(t, preConsume, session.GetPreConsumedQuota())
	assert.Equal(t, initialQuota-preConsume, getUserQuota(t, userID))
	assert.Equal(t, tokenQuota-preConsume, getTokenRemainQuota(t, tokenID))
	assert.EqualValues(t, subUsed, getSubscriptionUsed(t, subID))
}

func TestNewBillingSession_WalletOnly_InsufficientWalletReturnsError(t *testing.T) {
	prepareBillingSessionTables(t)
	truncate(t)

	const (
		userID       = 14
		tokenID      = 24
		initialQuota = 100
		tokenQuota   = 4000
		preConsume   = 300
	)

	seedUser(t, userID, initialQuota)
	seedToken(t, tokenID, userID, "sk-wallet-only-error", tokenQuota)

	ctx := makeBillingSessionContext()
	relayInfo := makeRelayInfoForBillingTest(
		userID,
		tokenID,
		"sk-wallet-only-error",
		"wallet_only",
		t.Name(),
		initialQuota,
	)

	session, apiErr := NewBillingSession(ctx, relayInfo, preConsume)
	require.Nil(t, session)
	require.NotNil(t, apiErr)

	assert.Equal(t, common.NormalizeBillingPreference("wallet_only"), relayInfo.UserSetting.BillingPreference)
	assert.Equal(t, initialQuota, getUserQuota(t, userID))
	assert.Equal(t, tokenQuota, getTokenRemainQuota(t, tokenID))
}
