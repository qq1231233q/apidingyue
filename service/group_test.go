package service

import (
	"maps"
	"slices"
	"testing"

	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
)

func TestGetUserUsableGroupsReturnsAllConfiguredGroups(t *testing.T) {
	originalGroupRatio := ratio_setting.GroupRatio2JSONString()
	originalAutoGroups := setting.AutoGroups2JsonString()
	t.Cleanup(func() {
		_ = ratio_setting.UpdateGroupRatioByJSONString(originalGroupRatio)
		_ = setting.UpdateAutoGroupsByJsonString(originalAutoGroups)
	})

	if err := ratio_setting.UpdateGroupRatioByJSONString(`{"default":1,"vip":1.5,"svip":2}`); err != nil {
		t.Fatalf("failed to update group ratio: %v", err)
	}
	if err := setting.UpdateAutoGroupsByJsonString(`["vip","svip"]`); err != nil {
		t.Fatalf("failed to update auto groups: %v", err)
	}

	groups := GetUserUsableGroups("default")
	for _, groupName := range []string{"default", "vip", "svip", "auto"} {
		if _, ok := groups[groupName]; !ok {
			t.Fatalf("expected usable groups to contain %q, got %#v", groupName, groups)
		}
	}
}

func TestGetUserAutoGroupReturnsConfiguredAutoGroups(t *testing.T) {
	originalGroupRatio := ratio_setting.GroupRatio2JSONString()
	originalAutoGroups := setting.AutoGroups2JsonString()
	t.Cleanup(func() {
		_ = ratio_setting.UpdateGroupRatioByJSONString(originalGroupRatio)
		_ = setting.UpdateAutoGroupsByJsonString(originalAutoGroups)
	})

	if err := ratio_setting.UpdateGroupRatioByJSONString(`{"default":1,"vip":1.5,"svip":2}`); err != nil {
		t.Fatalf("failed to update group ratio: %v", err)
	}
	if err := setting.UpdateAutoGroupsByJsonString(`["vip","svip"]`); err != nil {
		t.Fatalf("failed to update auto groups: %v", err)
	}

	autoGroups := GetUserAutoGroup("default")
	if !slices.Equal(autoGroups, []string{"vip", "svip"}) {
		t.Fatalf("unexpected auto groups: %#v", autoGroups)
	}
}

func TestUpdateAutoGroupsByJsonString_InvalidJSONDoesNotClearPreviousConfig(t *testing.T) {
	originalAutoGroups := setting.AutoGroups2JsonString()
	t.Cleanup(func() {
		_ = setting.UpdateAutoGroupsByJsonString(originalAutoGroups)
	})

	expected := []string{"vip", "svip"}
	if err := setting.UpdateAutoGroupsByJsonString(`["vip","svip"]`); err != nil {
		t.Fatalf("failed to seed auto groups: %v", err)
	}

	if err := setting.UpdateAutoGroupsByJsonString(`{"invalid":true}`); err == nil {
		t.Fatal("expected invalid auto groups JSON to fail")
	}

	if actual := setting.GetAutoGroups(); !slices.Equal(actual, expected) {
		t.Fatalf("expected auto groups to remain %#v, got %#v", expected, actual)
	}
}

func TestUpdateUserUsableGroupsByJSONString_InvalidJSONDoesNotClearPreviousConfig(t *testing.T) {
	originalUsableGroups := setting.UserUsableGroups2JSONString()
	t.Cleanup(func() {
		_ = setting.UpdateUserUsableGroupsByJSONString(originalUsableGroups)
	})

	expected := map[string]string{
		"default": "默认分组",
		"vip":     "VIP分组",
	}
	if err := setting.UpdateUserUsableGroupsByJSONString(`{"default":"默认分组","vip":"VIP分组"}`); err != nil {
		t.Fatalf("failed to seed usable groups: %v", err)
	}

	if err := setting.UpdateUserUsableGroupsByJSONString(`["invalid"]`); err == nil {
		t.Fatal("expected invalid usable groups JSON to fail")
	}

	if actual := setting.GetUserUsableGroupsCopy(); !maps.Equal(actual, expected) {
		t.Fatalf("expected usable groups to remain %#v, got %#v", expected, actual)
	}
}
