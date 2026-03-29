# 激活码功能需求文档

## 功能概述
为订阅系统添加"激活码"功能，参照现有的兑换码系统实现。激活码可以用于激活订阅套餐，而不是像兑换码那样充值额度。

## 需求分析

### 现有系统对比

#### 兑换码 (Redemption)
- **用途**: 充值用户额度 (quota)
- **数据模型**: `model/redemption.go`
- **核心字段**: `Quota` (充值额度)
- **兑换逻辑**: 增加用户的 quota 字段

#### 激活码 (SubscriptionCode) - 已实现
- **用途**: 激活订阅套餐
- **数据模型**: `model/subscription_code.go`
- **核心字段**: `PlanId` (订阅套餐ID)
- **兑换逻辑**: 创建用户订阅记录

## 后端实现状态

### ✅ 已完成的后端功能

1. **数据模型** (`model/subscription_code.go`)
   - ✅ `SubscriptionCode` 结构体定义
   - ✅ 包含所有必要字段: Id, UserId, Key, Status, Name, PlanId, CreatedTime, RedeemedTime, ExpiredTime 等

2. **数据库操作**
   - ✅ `GetAllSubscriptionCodes()` - 获取所有激活码
   - ✅ `SearchSubscriptionCodes()` - 搜索激活码
   - ✅ `GetSubscriptionCodeById()` - 根据ID获取
   - ✅ `RedeemSubscriptionCode()` - 兑换激活码
   - ✅ `Insert()`, `Update()`, `Delete()` - CRUD操作
   - ✅ `DeleteInvalidSubscriptionCodes()` - 删除无效激活码

3. **兑换逻辑**
   - ✅ 验证激活码有效性
   - ✅ 检查是否已使用
   - ✅ 检查是否过期
   - ✅ 创建用户订阅记录
   - ✅ 更新激活码状态
   - ✅ 记录操作日志

4. **API路由** (需要检查 `router/api-router.go`)
   - 需要添加管理端路由
   - 需要添加用户端兑换路由

5. **控制器** (需要检查 `controller/`)
   - 需要实现激活码管理控制器
   - 需要实现用户兑换控制器

## 前端实现需求

### 需要实现的功能

#### 1. 管理端 - 激活码管理页面
参照兑换码管理页面实现，需要包含:

**页面位置**: `web/src/pages/SubscriptionCode/` (新建)

**功能列表**:
- ✅ 激活码列表展示
  - 显示字段: ID, 名称, 激活码, 套餐名称, 状态, 创建时间, 兑换时间, 过期时间, 使用用户
  - 支持分页
  - 支持搜索

- ✅ 创建激活码
  - 输入字段:
    - 名称 (必填)
    - 选择订阅套餐 (下拉选择，必填)
    - 生成数量 (1-100)
    - 过期时间 (可选)
  - 批量生成激活码

- ✅ 编辑激活码
  - 可修改: 名称, 状态, 过期时间
  - 不可修改: 激活码、套餐ID (已使用后)

- ✅ 删除激活码
  - 单个删除
  - 批量删除无效激活码

- ✅ 状态管理
  - 启用/禁用激活码

#### 2. 用户端 - 激活码兑换
参照兑换码兑换功能实现

**页面位置**: `web/src/pages/Topup/` 或独立页面

**功能**:
- ✅ 激活码输入框
- ✅ 兑换按钮
- ✅ 兑换成功提示
- ✅ 错误处理 (无效、已使用、已过期等)

#### 3. 导航菜单
在管理端侧边栏添加"激活码管理"菜单项

### 参照文件

#### 兑换码相关文件 (作为参考)
```
web/src/pages/Redemption/
  ├── index.js                    # 兑换码列表页面
  ├── EditRedemption.js          # 编辑兑换码
  └── ...

web/src/components/
  └── ...
```

#### 需要创建的文件
```
web/src/pages/SubscriptionCode/
  ├── index.js                    # 激活码列表页面
  ├── EditSubscriptionCode.js    # 编辑激活码
  └── components/
      ├── SubscriptionCodeTable.js
      └── RedeemSubscriptionCodeModal.js

web/src/services/
  └── subscriptionCodeService.js  # API调用服务
```

## 国际化 (i18n)

### 需要添加的翻译键

#### 中文 (`web/src/i18n/locales/zh.json`)
```json
{
  "激活码": "激活码",
  "激活码管理": "激活码管理",
  "创建激活码": "创建激活码",
  "编辑激活码": "编辑激活码",
  "激活码名称": "激活码名称",
  "选择订阅套餐": "选择订阅套餐",
  "生成数量": "生成数量",
  "激活码已生成": "激活码已生成",
  "兑换激活码": "兑换激活码",
  "请输入激活码": "请输入激活码",
  "激活成功": "激活成功",
  "无效的激活码": "无效的激活码",
  "该激活码已被使用": "该激活码已被使用",
  "该激活码已过期": "该激活码已过期",
  "订阅套餐不存在": "订阅套餐不存在"
}
```

