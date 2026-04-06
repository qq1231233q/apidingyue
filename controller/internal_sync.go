package controller

import (
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

type ValidateUserBindingInternalRequest struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	AffCode  string `json:"aff_code"`
}

func ValidateUserBindingInternal(c *gin.Context) {
	var req ValidateUserBindingInternalRequest
	if err := common.DecodeJson(c.Request.Body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid request body",
		})
		return
	}

	if req.UserID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid user_id",
		})
		return
	}

	normalizedUsername := strings.TrimSpace(req.Username)
	if normalizedUsername == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "username is required",
		})
		return
	}

	normalizedAffCode := strings.TrimSpace(req.AffCode)
	if normalizedAffCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "aff_code is required",
		})
		return
	}

	user, err := model.GetUserById(req.UserID, false)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "user not found",
		})
		return
	}

	if strings.TrimSpace(user.Username) != normalizedUsername {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"message": "user_id and username do not match",
		})
		return
	}

	remoteAffCode := strings.TrimSpace(user.AffCode)
	if remoteAffCode == "" {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"message": "user aff_code is empty",
		})
		return
	}

	if remoteAffCode != normalizedAffCode {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"message": "user_id and aff_code do not match",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"user_id":      user.Id,
			"aff_code":     remoteAffCode,
			"username":     user.Username,
			"display_name": user.DisplayName,
			"status":       user.Status,
		},
	})
}
