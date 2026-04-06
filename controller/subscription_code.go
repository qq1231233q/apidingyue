package controller

import (
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/ratio_setting"

	"github.com/gin-gonic/gin"
)

func validateSubscriptionCodeConfig(code *model.SubscriptionCode) string {
	if code == nil {
		return "参数错误"
	}
	if code.Quota <= 0 {
		return "充值额度必须大于 0"
	}
	if code.DurationUnit == "" {
		code.DurationUnit = model.SubscriptionDurationMonth
	}
	switch code.DurationUnit {
	case model.SubscriptionDurationYear,
		model.SubscriptionDurationMonth,
		model.SubscriptionDurationDay,
		model.SubscriptionDurationHour,
		model.SubscriptionDurationCustom:
	default:
		return "无效的时长单位"
	}
	if code.DurationUnit == model.SubscriptionDurationCustom {
		if code.CustomSeconds <= 0 {
			return "自定义秒数必须大于 0"
		}
	} else if code.DurationValue <= 0 {
		return "时长数值必须大于 0"
	}
	code.AvailableGroup = strings.TrimSpace(code.AvailableGroup)
	if code.AvailableGroup != "" {
		if _, ok := ratio_setting.GetGroupRatioCopy()[code.AvailableGroup]; !ok {
			return "可用分组不存在"
		}
	}
	return ""
}

func GetAllSubscriptionCodes(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	codes, total, err := model.GetAllSubscriptionCodes(pageInfo.GetStartIdx(), pageInfo.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(codes)
	common.ApiSuccess(c, pageInfo)
}

func SearchSubscriptionCodes(c *gin.Context) {
	keyword := c.Query("keyword")
	if len(keyword) > 100 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "搜索关键词过长",
		})
		return
	}

	pageInfo := common.GetPageQuery(c)
	codes, total, err := model.SearchSubscriptionCodes(keyword, pageInfo.GetStartIdx(), pageInfo.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(codes)
	common.ApiSuccess(c, pageInfo)
}

func GetSubscriptionCode(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	code, err := model.GetSubscriptionCodeById(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    code,
	})
}

func AddSubscriptionCode(c *gin.Context) {
	code := model.SubscriptionCode{}
	if err := c.ShouldBindJSON(&code); err != nil {
		common.ApiError(c, err)
		return
	}
	if utf8.RuneCountInString(code.Name) == 0 || utf8.RuneCountInString(code.Name) > 20 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "激活码名称长度必须在 1-20 之间",
		})
		return
	}
	if code.Count <= 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "生成数量必须大于 0",
		})
		return
	}
	if code.Count > 100 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "单次最多生成 100 个激活码",
		})
		return
	}
	if valid, msg := validateExpiredTime(c, code.ExpiredTime); !valid {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": msg})
		return
	}
	if msg := validateSubscriptionCodeConfig(&code); msg != "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": msg,
		})
		return
	}

	var keys []string
	for i := 0; i < code.Count; i++ {
		key := common.GetUUID()
		cleanCode := model.SubscriptionCode{
			UserId:         c.GetInt("id"),
			Name:           code.Name,
			Key:            key,
			Quota:          code.Quota,
			DurationUnit:   code.DurationUnit,
			DurationValue:  code.DurationValue,
			CustomSeconds:  code.CustomSeconds,
			AvailableGroup: code.AvailableGroup,
			CreatedTime:    common.GetTimestamp(),
			ExpiredTime:    code.ExpiredTime,
		}
		if err := cleanCode.Insert(); err != nil {
			common.SysError("failed to insert subscription code: " + err.Error())
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "创建激活码失败",
				"data":    keys,
			})
			return
		}
		keys = append(keys, key)
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    keys,
	})
}

func DeleteSubscriptionCode(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := model.DeleteSubscriptionCodeById(id); err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

func UpdateSubscriptionCode(c *gin.Context) {
	statusOnly := c.Query("status_only")
	code := model.SubscriptionCode{}
	if err := c.ShouldBindJSON(&code); err != nil {
		common.ApiError(c, err)
		return
	}
	cleanCode, err := model.GetSubscriptionCodeById(code.Id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if statusOnly == "" {
		if utf8.RuneCountInString(code.Name) == 0 || utf8.RuneCountInString(code.Name) > 20 {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "激活码名称长度必须在 1-20 之间",
			})
			return
		}
		if valid, msg := validateExpiredTime(c, code.ExpiredTime); !valid {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": msg})
			return
		}
		if msg := validateSubscriptionCodeConfig(&code); msg != "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": msg,
			})
			return
		}
		cleanCode.Name = code.Name
		cleanCode.Quota = code.Quota
		cleanCode.DurationUnit = code.DurationUnit
		cleanCode.DurationValue = code.DurationValue
		cleanCode.CustomSeconds = code.CustomSeconds
		cleanCode.AvailableGroup = code.AvailableGroup
		cleanCode.ExpiredTime = code.ExpiredTime
	}
	if statusOnly != "" {
		cleanCode.Status = code.Status
	}
	if err := cleanCode.Update(); err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    cleanCode,
	})
}

func DeleteInvalidSubscriptionCodes(c *gin.Context) {
	rows, err := model.DeleteInvalidSubscriptionCodes()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    rows,
	})
}


func RedeemSubscriptionCode(c *gin.Context) {
	type RedeemRequest struct {
		Code string `json:"code"`
	}

	var req RedeemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "请输入激活码",
		})
		return
	}
	if req.Code == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "请输入激活码",
		})
		return
	}
	if len(req.Code) != 32 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "激活码格式无效",
		})
		return
	}

	userId := c.GetInt("id")
	if userId == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户未登录",
		})
		return
	}
	quota, err := model.RedeemSubscriptionCode(req.Code, userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "兑换成功",
		"data":    quota,
	})
}
