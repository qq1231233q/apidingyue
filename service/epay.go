package service

import (
	"net/url"
	"path"
	"strings"

	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/QuantumNous/new-api/setting/system_setting"
)

const defaultEpayAddress = "http://localhost:3000"

func sanitizeAddress(raw string) string {
	address := strings.TrimSpace(raw)
	if address == "" {
		return ""
	}

	lower := strings.ToLower(address)
	if lower == "<nil>" || lower == "%3cnil%3e" || lower == "nil" || lower == "null" {
		return ""
	}

	return strings.TrimRight(address, "/")
}

func GetCallbackAddress() string {
	if custom := sanitizeAddress(operation_setting.CustomCallbackAddress); custom != "" {
		return custom
	}
	if systemAddress := sanitizeAddress(system_setting.ServerAddress); systemAddress != "" {
		return systemAddress
	}
	return defaultEpayAddress
}

func BuildAbsoluteURL(baseAddress, endpointPath string) *url.URL {
	base := sanitizeAddress(baseAddress)
	if base == "" {
		base = defaultEpayAddress
	}

	u, err := url.Parse(base + endpointPath)
	if err == nil {
		return u
	}

	fallback, _ := url.Parse(defaultEpayAddress + endpointPath)
	return fallback
}

func ResolveFrontendPayURL(payAddress, purchaseURL string) string {
	cleanPayAddress := sanitizeAddress(payAddress)
	if cleanPayAddress == "" {
		return purchaseURL
	}

	parsedPayAddress, err := url.Parse(cleanPayAddress)
	if err != nil || parsedPayAddress.Scheme == "" || parsedPayAddress.Host == "" {
		return purchaseURL
	}

	baseName := path.Base(parsedPayAddress.Path)
	// If user configures a specific gateway script/html endpoint,
	// return it directly instead of appending /submit.php.
	if strings.Contains(baseName, ".") {
		return parsedPayAddress.String()
	}

	return purchaseURL
}
