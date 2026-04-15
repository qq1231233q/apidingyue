package controller

import (
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	sessionKeyInviteAffCode   = "invite_aff_code"
	sessionKeyInviteInviterID = "invite_inviter_id"
	sessionKeyInviteLockedAt  = "invite_locked_at"
	inviteBindingTTLSeconds   = 2 * 60 * 60
)

func getLockedInviteBinding(c *gin.Context) (int, string, bool) {
	session := sessions.Default(c)
	lockedAt, ok := session.Get(sessionKeyInviteLockedAt).(int64)
	if !ok || lockedAt <= 0 {
		if lockedAtInt, ok := session.Get(sessionKeyInviteLockedAt).(int); ok {
			lockedAt = int64(lockedAtInt)
		}
	}
	if lockedAt <= 0 {
		return 0, "", false
	}
	if common.GetTimestamp()-lockedAt > inviteBindingTTLSeconds {
		clearInviteBinding(c)
		return 0, "", false
	}

	affCode, _ := session.Get(sessionKeyInviteAffCode).(string)
	inviterID := 0
	switch v := session.Get(sessionKeyInviteInviterID).(type) {
	case int:
		inviterID = v
	case int64:
		inviterID = int(v)
	case float64:
		inviterID = int(v)
	}
	if inviterID <= 0 || affCode == "" {
		return 0, "", false
	}
	return inviterID, affCode, true
}

func lockInviteBinding(c *gin.Context, affCode string) {
	affCode = strings.TrimSpace(affCode)
	if affCode == "" {
		return
	}
	if inviterID, _, ok := getLockedInviteBinding(c); ok && inviterID > 0 {
		return
	}

	inviterID, err := model.GetUserIdByAffCode(affCode)
	if err != nil || inviterID <= 0 {
		common.SysLog(fmt.Sprintf("ignore invalid invite aff_code=%q ip=%s path=%s", affCode, c.ClientIP(), c.Request.URL.Path))
		return
	}

	session := sessions.Default(c)
	session.Set("aff", affCode)
	session.Set(sessionKeyInviteAffCode, affCode)
	session.Set(sessionKeyInviteInviterID, inviterID)
	session.Set(sessionKeyInviteLockedAt, int64(common.GetTimestamp()))
	if err := session.Save(); err != nil {
		common.SysError("failed to save invite binding session: " + err.Error())
		return
	}
	common.SysLog(fmt.Sprintf("locked invite binding aff_code=%q inviter_id=%d ip=%s path=%s", affCode, inviterID, c.ClientIP(), c.Request.URL.Path))
}

func LockInviteBindingForRegisterEntry(c *gin.Context) {
	lockInviteBinding(c, c.Query("aff"))
}

func resolveInviteBinding(c *gin.Context) (int, string) {
	inviterID, affCode, ok := getLockedInviteBinding(c)
	if !ok {
		return 0, ""
	}
	return inviterID, affCode
}

func clearInviteBinding(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete("aff")
	session.Delete(sessionKeyInviteAffCode)
	session.Delete(sessionKeyInviteInviterID)
	session.Delete(sessionKeyInviteLockedAt)
	if err := session.Save(); err != nil {
		common.SysError("failed to clear invite binding session: " + err.Error())
	}
}
