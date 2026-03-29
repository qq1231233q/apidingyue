# 订阅可用分组功能实现总结

## 功能说明
为订阅系统添加"可用分组"功能,允许管理员限制订阅额度只能用于特定分组的模型。

## 已完成的修改

### 1. 后端数据模型 (model/subscription.go)
- ✅ 在 `SubscriptionPlan` 结构体中添加 `AvailableGroup` 字段
- ✅ 添加 `isModelInGroup()` 辅助函数检查模型是否属于指定分组
- ✅ 修改 `PreConsumeUserSubscription()` 函数,在消费订阅额度时验证模型分组

### 2. 数据库迁移 (model/migration_add_available_group.go)
- ✅ 创建迁移函数 `MigrateAddAvailableGroupToSubscriptionPlan()`
- ✅ 支持 SQLite、MySQL、PostgreSQL 三种数据库
- ✅ 使用 `IF NOT EXISTS` 确保迁移可安全重复执行

### 3. 数据库初始化 (model/main.go)
- ✅ 在 `migrateDB()` 函数中调用迁移函数
- ✅ 在 `ensureSubscriptionPlanTableSQLite()` 中添加 `available_group` 列定义

### 4. 前端管理界面 (web/src/components/table/subscriptions/modals/AddEditSubscriptionModal.jsx)
- ✅ 添加"可用分组"下拉选择字段
- ✅ 从后端加载所有分组选项
- ✅ 表单提交时包含 `available_group` 字段

## 使用说明

### 管理员操作
1. 进入订阅管理页面
2. 创建或编辑订阅套餐
3. 在"可用分组"字段中选择一个分组(或留空表示不限制)
4. 保存套餐

### 用户使用
- 当用户使用订阅额度调用 API 时:
  - 如果订阅设置了可用分组,系统会检查请求的模型是否属于该分组
  - 只有匹配的订阅才会被消费
  - 如果没有匹配的订阅,会尝试使用用户的普通额度

## 技术细节

### 数据库字段
```sql
available_group VARCHAR(64) DEFAULT ''
```
- 空字符串表示不限制,可用于所有分组
- 非空值表示只能用于指定分组的模型

### 分组检查逻辑
```go
func isModelInGroup(modelName string, group string) (bool, error) {
    // 查询 abilities 表检查模型是否属于指定分组
    var count int64
    err := DB.Model(&Ability{}).
        Where(commonGroupCol+" = ? AND model = ? AND enabled = ?", group, modelName, true).
        Count(&count).Error
    return count > 0, err
}
```

### 消费逻辑
在 `PreConsumeUserSubscription()` 中:
1. 获取用户的所有活跃订阅
2. 遍历每个订阅
3. 如果订阅设置了 `available_group`,检查模型是否在该分组中
4. 不匹配则跳过该订阅,继续检查下一个

## 下一步操作

1. **安装 Go** (如果还没有):
   - 访问 https://golang.org/dl/
   - 下载并安装适合你系统的版本
   - 确保 `go` 命令在 PATH 中

2. **编译项目**:
   ```bash
   go build -o new-api.exe
   ```

3. **启动应用**:
   - 数据库迁移会在启动时自动执行
   - 检查日志确认迁移成功

4. **测试功能**:
   - 创建一个订阅套餐并设置可用分组
   - 使用该订阅调用不同分组的模型
   - 验证只有匹配分组的模型能消费订阅额度

## 兼容性
- ✅ SQLite
- ✅ MySQL 5.7.8+
- ✅ PostgreSQL 9.6+
- ✅ 向后兼容:现有订阅的 `available_group` 默认为空,不影响现有功能

## 注意事项
- 迁移脚本会在应用启动时自动执行
- 如果迁移失败,会记录警告日志但不会阻止应用启动
- 建议在生产环境部署前先在测试环境验证