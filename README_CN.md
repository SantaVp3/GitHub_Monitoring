# GitHub 监控平台

一个全面的 GitHub 信息泄漏监控平台，帮助组织检测和防止敏感数据在 GitHub 仓库上泄露。

**版本**: 0.2
**状态**: 生产就绪
**Language**: [English Documentation](./README.md)

---

## 概述

本平台基于可自定义的规则和关键词自动监控 GitHub 上的潜在信息泄漏。支持令牌轮换、多种通知渠道，并提供用户友好的 Web 界面进行管理。

### 核心功能

- **自动化监控**: 基于自定义规则持续扫描 GitHub 仓库
- **令牌池管理**: 自动轮换 GitHub API 令牌以处理速率限制
- **多渠道通知**: 支持企业微信、钉钉、飞书和自定义 Webhook
- **灵活匹配**: 支持模糊匹配和精确匹配算法
- **白名单系统**: 过滤已知安全的仓库和用户
- **批量操作**: 高效管理大量搜索结果
- **代理支持**: 支持 HTTP、HTTPS 和 SOCKS5 代理配置
- **JWT 认证**: 通过密码保护实现安全访问控制
- **实时仪表板**: 一目了然地监控系统状态和统计信息

---

## 技术栈

### 前端
- **框架**: React 18 + TypeScript
- **构建工具**: Vite 7.2.2
- **UI 组件**: shadcn/ui
- **样式**: Tailwind CSS 3.4.0
- **HTTP 客户端**: Axios
- **图标**: Lucide React

### 后端
- **语言**: Go (Golang)
- **Web 框架**: Gin
- **ORM**: GORM
- **数据库**: MySQL 8.x
- **认证**: JWT (golang-jwt/jwt/v5)
- **配置管理**: Viper
- **GitHub API**: google/go-github/v57

---

## 安装

### 前置要求

- Node.js 16+ 和 npm
- Go 1.18+
- MySQL 8.0+

### 后端设置

1. 克隆仓库:
```bash
git clone <repository-url>
cd GitHub-Monitoring/backend
```

2. 安装 Go 依赖:
```bash
go mod download
```

3. 配置数据库:
```bash
mysql -u root -p
CREATE DATABASE github_monitor CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

4. 配置应用程序:
```bash
cp config.yaml.example config.yaml
# 编辑 config.yaml 填入您的配置
```

5. 运行后端:
```bash
go run main.go
```

后端服务将在 `http://localhost:8080` 启动

### 前端设置

1. 进入前端目录:
```bash
cd ../frontend
```

2. 安装依赖:
```bash
npm install
```

3. 启动开发服务器:
```bash
npm run dev
```

前端将在 `http://localhost:5174` 可用

---

## 配置

### 后端配置 (config.yaml)

```yaml
server:
  port: 8080
  mode: debug  # 生产环境使用 "release"

database:
  host: localhost
  port: 3306
  user: root
  password: your_password
  database: github_monitor

auth:
  enabled: true
  password: "admin123"  # 请修改此密码！
  jwt_secret: "your-secret-key"  # 请修改此密钥！
  token_expiry: "24h"

github:
  tokens:
    - token: "ghp_your_token_1"
      name: "令牌 1"
    - token: "ghp_your_token_2"
      name: "令牌 2"

  # 代理配置（可选）
  proxy_enabled: false
  proxy_url: ""
  proxy_type: "http"  # http, https, 或 socks5
  proxy_username: ""
  proxy_password: ""

monitor:
  scan_interval: "5m"  # 扫描间隔
  max_results_per_rule: 100
```

### GitHub Token 所需权限

要使用本平台，您需要具有以下范围的 GitHub Personal Access Token：
- `public_repo` - 搜索公开仓库
- `repo`（可选）- 如果需要搜索私有仓库

在此生成令牌：https://github.com/settings/tokens

---

## 使用方法

### 初始设置

1. 访问登录页面 `http://localhost:5174`
2. 使用默认密码登录：`admin123`
3. 在 `backend/config.yaml` 中修改密码（推荐）

### 添加 GitHub Token

1. 导航到 **Settings** 页面
2. 展开 **GitHub Tokens** 部分
3. 点击 **Add Token**
4. 输入令牌名称和令牌值
5. 点击 **Add Token** 保存

### 创建监控规则

1. 导航到 **Monitor Rules** 页面
2. 点击 **Add Rule**
3. 填写表单：
   - **规则名称**: 规则的描述性名称
   - **匹配类型**: 选择模糊匹配或精确匹配
   - **关键词**: 逗号分隔的关键词（例如：`password, api_key, secret`）
   - **描述**: 可选描述
   - **激活**: 勾选以立即启用
4. 点击 **Create Rule**

### 管理搜索结果

