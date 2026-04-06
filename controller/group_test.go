package controller

import (
	"fmt"
	"net/http"
	"slices"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupGroupControllerTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	gin.SetMode(gin.TestMode)
	common.UsingSQLite = true
	common.UsingMySQL = false
	common.UsingPostgreSQL = false
	common.RedisEnabled = false

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	model.DB = db
	model.LOG_DB = db

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

func TestGetGroups_ReturnsSortedGroupNames(t *testing.T) {
	setupGroupControllerTestDB(t)

	originalGroupRatio := ratio_setting.GroupRatio2JSONString()
	t.Cleanup(func() {
		_ = ratio_setting.UpdateGroupRatioByJSONString(originalGroupRatio)
	})

	if err := ratio_setting.UpdateGroupRatioByJSONString(`{"vip":1.5,"default":1,"svip":2}`); err != nil {
		t.Fatalf("failed to update group ratio: %v", err)
	}

	ctx, recorder := newAuthenticatedContext(t, http.MethodGet, "/api/group/", nil, 1)
	GetGroups(ctx)

	response := decodeAPIResponse(t, recorder)
	if !response.Success {
		t.Fatalf("expected success response, got message: %s", response.Message)
	}

	var groups []string
	if err := common.Unmarshal(response.Data, &groups); err != nil {
		t.Fatalf("failed to decode group list: %v", err)
	}

	if !slices.Equal(groups, []string{"default", "svip", "vip"}) {
		t.Fatalf("expected sorted groups, got %#v", groups)
	}
}

func TestGetUserGroups_IncludesAutoRatioLabel(t *testing.T) {
	db := setupGroupControllerTestDB(t)

	originalGroupRatio := ratio_setting.GroupRatio2JSONString()
	originalAutoGroups := setting.AutoGroups2JsonString()
	originalUsableGroups := setting.UserUsableGroups2JSONString()
	t.Cleanup(func() {
		_ = ratio_setting.UpdateGroupRatioByJSONString(originalGroupRatio)
		_ = setting.UpdateAutoGroupsByJsonString(originalAutoGroups)
		_ = setting.UpdateUserUsableGroupsByJSONString(originalUsableGroups)
	})

	if err := ratio_setting.UpdateGroupRatioByJSONString(`{"default":1,"vip":1.5}`); err != nil {
		t.Fatalf("failed to update group ratio: %v", err)
	}
	if err := setting.UpdateAutoGroupsByJsonString(`["vip"]`); err != nil {
		t.Fatalf("failed to update auto groups: %v", err)
	}
	if err := setting.UpdateUserUsableGroupsByJSONString(`{"default":"默认分组","vip":"VIP分组","auto":"自动选择可用分组"}`); err != nil {
		t.Fatalf("failed to update usable groups: %v", err)
	}

	user := &model.User{
		Id:       2001,
		Username: "group_user",
		Password: "password123",
		Status:   common.UserStatusEnabled,
		Role:     common.RoleCommonUser,
		Group:    "default",
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	ctx, recorder := newAuthenticatedContext(t, http.MethodGet, "/api/user/self/groups", nil, user.Id)
	GetUserGroups(ctx)

	response := decodeAPIResponse(t, recorder)
	if !response.Success {
		t.Fatalf("expected success response, got message: %s", response.Message)
	}

	var groups map[string]map[string]any
	if err := common.Unmarshal(response.Data, &groups); err != nil {
		t.Fatalf("failed to decode user groups: %v", err)
	}

	autoInfo, ok := groups["auto"]
	if !ok {
		t.Fatalf("expected auto group in response, got %#v", groups)
	}
	if autoInfo["ratio"] != "自动" {
		t.Fatalf("expected auto ratio label 自动, got %#v", autoInfo["ratio"])
	}
	if autoInfo["desc"] != "自动选择可用分组" {
		t.Fatalf("expected auto desc to round-trip, got %#v", autoInfo["desc"])
	}
}
