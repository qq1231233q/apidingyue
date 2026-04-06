package controller

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupInternalSyncTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	gin.SetMode(gin.TestMode)
	common.UsingSQLite = true
	common.UsingMySQL = false
	common.UsingPostgreSQL = false

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	model.DB = db

	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("failed to migrate user table: %v", err)
	}

	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	return db
}

func TestValidateUserBindingInternal_Succeeds(t *testing.T) {
	db := setupInternalSyncTestDB(t)

	originalSecret := system_setting.InternalSyncSecret
	system_setting.InternalSyncSecret = "shared-secret"
	t.Cleanup(func() {
		system_setting.InternalSyncSecret = originalSecret
	})

	user := model.User{
		Username:    "agent_owner",
		Password:    "password123",
		DisplayName: "Agent Owner",
		Role:        common.RoleCommonUser,
		Status:      common.UserStatusEnabled,
		AffCode:     "AFF123",
		Group:       "default",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	body, err := common.Marshal(map[string]any{
		"user_id":  user.Id,
		"username": "agent_owner",
		"aff_code": "AFF123",
	})
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	router := gin.New()
	router.POST("/api/internal/user-binding/validate", middleware.InternalSyncAuth(), ValidateUserBindingInternal)

	req := httptest.NewRequest(http.MethodPost, "/api/internal/user-binding/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer shared-secret")
	req.RemoteAddr = "127.0.0.1:12345"
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			UserID      int    `json:"user_id"`
			AffCode     string `json:"aff_code"`
			Username    string `json:"username"`
			DisplayName string `json:"display_name"`
			Status      int    `json:"status"`
		} `json:"data"`
	}
	if err := common.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !response.Success {
		t.Fatalf("expected success response, got body: %s", recorder.Body.String())
	}
	if response.Data.UserID != user.Id || response.Data.AffCode != "AFF123" || response.Data.Status != common.UserStatusEnabled {
		t.Fatalf("unexpected response payload: %+v", response.Data)
	}
}

func TestValidateUserBindingInternal_RejectsUsernameMismatch(t *testing.T) {
	db := setupInternalSyncTestDB(t)

	originalSecret := system_setting.InternalSyncSecret
	system_setting.InternalSyncSecret = "shared-secret"
	t.Cleanup(func() {
		system_setting.InternalSyncSecret = originalSecret
	})

	user := model.User{
		Username:    "agent_owner",
		Password:    "password123",
		DisplayName: "Agent Owner",
		Role:        common.RoleCommonUser,
		Status:      common.UserStatusEnabled,
		AffCode:     "AFF123",
		Group:       "default",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	body, err := common.Marshal(map[string]any{
		"user_id":  user.Id,
		"username": "wrong_name",
		"aff_code": "AFF123",
	})
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	router := gin.New()
	router.POST("/api/internal/user-binding/validate", middleware.InternalSyncAuth(), ValidateUserBindingInternal)

	req := httptest.NewRequest(http.MethodPost, "/api/internal/user-binding/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer shared-secret")
	req.RemoteAddr = "127.0.0.1:12345"
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestValidateUserBindingInternal_RejectsInvalidSecret(t *testing.T) {
	setupInternalSyncTestDB(t)

	originalSecret := system_setting.InternalSyncSecret
	system_setting.InternalSyncSecret = "shared-secret"
	t.Cleanup(func() {
		system_setting.InternalSyncSecret = originalSecret
	})

	body, err := common.Marshal(map[string]any{
		"user_id":  1,
		"username": "agent_owner",
		"aff_code": "AFF123",
	})
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	router := gin.New()
	router.POST("/api/internal/user-binding/validate", middleware.InternalSyncAuth(), ValidateUserBindingInternal)

	req := httptest.NewRequest(http.MethodPost, "/api/internal/user-binding/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer wrong-secret")
	req.RemoteAddr = "127.0.0.1:12345"
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestValidateUserBindingInternal_RejectsExternalSource(t *testing.T) {
	setupInternalSyncTestDB(t)

	originalSecret := system_setting.InternalSyncSecret
	system_setting.InternalSyncSecret = "shared-secret"
	t.Cleanup(func() {
		system_setting.InternalSyncSecret = originalSecret
	})

	body, err := common.Marshal(map[string]any{
		"user_id":  1,
		"username": "agent_owner",
		"aff_code": "AFF123",
	})
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	router := gin.New()
	router.POST("/api/internal/user-binding/validate", middleware.InternalSyncAuth(), ValidateUserBindingInternal)

	req := httptest.NewRequest(http.MethodPost, "/api/internal/user-binding/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer shared-secret")
	req.RemoteAddr = "8.8.8.8:12345"
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d: %s", recorder.Code, recorder.Body.String())
	}
}
