# AgentHub HTTP 接口文档

> 说明：本文档基于当前 `internal/route`、`internal/dto` 与 `internal/controller` 下的实现整理，适用于后端接口开发阶段。所有接口前缀均为 `/api/v1`。

## 1. 通用说明

### 1.1 基础地址

```text
http://<host>:<port>/api/v1
```

例如：

```text
http://localhost:8000/api/v1
```

### 1.2 认证方式

除认证接口外，其余接口均需登录后访问。推荐通过请求头携带登录接口返回的完整 `token`：

```http
Authorization: Bearer <jwt>
```

中间件也兼容 URL 查询参数。按照当前实现，参数值同样需要包含 `Bearer ` 前缀，并进行 URL 编码：

```text
?token=Bearer%20<jwt>
```

### 1.3 通用返回格式

除文档下载成功时直接返回文件外，接口统一返回 JSON：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

字段说明：

- `code`：业务状态码
- `msg`：返回说明；成功时通常为 `ok`
- `data`：返回数据，可以是对象、数组或 `null`

对于只表示操作结果的接口，成功返回如下：

```json
{
  "code": 0,
  "msg": "ok",
  "data": null
}
```

### 1.4 业务状态码与 HTTP 状态

业务状态码定义如下：

| `code` | 含义 |
| --- | --- |
| `0` | 成功 |
| `1001` | 认证失败，Token 无效或过期 |
| `1002` | 权限不足 |
| `2001` | 请求参数或业务数据错误 |
| `3001` | 资源不存在 |
| `4001` | 服务端内部错误 |

当前控制器已配置的 HTTP 状态映射如下：

| 业务状态码 | HTTP 状态 |
| --- | --- |
| `0` | `200 OK` |
| `1001` | `401 Unauthorized` |
| `2001` | `400 Bad Request` |
| `4001` | `500 Internal Server Error` |

> `1002` 和 `3001` 已在常量中定义，但当前通用返回函数尚未配置对应的 HTTP 状态映射。

### 1.5 枚举值

接口中的枚举字段当前使用以下值：

- 模型类型 `type`：`1` 对话模型，`2` 嵌入模型
- 模型提供商 `provider`：`1` OpenAI，`2` Ark
- 工作流状态 `status`：`1` 草稿，`2` 启用
- 工作流节点类型 `type`：`1` 对话模型、`2` 对话模板、`3` 分支、`4` 工具、`5` 检索器、`6` Agent、`7` 开始、`8` 结束
- 消息类型 `type`：`1` 用户、`2` 助手、`3` 工具、`4` 系统
- 工具状态 `status`：`1` 禁用、`2` 启用
- 用户工具状态 `status`：`1` 未订阅、`2` 已订阅
- 文档类型 `type`：`.md` 或 `.txt`
- 文档状态 `status`：`1` 等待处理、`2` 解析中、`3` 索引中、`4` 成功、`5` 失败

---

## 2. 认证相关接口

### 2.1 用户登录

- Method: `POST`
- Path: `/api/v1/auth/login`
- Auth: 否

请求体：

```json
{
  "email": "user@example.com",
  "password": "123456"
}
```

- `email`：必填，邮箱
- `password`：必填，密码

返回示例：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "token": "Bearer <jwt>",
    "id": 1,
    "name": "Alice",
    "email": "user@example.com",
    "avatar": "/api/v1/static/avatars/user_1.png"
  }
}
```

登录后，后续请求都应该在 `Header` 中的 `Authorization` 携带 `token`，前端可以考虑存放在 Local Storage 或 Session Storage 中

### 2.2 用户注册

- Method: `POST`
- Path: `/api/v1/auth/register`
- Auth: 否

请求体：

```json
{
  "name": "Alice",
  "email": "user@example.com",
  "password": "123456",
  "captcha": "123456"
}
```

- `name`：用户名
- `email`：邮箱
- `password`：密码
- `captcha`：通过发送验证码接口获取的邮箱验证码

成功时 `data` 为 `null`。

### 2.3 发送验证码

- Method: `POST`
- Path: `/api/v1/auth/sendCaptcha`
- Auth: 否

请求体：

```json
{
  "email": "user@example.com"
}
```

成功时 `data` 为 `null`。

---

## 3. 用户相关接口

### 3.1 获取当前登录用户信息

- Method: `GET`
- Path: `/api/v1/users/me`
- Auth: 是

返回示例：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "name": "Alice",
    "email": "user@example.com",
    "avatar": "/api/v1/static/avatars/user_1.png"
  }
}
```

