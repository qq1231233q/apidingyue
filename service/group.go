package service

import (
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
)

func GetUserUsableGroups(userGroup string) map[string]string {
	groupsCopy := make(map[string]string)
	for groupName := range ratio_setting.GetGroupRatioCopy() {
		groupsCopy[groupName] = setting.GetUsableGroupDescription(groupName)
	}
	if userGroup != "" {
		if _, ok := groupsCopy[userGroup]; !ok {
			groupsCopy[userGroup] = setting.GetUsableGroupDescription(userGroup)
		}
	}
	if len(setting.GetAutoGroups()) > 0 {
		groupsCopy["auto"] = setting.GetUsableGroupDescription("auto")
	}
	return groupsCopy
}

func GroupInUserUsableGroups(userGroup, groupName string) bool {
	_, ok := GetUserUsableGroups(userGroup)[groupName]
	return ok
}

// GetUserAutoGroup returns the configured auto groups that are currently usable.
func GetUserAutoGroup(userGroup string) []string {
	groups := GetUserUsableGroups(userGroup)
	autoGroups := make([]string, 0)
	for _, group := range setting.GetAutoGroups() {
		if _, ok := groups[group]; ok {
			autoGroups = append(autoGroups, group)
		}
	}
	return autoGroups
}

func GetUserGroupRatio(userGroup, group string) float64 {
	ratio, ok := ratio_setting.GetGroupGroupRatio(userGroup, group)
	if ok {
		return ratio
	}
	return ratio_setting.GetGroupRatio(group)
}
