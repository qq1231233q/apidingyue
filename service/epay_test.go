package service

import (
	"testing"

	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/stretchr/testify/require"
)

func withEpaySettingBackup(t *testing.T) {
	t.Helper()

	originCustomCallbackAddress := operation_setting.CustomCallbackAddress
	originServerAddress := system_setting.ServerAddress
	originPayAddress := operation_setting.PayAddress

	t.Cleanup(func() {
		operation_setting.CustomCallbackAddress = originCustomCallbackAddress
		system_setting.ServerAddress = originServerAddress
		operation_setting.PayAddress = originPayAddress
	})
}

func TestGetCallbackAddress_UseCustomAddress(t *testing.T) {
	withEpaySettingBackup(t)

	operation_setting.CustomCallbackAddress = " https://callback.example.com/ "
	system_setting.ServerAddress = "https://api.tokln.com"

	require.Equal(t, "https://callback.example.com", GetCallbackAddress())
}

func TestGetCallbackAddress_FallbackToSystemAddress(t *testing.T) {
	withEpaySettingBackup(t)

	operation_setting.CustomCallbackAddress = "<nil>"
	system_setting.ServerAddress = "https://api.tokln.com"

	require.Equal(t, "https://api.tokln.com", GetCallbackAddress())
}

func TestGetCallbackAddress_FallbackToDefaultAddress(t *testing.T) {
	withEpaySettingBackup(t)

	operation_setting.CustomCallbackAddress = "<nil>"
	system_setting.ServerAddress = "%3Cnil%3E"

	require.Equal(t, defaultEpayAddress, GetCallbackAddress())
}

func TestBuildAbsoluteURL_StripsNilLikeAddress(t *testing.T) {
	u := BuildAbsoluteURL("<nil>", "/api/user/epay/notify")
	require.Equal(t, defaultEpayAddress+"/api/user/epay/notify", u.String())
}

func TestResolveFrontendPayURL_UseConfiguredGatewayEndpoint(t *testing.T) {
	payAddress := "https://api.xunhupay.com/payment/do.html"
	purchaseURL := "https://api.xunhupay.com/payment/do.html/submit.php"

	require.Equal(t, payAddress, ResolveFrontendPayURL(payAddress, purchaseURL))
}

func TestResolveFrontendPayURL_KeepLibraryURLWhenBaseIsDirectory(t *testing.T) {
	payAddress := "https://api.xunhupay.com/payment"
	purchaseURL := "https://api.xunhupay.com/payment/submit.php"

	require.Equal(t, purchaseURL, ResolveFrontendPayURL(payAddress, purchaseURL))
}