---

## 4. 模型相关接口

### 4.1 获取当前用户模型列表

- Method: `GET`
- Path: `/api/v1/models`
- Auth: 是

返回示例：

```json
{
  "code": 0,
  "msg": "ok",
  "data": [
    {
      "id": 1,
      "name": "OpenAI GPT-4o",
      "provider": 1,
      "base_url": "https://api.openai.com/v1",
      "type": 1,
      "created_at": "2026-09-24T10:00:00+08:00"
    }
  ]
}
```

> 模型列表不会返回 `api_key`。

### 4.2 创建模型

- Method: `POST`
- Path: `/api/v1/models`
- Auth: 是

请求体：

```json
{
  "name": "OpenAI GPT-4o",
  "base_url": "https://api.openai.com/v1",
  "api_key": "sk-xxxx",
  "type": 1
}
```

- `name`、`base_url`、`api_key`：不能为空
- `type`：必填，`1` 为对话模型，`2` 为嵌入模型

成功时 `data` 为 `null`。

### 4.3 更新模型

- Method: `PUT`
- Path: `/api/v1/models/:id`
- Auth: 是

请求体：

```json
{
  "name": "New Model Name",
  "base_url": "https://api.example.com/v1",
  "api_key": "new_key"
}
```

- `name`、`base_url`、`api_key` 均为可选字段，只更新非空字段
- 当前控制器不支持通过该接口修改模型类型

成功时 `data` 为 `null`。

### 4.4 删除模型

- Method: `DELETE`
- Path: `/api/v1/models/:id`
- Auth: 是

成功时 `data` 为 `null`。

---

## 5. 工作流相关接口

### 5.1 分页查询工作流

- Method: `GET`
- Path: `/api/v1/workflows`
- Auth: 是

查询参数：

- `page`：必填，页码，从 `1` 开始
- `size`：必填，每页数量，必须大于 `0`

```text
GET /api/v1/workflows?page=1&size=10
```

返回示例：

```json
{
  "code": 0,
  "msg": "ok",
  "data": [
    {
      "id": 1,
      "name": "测试工作流",
      "description": "用于测试 Agent 执行流程",
      "status": 1,
      "created_at": "2026-09-24T10:00:00+08:00"
    }
  ]
}
```

> 当前分页响应仅返回本页数组，不包含总数、总页数等元数据。

### 5.2 创建工作流

- Method: `POST`
- Path: `/api/v1/workflows`
- Auth: 是

请求体：

```json
{
  "name": "测试工作流",
  "description": "用于测试 Agent 执行流程"
}
```

新建工作流状态默认为草稿。成功时 `data` 为 `null`。

### 5.3 获取工作流详情

- Method: `GET`
- Path: `/api/v1/workflows/:id`
- Auth: 是

返回示例：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "name": "测试工作流",
    "description": "用于测试 Agent 执行流程",
    "nodes": [
      {
        "id": 11,
        "name": "开始",
        "type": 7,
        "position_x": 100,
        "position_y": 200,
        "config": {}
      }
    ],
    "edges": [
      {
        "id": 21,
        "source_node_id": 11,
        "target_node_id": 12,
        "config": {}
      }
    ]
  }
}
```

### 5.4 更新工作流元信息

- Method: `PUT`
- Path: `/api/v1/workflows/:id`
- Auth: 是

请求体：

```json
{
  "name": "更新后的工作流名",
  "description": "更新后的描述",
  "status": 2
}
```

- 三个字段均为可选字段，但至少需要传入一个可更新字段
- `status`：`1` 为草稿，`2` 为启用

成功时 `data` 为 `null`。

### 5.5 删除工作流

- Method: `DELETE`
- Path: `/api/v1/workflows/:id`
- Auth: 是

成功时 `data` 为 `null`。

### 5.6 运行工作流

- Method: `POST`
- Path: `/api/v1/workflows/run`
- Auth: 是

请求体：

```json
{
  "id": 1,
  "input": "请总结这份文档内容"
}
```

- `id`：工作流 ID
- `input`：本次任务的输入内容
- 当前接口只负责校验并将任务提交到 RabbitMQ，不会在 HTTP 请求内等待工作流执行完成

受理成功返回示例：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "task_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "queued"
  }
}
```