1. 导航到 **Search Results** 页面
2. 查看检测到的潜在泄漏
3. 使用复选框选择多个结果
4. 使用批量操作：
   - **标记为已确认**: 标记为真实泄漏
   - **标记为误报**: 标记为安全

### 配置通知

1. 导航到 **Settings** 页面
2. 展开 **Notification Channels** 部分
3. 点击 **Add Channel**
4. 配置：
   - **名称**: 渠道标识符
   - **类型**: 选择企业微信、钉钉、飞书或 Webhook
   - **Webhook URL**: 您的 Webhook 端点
   - **Secret**: 用于钉钉/飞书签名验证
   - **通知条件**: 选择何时接收通知
5. 点击 **Create Channel**
6. 使用 **Test** 按钮测试通知

### 使用白名单

1. 导航到 **Whitelist** 页面
2. 点击 **Add to Whitelist**
3. 选择类型：
   - **用户**: 将 GitHub 用户加入白名单
   - **仓库**: 将特定仓库加入白名单
4. 输入值和可选描述
5. 点击 **Add**

---

## API 文档

### 认证

所有 API 端点（除了 `/api/v1/login`）都需要 JWT 认证。

**登录**
```http
POST /api/v1/login
Content-Type: application/json

{
  "password": "your-password"
}

响应:
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "message": "Login successful"
}
```

**认证请求**
```http
GET /api/v1/dashboard/stats
Authorization: Bearer <your-token>
```

### API 端点

#### 仪表板
- `GET /api/v1/dashboard/stats` - 获取仪表板统计数据

#### Token 管理
- `GET /api/v1/tokens` - 列出所有令牌
- `POST /api/v1/tokens` - 创建新令牌
- `DELETE /api/v1/tokens/:id` - 删除令牌
- `GET /api/v1/tokens/stats` - 获取令牌使用统计

#### 监控规则
- `GET /api/v1/rules` - 列出所有规则
- `GET /api/v1/rules/:id` - 获取特定规则
- `POST /api/v1/rules` - 创建新规则
- `PUT /api/v1/rules/:id` - 更新规则
- `DELETE /api/v1/rules/:id` - 删除规则

#### 搜索结果
- `GET /api/v1/results` - 列出搜索结果（支持分页）
- `PUT /api/v1/results/:id` - 更新结果状态
- `POST /api/v1/results/batch` - 批量更新结果状态

#### 白名单
- `GET /api/v1/whitelist` - 列出白名单条目
- `POST /api/v1/whitelist` - 添加白名单条目
- `DELETE /api/v1/whitelist/:id` - 删除白名单条目

#### 监控控制
- `GET /api/v1/monitor/status` - 获取监控服务状态
- `POST /api/v1/monitor/start` - 启动监控
- `POST /api/v1/monitor/stop` - 停止监控

#### 通知
- `GET /api/v1/notifications` - 列出通知渠道
- `POST /api/v1/notifications` - 创建通知渠道
- `PUT /api/v1/notifications/:id` - 更新通知渠道
- `DELETE /api/v1/notifications/:id` - 删除通知渠道
- `POST /api/v1/notifications/:id/test` - 测试通知渠道

#### 扫描历史
- `GET /api/v1/history` - 获取扫描历史（支持分页）

---

## 架构

### 系统组件

1. **前端（React + TypeScript）**
   - 管理和监控的用户界面
   - 实时数据更新
   - 响应式设计

2. **后端（Go + Gin）**
   - RESTful API 服务器
   - JWT 认证
   - 后台监控服务
   - 令牌池管理

3. **数据库（MySQL）**
   - 数据持久化
   - 搜索结果存储
   - 配置管理

4. **GitHub API 集成**
   - 代码搜索功能
   - 速率限制处理
   - 代理支持

### 数据模型

**GitHubToken**: 存储用于轮换的 GitHub API 令牌
**MonitorRule**: 定义监控规则和关键词
**SearchResult**: 存储检测到的潜在泄漏
**Whitelist**: 包含白名单用户和仓库
**ScanHistory**: 记录扫描活动
**NotificationConfig**: 通知渠道配置

---

## 安全注意事项

### 密码保护
- 立即更改默认密码
- 使用混合字符的强密码
- 在 `config.yaml` 中安全存储密码

### JWT 密钥
- 生成随机的长密钥
- 永远不要将密钥提交到版本控制
- 定期轮换密钥

### 令牌安全
- 使用最小必需权限的令牌
- 定期轮换令牌
- 监控令牌使用情况

### 生产环境使用 HTTPS
- 在生产环境中始终使用 HTTPS
- 保护令牌传输
- 使用安全的 WebSocket 连接

### 配置文件
- 设置适当的文件权限（600 或 400）
- 将 `config.yaml` 添加到 `.gitignore`
- 对敏感数据使用环境变量

