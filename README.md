# SLS Migrate

SLS Alert 迁移系统的微服务架构实现，使用 Go + Gin + GORM + MySQL 构建。

## 功能特性

- 🚀 完整的 Alert CRUD 操作
- 🗄️ 分表设计，支持复杂 Alert 结构
- 🔄 事务支持，保证数据一致性
- ☁️ 阿里云 SLS 规则获取和同步
- 🔄 双向同步：SLS ↔ 数据库
- 📊 同步状态监控和统计
- 📚 Swagger API 文档自动生成
- 🧪 Postman 测试用例集合
- 🏗️ MVP 架构，清晰的分层设计

## 技术栈

- **语言**: Go 1.24.4
- **Web 框架**: Gin
- **ORM**: GORM
- **数据库**: MySQL
- **API 文档**: Swagger
- **配置管理**: 环境变量

## 项目结构

```
sls-migrate/
├── internal/                 # 内部包
│   ├── config/              # 配置管理
│   ├── handler/             # HTTP 处理器
│   ├── models/              # 数据模型
│   ├── service/             # 业务逻辑层
│   └── store/               # 数据存储层
├── pkg/                     # 公共包
│   ├── database/            # 数据库连接
│   └── utils/               # 工具函数
├── api/                     # API 定义
├── docs/                    # Swagger 文档
├── sql/                     # 数据库 Schema
├── main.go                  # 主程序入口
├── go.mod                   # Go 模块文件
└── README.md                # 项目说明
```

## 数据库设计

### 分表策略

考虑到 Alert 结构的复杂性，采用以下分表策略：

1. **主表**: `alerts` - 存储基本信息和关联ID
2. **配置表**: `alert_configurations` - 存储 AlertConfiguration 的复杂配置
3. **调度表**: `alert_schedules` - 存储 Schedule 信息
4. **标签表**: `alert_tags` - 存储 annotations 和 labels
5. **查询表**: `alert_queries` - 存储 queryList
6. **配置子表**: 条件配置、分组配置、策略配置、模板配置、严重程度配置

### 关联关系

- 使用外键约束保证数据一致性
- 支持级联删除
- 优化查询性能的复合索引

## 快速开始

### 环境要求

- Go 1.24.4+
- MySQL 8.0+
- Git

### 安装依赖

```bash
go mod tidy
```

### 环境配置

复制环境变量配置文件：

```bash
cp env.example .env
```

编辑 `.env` 文件，配置数据库连接信息：

```bash
# 数据库配置
DB_HOST=localhost
DB_PORT=3306
DB_USERNAME=root
DB_PASSWORD=your_password
DB_DATABASE=sls_migrate

# 阿里云 SLS 配置
SLS_ENDPOINT=cn-qingdao.log.aliyuncs.com
SLS_ACCESS_KEY_ID=your_access_key_id
SLS_ACCESS_KEY_SECRET=your_access_key_secret
SLS_PROJECT=your_project_name
SLS_LOG_STORE=your_log_store_name
```

### 数据库初始化

```bash
# 创建数据库和表结构
mysql -u root -p < sql/schema.sql
```

### 运行服务

```bash
go run main.go
```

服务将在 `http://localhost:8080` 启动。

## API 接口

### 基础接口

- `GET /health` - 健康检查
- `GET /swagger/*` - Swagger API 文档

### Alert 管理接口

- `POST /api/v1/alerts` - 创建 Alert
- `GET /api/v1/alerts` - 获取 Alert 列表
- `GET /api/v1/alerts/{id}` - 根据 ID 获取 Alert
- `GET /api/v1/alerts/name/{name}` - 根据名称获取 Alert
- `PUT /api/v1/alerts/{id}` - 更新 Alert
- `DELETE /api/v1/alerts/{id}` - 删除 Alert
- `GET /api/v1/alerts/status/{status}` - 根据状态获取 Alert 列表

### 阿里云 SLS 接口

- `GET /api/v1/sls/alerts` - 从 SLS 获取所有 Alert 规则
- `GET /api/v1/sls/alerts/name/{name}` - 从 SLS 根据名称获取 Alert 规则
- `POST /api/v1/sls/sync` - 同步 SLS Alert 规则到本地数据库
- `POST /api/v1/sls/sync/db-to-sls` - 同步本地数据库 Alert 规则到 SLS
- `GET /api/v1/sls/sync/status` - 获取同步状态和统计信息
- `GET /api/v1/sls/status` - 获取 SLS 连接状态

## 测试

### Postman 测试

1. 导入 `postman_collection.json` 到 Postman
2. 设置环境变量 `base_url` 为 `http://localhost:8080`
3. 运行测试用例

### 测试用例说明

- **Health Check**: 验证服务健康状态
- **Create Alert**: 创建完整的 Alert 记录
- **CRUD 操作**: 完整的增删改查测试
- **分页查询**: 测试列表分页功能
- **状态过滤**: 测试按状态筛选功能

## 开发指南

### 添加新的 Alert 字段

1. 在 `internal/models/alert.go` 中添加字段定义
2. 更新数据库 Schema (`sql/schema.sql`)
3. 更新 Store 层的 CRUD 操作
4. 更新 Service 层的业务逻辑
5. 更新 Handler 层的 API 接口
6. 更新 Swagger 注释

### 数据库迁移

```bash
# 自动迁移（开发环境）
go run main.go

# 手动执行 SQL（生产环境）
mysql -u root -p sls_migrate < sql/schema.sql
```

## 部署

### Docker 部署

```dockerfile
FROM golang:1.24.4-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
```

### 环境变量配置

生产环境建议使用环境变量或配置文件：

```bash
export SERVER_PORT=8080
export GIN_MODE=release
export DB_HOST=your_db_host
export DB_PORT=3306
export DB_USERNAME=your_username
export DB_PASSWORD=your_password
export DB_DATABASE=sls_migrate
```

## 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 许可证

本项目采用 Apache 2.0 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 联系方式

- 项目维护者: [Your Name]
- 邮箱: [your.email@example.com]
- 项目地址: [https://github.com/Ghostbaby/sls-migrate](https://github.com/Ghostbaby/sls-migrate)

## 更新日志

### v1.0.0 (2024-12-19)
- ✨ 初始版本发布
- 🚀 完整的 Alert CRUD 功能
- 🗄️ 分表数据库设计
- 📚 Swagger API 文档
- 🧪 Postman 测试用例