> 当前受理成功的 HTTP 状态仍为 `200 OK`。

### 5.7 创建节点

- Method: `POST`
- Path: `/api/v1/workflows/:id/nodes`
- Auth: 是

请求体：

```json
{
  "name": "LLM 节点",
  "type": 1,
  "position_x": 100,
  "position_y": 200,
  "config": {
    "model_id": 1,
    "prompt": "请回答问题"
  }
}
```

- `:id`：工作流 ID
- `name`、`type`、`position_x`、`position_y`：创建节点所需字段
- `config`：节点配置 JSON，具体内容由节点类型决定

成功时 `data` 为 `null`。

### 5.8 更新节点

- Method: `PUT`
- Path: `/api/v1/workflows/:id/nodes/:node_id`
- Auth: 是

请求体：

```json
{
  "name": "更新后的节点名",
  "type": 1,
  "position_x": 120,
  "position_y": 260,
  "config": {
    "model_id": 1,
    "prompt": "更新后的提示词"
  }
}
```

以上字段均可按需传入，至少需要传入一个可更新字段。成功时 `data` 为 `null`。

### 5.9 删除节点

- Method: `DELETE`
- Path: `/api/v1/workflows/:id/nodes/:node_id`
- Auth: 是

成功时 `data` 为 `null`。

### 5.10 创建边

- Method: `POST`
- Path: `/api/v1/workflows/:id/edges`
- Auth: 是

请求体：

```json
{
  "source_node_id": 1,
  "target_node_id": 2,
  "config": {}
}
```

- `source_node_id`：必填，起点节点 ID
- `target_node_id`：必填，终点节点 ID
- `config`：可选，边配置 JSON

成功时 `data` 为 `null`。

### 5.11 删除边

- Method: `DELETE`
- Path: `/api/v1/workflows/:id/edges/:edge_id`
- Auth: 是

成功时 `data` 为 `null`。

---

## 6. 会话相关接口

### 6.1 分页查询工作流会话

- Method: `GET`
- Path: `/api/v1/sessions`
- Auth: 是

查询参数：

- `wf_id`：必填，工作流 ID
- `page`：必填，页码，从 `1` 开始
- `size`：必填，每页数量，必须大于 `0`

```text
GET /api/v1/sessions?wf_id=1&page=1&size=10
```

返回示例：

```json
{
  "code": 0,
  "msg": "ok",
  "data": [
    {
      "id": 1,
      "title": "请总结这份文档",
      "created_at": "2026-09-24T10:00:00+08:00"
    }
  ]
}
```

### 6.2 分页获取会话消息

- Method: `GET`
- Path: `/api/v1/sessions/:id`
- Auth: 是

查询参数：

- `page`：必填，页码，从 `1` 开始
- `size`：必填，每页数量，必须大于 `0`

返回示例：

```json
{
  "code": 0,
  "msg": "ok",
  "data": [
    {
      "id": 1,
      "content": "你好",
      "type": 1,
      "created_at": "2026-09-24T10:00:00+08:00"
    }
  ]
}
```

> 当前分页响应仅返回本页数组，消息按创建时间倒序排列。

---

## 7. 消息相关接口

### 7.1 创建新消息

- Method: `POST`
- Path: `/api/v1/messages`
- Auth: 是

请求体：

```json
{
  "session_id": 1,
  "content": "你好，这是一条测试消息",
  "type": 1
}
```

- `session_id`：必填，会话 ID
- `content`：必填，消息内容
- `type`：必填，消息类型，取值见枚举说明

成功时 `data` 为 `null`。

---

## 8. 知识库相关接口

### 8.1 获取知识库列表

- Method: `GET`
- Path: `/api/v1/kb`
- Auth: 是

返回示例：

```json
{
  "code": 0,
  "msg": "ok",
  "data": [
    {
      "id": 1,
      "name": "技术文档库",
      "description": "存放项目设计文档和规范",
      "created_at": "2026-09-24T10:00:00+08:00"
    }
  ]
}
```

### 8.2 创建知识库

- Method: `POST`
- Path: `/api/v1/kb`
- Auth: 是

请求体：

```json
{
  "name": "技术文档库",
  "description": "存放项目设计文档和规范",
  "embedder_id": 1
}
```

- `name`：必填，知识库名称
- `description`：可选，知识库描述
- `embedder_id`：必填，嵌入模型 ID

成功时 `data` 为 `null`。