---

## 故障排除

### 问题：成功认证后立即登录失败

**可能原因**:
- JWT 密钥不匹配
- Token 格式错误
- 系统时间不正确

**解决方案**:
- 验证 `config.yaml` 中的 `jwt_secret`
- 检查浏览器控制台错误
- 确保系统时间正确

### 问题：401 未授权错误

**可能原因**:
- Token 已过期
- Token 无效
- 认证中间件配置错误

**解决方案**:
- 重新登录获取新 Token
- 检查后端日志
- 验证 `auth.enabled` 配置

### 问题：GitHub API 返回 401 错误凭据

**可能原因**:
- GitHub Token 无效
- Token 已过期
- Token 权限错误

**解决方案**:
- 在 https://github.com/settings/tokens 生成新 Token
- 通过 Settings 页面添加 Token
- 确保 Token 具有 `public_repo` 权限

### 问题：没有搜索结果

**可能原因**:
- 没有活跃的监控规则
- 关键词过于具体
- 白名单过滤太宽泛

**解决方案**:
- 创建并激活监控规则
- 使用更常见的关键词
- 检查白名单条目

---

## 监控服务

本平台包含一个后台监控服务，它：
- 每 5 分钟运行一次（可配置）
- 执行所有活跃的监控规则
- 速率受限时自动轮换令牌
- 将结果存储在数据库中
- 在配置后发送通知
- 记录扫描历史用于审计

### 启动/停止监控

**通过 Web 界面**:
1. 导航到 Settings 页面
2. 使用 **Start Monitoring** / **Stop Monitoring** 按钮

**通过 API**:
```bash
# 启动
curl -X POST http://localhost:8080/api/v1/monitor/start \
  -H "Authorization: Bearer <token>"

# 停止
curl -X POST http://localhost:8080/api/v1/monitor/stop \
  -H "Authorization: Bearer <token>"
```

---

## 通知配置

### 企业微信（WeCom）
```json
{
  "name": "企业微信渠道",
  "type": "wecom",
  "webhook_url": "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=YOUR_KEY",
  "enabled": true,
  "notify_on_new": true,
  "notify_on_confirmed": true
}
```

### 钉钉（DingTalk）
```json
{
  "name": "钉钉渠道",
  "type": "dingtalk",
  "webhook_url": "https://oapi.dingtalk.com/robot/send?access_token=YOUR_TOKEN",
  "secret": "YOUR_SECRET",
  "enabled": true,
  "notify_on_new": true,
  "notify_on_confirmed": true
}
```

### 飞书（Feishu/Lark）
```json
{
  "name": "飞书渠道",
  "type": "feishu",
  "webhook_url": "https://open.feishu.cn/open-apis/bot/v2/hook/YOUR_HOOK",
  "secret": "YOUR_SECRET",
  "enabled": true,
  "notify_on_new": true,
  "notify_on_confirmed": true
}
```

### 自定义 Webhook
```json
{
  "name": "自定义 Webhook",
  "type": "webhook",
  "webhook_url": "https://your-api.com/webhook",
  "enabled": true,
  "notify_on_new": true,
  "notify_on_confirmed": true
}
```

---

## 性能

### 优化特性

- **数据库索引**: 所有频繁查询的字段都已建立索引
- **批量操作**: 多个更新使用单个 SQL 查询
- **令牌轮换**: 自动故障转移防止速率限制阻塞
- **分页**: 高效加载大型结果集
- **缓存**: 对频繁访问的数据进行战略性缓存

### 可扩展性

- 支持多个 GitHub Token 以获得更高的速率限制
- 可通过负载均衡实现水平扩展
- 数据库可独立扩展
- 无状态后端设计便于复制

---

## 贡献

欢迎贡献！请遵循以下指南：

1. Fork 仓库
2. 创建功能分支
3. 进行更改
4. 编写或更新测试
5. 提交 Pull Request

---

## 许可证

本项目采用 MIT 许可证。详见 LICENSE 文件。

---

## 更新日志

### 版本 1.2.0 (2025-11-13)
- 在前端添加 GitHub Token 管理
- 实现 JWT 认证系统
- 为搜索结果添加批量操作
- 添加代理支持（HTTP/HTTPS/SOCKS5）
- 优化 Settings 页面，采用可折叠部分
- 添加监控规则 CRUD 功能
- 改进错误处理和验证

### 版本 1.1.0
- 添加通知系统（企业微信、钉钉、飞书、Webhook）
- 实现白名单管理
- 添加设置页面
- 改进仪表板统计

### 版本 1.0.0
- 初始版本
- 核心监控功能
- 令牌轮换系统
- 基本 Web 界面

---

**最后更新**: 2025-11-13
**项目状态**: 生产就绪
**维护状态**: 是
