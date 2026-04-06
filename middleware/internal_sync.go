package middleware

import (
	"crypto/subtle"
	"net"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/setting/system_setting"

	"github.com/gin-gonic/gin"
)

func InternalSyncAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		if !isInternalSyncSourceAllowed(c.Request.RemoteAddr) {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "internal sync source is not allowed",
			})
			c.Abort()
			return
		}

		expectedSecret := strings.TrimSpace(system_setting.InternalSyncSecret)
		if expectedSecret == "" {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"success": false,
				"message": "internal sync secret is not configured",
			})
			c.Abort()
			return
		}

		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "missing internal sync authorization",
			})
			c.Abort()
			return
		}

		providedSecret := strings.TrimSpace(authHeader[7:])
		if subtle.ConstantTimeCompare([]byte(providedSecret), []byte(expectedSecret)) != 1 {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "invalid internal sync authorization",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func isInternalSyncSourceAllowed(remoteAddr string) bool {
	host := strings.TrimSpace(remoteAddr)
	if host == "" {
		return false
	}

	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		host = parsedHost
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}

	return ip.IsLoopback() || ip.IsPrivate()
}