#### 英文 (`web/src/i18n/locales/en.json`)
```json
{
  "激活码": "Activation Code",
  "激活码管理": "Activation Code Management",
  "创建激活码": "Create Activation Code",
  "编辑激活码": "Edit Activation Code",
  "激活码名称": "Code Name",
  "选择订阅套餐": "Select Subscription Plan",
  "生成数量": "Generate Count",
  "激活码已生成": "Activation codes generated",
  "兑换激活码": "Redeem Activation Code",
  "请输入激活码": "Please enter activation code",
  "激活成功": "Activation successful",
  "无效的激活码": "Invalid activation code",
  "该激活码已被使用": "This activation code has been used",
  "该激活码已过期": "This activation code has expired",
  "订阅套餐不存在": "Subscription plan does not exist"
}
```

## API接口设计

### 管理端接口

```
GET    /api/subscription_code/           # 获取所有激活码
GET    /api/subscription_code/search     # 搜索激活码
GET    /api/subscription_code/:id        # 获取单个激活码
POST   /api/subscription_code/           # 创建激活码
PUT    /api/subscription_code/           # 更新激活码
DELETE /api/subscription_code/:id        # 删除激活码
DELETE /api/subscription_code/invalid   # 删除无效激活码
```

### 用户端接口

```
POST   /api/subscription_code/redeem     # 兑换激活码
```

## 实现步骤

### 第一阶段: 后端完善
1. ✅ 检查数据模型是否完整
2. ⬜ 实现控制器 (`controller/subscription_code.go`)
3. ⬜ 添加API路由 (`router/api-router.go`)
4. ⬜ 添加后端国际化消息 (`i18n/`)
5. ⬜ 测试后端API

### 第二阶段: 前端实现
1. ⬜ 创建激活码管理页面
2. ⬜ 实现激活码列表和CRUD功能
3. ⬜ 实现用户兑换功能
4. ⬜ 添加前端国际化
5. ⬜ 添加导航菜单
6. ⬜ 测试前端功能

### 第三阶段: 集成测试
1. ⬜ 端到端测试
2. ⬜ 边界情况测试
3. ⬜ 性能测试
4. ⬜ 文档更新

## 技术细节

### 数据库表结构
```sql
CREATE TABLE subscription_codes (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    key CHAR(32) UNIQUE NOT NULL,
    status INT DEFAULT 1,
    name VARCHAR(255),
    plan_id INT NOT NULL,
    created_time BIGINT,
    redeemed_time BIGINT,
    used_user_id INT,
    expired_time BIGINT,
    deleted_at TIMESTAMP NULL,
    INDEX idx_name (name),
    INDEX idx_deleted_at (deleted_at)
);
```

### 状态码定义
```go
const (
    RedemptionCodeStatusEnabled  = 1  // 启用
    RedemptionCodeStatusUsed     = 2  // 已使用
    RedemptionCodeStatusDisabled = 3  // 禁用
)
```

## 安全考虑

1. **激活码生成**: 使用 UUID 确保唯一性和随机性
2. **并发控制**: 使用数据库行锁 (`FOR UPDATE`) 防止重复兑换
3. **权限控制**: 
   - 管理端接口需要管理员权限
   - 用户端接口需要用户登录
4. **输入验证**: 
   - 验证激活码格式
   - 验证套餐ID有效性
   - 验证过期时间合法性

## 测试用例

### 功能测试
1. 创建激活码
2. 批量生成激活码
3. 编辑激活码
4. 删除激活码
5. 用户兑换激活码
6. 兑换已使用的激活码 (应失败)
7. 兑换已过期的激活码 (应失败)
8. 兑换无效的激活码 (应失败)

### 边界测试
1. 生成数量边界 (1, 100, 101)
2. 过期时间边界
3. 并发兑换同一激活码

## 注意事项

1. **与兑换码的区别**:
   - 兑换码充值额度，激活码激活订阅
   - 两者可以共存，互不影响

2. **数据库兼容性**:
   - 确保支持 SQLite, MySQL, PostgreSQL
   - 使用 GORM 抽象层

3. **前端组件复用**:
   - 尽量复用现有的 UI 组件
   - 保持界面风格一致

4. **错误处理**:
   - 提供清晰的错误提示
   - 记录详细的操作日志

## 参考资料

- 兑换码实现: `model/redemption.go`, `controller/redemption.go`
- 订阅系统: `model/subscription.go`, `model/subscription_code.go`
- 前端组件: `web/src/pages/Redemption/`