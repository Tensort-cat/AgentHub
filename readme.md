# AgentHub

AgentHub 是一个面向开发者的 AI Workflow 平台，目标是让用户能够通过节点编排的方式构建自定义 Agent 工作流，并结合知识库、外部工具和大模型能力完成复杂任务。

目前本仓库仍处于后端开发阶段，前端界面与完整业务闭环尚未完全落地，核心能力主要集中在后端服务、工作流执行引擎、知识库管理、工具接入和安全认证等方面。

## 项目状态

- 当前版本：后端服务为主，未开发完毕
- 启动入口：`cmd/main.go`
- 配置文件：`configs/config.toml`（参考 `configs/config_tmp.toml` 修改）
- 运行方式：在项目根目录执行启动命令即可启动后端服务

## 技术栈

- Go
- Gin
- Eino
- GORM
- MySQL
- Redis
- RabbitMQ
- RAG
- MCP
- JWT
- Zap

## 项目定位

AgentHub 旨在为开发者提供一套可扩展、可编排、可集成的 AI Workflow 平台，帮助用户在业务场景中通过图结构工作流组合 LLM、知识库、函数工具和外部服务能力，实现更复杂、更可控的任务执行。

## 核心能力

### 1. Workflow 编排

- 基于 Eino Graph 构建工作流执行引擎
- 将大模型、知识库、工具能力抽象为可组合节点
- 支持节点编排和图结构管理
- 运行前校验节点、边和图结构
- 避免环路、孤立节点、非法连接等问题导致运行时异常

### 2. 并发与异步任务处理

- 针对文档解析、切片、Embedding、向量写入等耗时操作采用异步任务机制
- 通过 RabbitMQ 解耦文件上传与知识库索引构建
- 避免同步处理阻塞 HTTP 请求
- 提升后台任务吞吐能力和系统稳定性

### 3. RAG 知识库

- 使用 Recursive Chunk 对文档进行语义切分
- 对切片内容执行 Embedding，生成向量表示
- 将向量数据存储到 Redis，利用 Redis Vector Search 完成相似度检索
- 将文档元数据持久化至 MySQL
- 采用 “MySQL 持久化 + Redis 向量检索” 的架构，提高数据可靠性和检索效率

### 4. MCP 工具体系

- 基于 MCP 接入外部工具能力
- 将工具能力统一抽象并交付给 Agent Workflow
- 结合 Function Calling 机制实现模型根据任务自主选择工具
- 提升工作流扩展性和集成能力

### 5. 服务端安全与工程化

- 使用 JWT 实现用户身份认证与接口鉴权
- 使用 Gin Middleware 统一处理认证、异常恢复和通用横切逻辑
- 使用 GORM 管理 MySQL 数据访问
- 针对高频查询设计索引与 Redis 缓存策略
- 使用 Zap 统一日志输出，提升运维和排查效率

## 项目目录说明

```text
AgentHub/
├── cmd/                 # 程序启动入口
│   └── main.go          # 运行后端服务的主入口
├── configs/             # 配置文件目录
│   ├── config_tmp.toml  # 配置模板
│   └── config.toml      # 实际修改用配置文件（未加入模板）
├── internal/            # 内部业务代码
│   ├── config/          # 配置加载
│   ├── controller/      # 控制器
│   ├── dao/             # 数据访问层
│   ├── dto/             # 请求/响应 DTO
│   ├── middleware/      # 中间件
│   ├── model/           # 数据模型
│   ├── route/           # 路由注册
│   └── service/         # 业务服务层
├── pkg/                 # 公共工具包
├── static/              # 静态资源目录
├── logs/                # 日志目录
├── docs/                # 项目文档、方案与设计说明
├── test/                # 单元测试
├── go.mod               # Go 模块文件
├── tables.sql           # 数据库初始化脚本
├── readme.md            # 项目说明文档
└── README.zh.md         # 可按需扩展中文说明（若后续补充）
```

## 快速开始

### 1. 环境准备

请确保本机已安装以下依赖：

- Go 1.25+
- MySQL
- Redis
- RabbitMQ
- 可用的大模型 API（如 OpenAI / 火山引擎等）

### 2. 配置文件

在 `configs/` 目录下复制一份配置文件：

```bash
copy configs\config_tmp.toml configs\config.toml
```

然后根据自己的环境修改 `configs/config.toml`，至少需要配置：

- MySQL 连接信息
- Redis 连接信息
- 服务端口与 host
- 日志路径
- 静态资源路径
- 邮箱配置（验证码发送使用）

> 当前项目中的配置初始化逻辑会读取固定路径：`D:\dev_soft\AgentHub\configs\config.toml`，因此请务必确保 `config.toml` 位于该路径，或者按需修改配置加载逻辑。

示例配置结构：

```toml
[mainConfig]
appName = "AgentHub"
host = "127.0.0.1"
port = 8000

[mysqlConfig]
host = "127.0.0.1"
port = 3306
user = "root"
password = "your_password"
database = "agenthub"

[redisConfig]
host = "127.0.0.1"
port = "6379"
password = ""
db = 0

[logConfig]
logPath = "./logs/app.log"

[staticSrcConfig]
staticAvatarPath = "../static/avatars"
staticFilePath = "../static/files"
```

### 3. 启动服务

在项目根目录执行：

```bash
go run ./cmd/main.go
```

或者先构建再运行：

```bash
go build -o agenthub.exe ./cmd
./agenthub.exe
```

启动成功后，服务将按配置中的端口监听，例如：

```text
http://localhost:8000
```

## 运行说明

- 后端服务入口在 `cmd/main.go`
- 程序启动时会初始化：
  - 配置文件
  - MySQL 连接
  - Redis 连接
  - 路由与 Web 服务
- 日志默认输出到 `./logs/app.log`
- 静态文件目录位于 `static/`，包括头像和上传文件

## 设计亮点

1. 面向开发者的可编排 Agent 工作流
2. 可扩展的工具系统与 MCP 集成
3. 文档知识库的高性能 RAG 方案
4. 异步任务解耦，提升系统吞吐
5. 统一认证、日志和异常处理能力
6. 适合继续扩展为完整的 AI Workflow 平台

## 计划与后续完善方向

当前项目仍处于持续迭代阶段，后续可进一步完善：

- 前端交互与工作流可视化设计器
- 节点运行时执行状态管理
- 更完整的工作流执行监控与任务追踪
- 更丰富的知识库文档解析与索引能力
- MCP 工具注册与动态加载机制
- 更完善的权限体系与多租户支持
- 高并发性能优化与监控治理

## 备注

本项目当前更偏向后端原型/基础架构实现，适合用于：

- 学习 AI Workflow + Agent 平台的工程设计
- 作为 Go 后端开发练手项目
- 作为个人作品展示或简历项目的基础版本

## License

本项目目前未声明正式 License，适用于学习、研究和个人开发场景。