### 8.3 获取知识库详情与文档列表

- Method: `GET`
- Path: `/api/v1/kb/:id`
- Auth: 是

返回示例：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "name": "技术文档库",
    "description": "存放项目设计文档和规范",
    "embedder_id": 2,
    "created_at": "2026-09-24T10:00:00+08:00",
    "docs": [
      {
        "id": 11,
        "name": "README.md",
        "size": 1024,
        "status": 4,
        "type": ".md",
        "created_at": "2026-09-24T10:05:00+08:00"
      }
    ]
  }
}
```

---

## 9. 文档相关接口

### 9.1 上传文档并加入指定知识库

- Method: `POST`
- Path: `/api/v1/docs`
- Auth: 是
- Content-Type: `multipart/form-data`

表单字段：

- `knowledge_base_id`：必填，所属知识库 ID
- `file`：必填，上传文件；当前文档类型枚举为 `.md` 和 `.txt`

```text
POST /api/v1/docs
Content-Type: multipart/form-data

knowledge_base_id=1
file=@/path/to/README.md
```

> 文件名、类型和大小由服务端从上传文件中读取，不需要额外提交 `name`、`type` 或 `size` 字段。

返回示例：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 11,
    "name": "README.md",
    "type": ".md",
    "size": 1024,
    "status": 4
  }
}
```

### 9.2 下载文档

- Method: `GET`
- Path: `/api/v1/docs/:id/download`
- Auth: 是

成功时直接返回文件附件，不使用通用 JSON 返回格式。

> 当前路由参数名为 `id`，但控制器读取的是 `docID`。在两者统一之前，该接口无法按路由中的文档 ID 正常查询文件。

---

## 10. 工具中心相关接口

### 10.1 获取可用工具列表

- Method: `GET`
- Path: `/api/v1/tools`
- Auth: 是

返回示例：

```json
{
  "code": 0,
  "msg": "ok",
  "data": [
    {
      "id": 1,
      "name": "天气查询工具",
      "description": "查询指定城市的天气"
    }
  ]
}
```

### 10.2 获取工具详情

- Method: `GET`
- Path: `/api/v1/tools/:id`
- Auth: 是

返回示例：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "name": "天气查询工具",
    "description": "查询指定城市的天气",
    "status": 2,
    "avatar": "/api/v1/static/avatars/tool_1.png",
    "created_at": "2026-09-24T10:00:00+08:00"
  }
}
```

### 10.3 启用或禁用工具

- Method: `POST`
- Path: `/api/v1/tools`
- Auth: 是

请求体：

```json
{
  "tool_id": 1,
  "status": 2
}
```

- `tool_id`：工具 ID
- `status`：`1` 为禁用，`2` 为启用

成功时 `data` 为 `null`。

### 10.4 用户订阅或取消订阅工具

- Method: `POST`
- Path: `/api/v1/user_tools`
- Auth: 是

请求体：

```json
{
  "tool_id": 1,
  "status": 2
}
```

- `tool_id`：工具 ID
- `status`：`1` 为未订阅，`2` 为已订阅

成功时 `data` 为 `null`。

---

## 11. 静态资源接口

### 11.1 获取头像

- Method: `GET`
- Path: `/api/v1/static/avatars/*filepath`
- Auth: 是

- 通过 Gin 静态资源路由读取头像
- 实际资源目录由 `config.toml` 中的 `staticSrcConfig.staticAvatarPath` 控制

例如：

```text
/api/v1/static/avatars/user_1.png
```

---

## 12. 接口调用建议

1. 先调用登录接口获取包含 `Bearer ` 前缀的 Token。
2. 后续请求通过 `Authorization` 请求头携带完整 Token。
3. 除文档上传外，POST 和 PUT 请求统一使用 `application/json`。
4. 工作流运行接口是异步受理接口；客户端应保存返回的 `task_id`，但当前 HTTP 路由尚未提供任务结果查询接口。
5. `config` 字段均为 JSON，具体结构需与对应节点或边类型匹配。

---

## 13. 当前文档覆盖范围

本文档覆盖当前 `internal/route` 中已经注册的接口：

- 认证
- 用户
- 模型
- 工作流、节点与边
- 会话
- 消息
- 知识库
- 文档
- 工具与用户工具
- 静态资源

`internal/dto/request/session_create.go` 中虽然定义了创建会话请求结构，但当前未注册对应的 HTTP 路由，因此不列为可用接口。
