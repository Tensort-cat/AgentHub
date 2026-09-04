# AgentHub HTTP 接口文档

> 说明：本文档基于当前 `internal/route` 下的路由注册代码整理，适用于后端接口开发阶段的接口说明。所有接口前缀均为 `/api/v1`。

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

大部分接口均需要登录后访问，认证方式为：

```http
Authorization: Bearer <token>
```

兼容 URL 参数传 token：

```text
?token=<token>
```

### 1.3 通用返回格式

项目统一返回 JSON，格式如下：

```json
{
  "code": 200,
  "msg": "OK",
  "data": {}
}
```

字段说明：

- `code`: 业务状态码
- `msg`: 返回说明
- `data`: 返回数据，成功时通常是对象或数组，失败时一般为 `null`

### 1.4 状态码

当前控制器中根据返回值映射 HTTP 状态：

- `200 OK`：成功
- `400 Bad Request`：请求参数错误/业务校验失败
- `401 Unauthorized`：未登录或 JWT 无效
- `500 Internal Server Error`：服务端异常

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

说明：

- `email`：邮箱
- `password`：密码

返回示例：

```json
{
  "code": 200,
  "msg": "OK",
  "data": {
    "token": "jwt_token",
    "user_id": 1,
    "email": "user@example.com"
  }
}
```

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

说明：

- `name`：用户名
- `email`：邮箱
- `password`：密码
- `captcha`：验证码

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

---

## 3. 用户相关接口

### 3.1 获取当前登录用户信息

- Method: `GET`
- Path: `/api/v1/users/me`
- Auth: 是

返回示例：

```json
{
  "code": 200,
  "msg": "OK",
  "data": {
    "id": 1,
    "name": "Alice",
    "email": "user@example.com"
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
  "code": 200,
  "msg": "OK",
  "data": [
    {
      "id": 1,
      "name": "OpenAI GPT-4o",
      "base_url": "https://api.openai.com/v1",
      "type": "openai"
    }
  ]
}
```

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
  "type": "openai"
}
```

说明：

- `type`：模型类型，具体取值由项目枚举定义

### 4.3 更新模型

- Method: `PUT`
- Path: `/api/v1/models/:id`
- Auth: 是

请求体：

```json
{
  "name": "New Model Name",
  "base_url": "https://api.example.com/v1",
  "api_key": "new_key",
  "type": "openai"
}
```

### 4.4 删除模型

- Method: `DELETE`
- Path: `/api/v1/models/:id`
- Auth: 是

---

## 5. 工作流相关接口

### 5.1 分页查询工作流

- Method: `GET`
- Path: `/api/v1/workflows`
- Auth: 是

支持查询参数：

- `page`：页码
- `size`：每页数量

示例：

```text
GET /api/v1/workflows?page=1&size=10
```

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

### 5.3 获取工作流详情

- Method: `GET`
- Path: `/api/v1/workflows/:id`
- Auth: 是

返回示例：

```json
{
  "code": 200,
  "msg": "OK",
  "data": {
    "id": 1,
    "name": "测试工作流",
    "description": "用于测试 Agent 执行流程",
    "nodes": [],
    "edges": []
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
  "description": "更新描述",
  "status": "active"
}
```

### 5.5 删除工作流

- Method: `DELETE`
- Path: `/api/v1/workflows/:id`
- Auth: 是

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

说明：

- `id`: 工作流 ID
- `input`: 任务输入内容

### 5.7 创建节点

- Method: `POST`
- Path: `/api/v1/workflows/:id/nodes`
- Auth: 是

请求体：

```json
{
  "name": "LLM节点",
  "type": "llm",
  "position_x": 100,
  "position_y": 200,
  "config": {
    "model_id": 1,
    "prompt": "请回答问题"
  }
}
```

说明：

- `type`：节点类型
- `config`：节点配置，使用 JSON 对象

### 5.8 更新节点

- Method: `PUT`
- Path: `/api/v1/workflows/:id/nodes/:node_id`
- Auth: 是

请求体：

```json
{
  "name": "更新后的节点名",
  "type": "llm",
  "position_x": 120,
  "position_y": 260,
  "config": {
    "model_id": 1,
    "prompt": "更新后的提示词"
  }
}
```

### 5.9 删除节点

- Method: `DELETE`
- Path: `/api/v1/workflows/:id/nodes/:node_id`
- Auth: 是

### 5.10 创建边

- Method: `POST`
- Path: `/api/v1/workflows/:id/edges`
- Auth: 是

请求体：

```json
{
  "source_node_id": 1,
  "target_node_id": 2
}
```

说明：

- 从 `source_node_id` 指向 `target_node_id`

### 5.11 删除边

- Method: `DELETE`
- Path: `/api/v1/workflows/:id/edges/:edge_id`
- Auth: 是

---

## 6. 会话相关接口

### 6.1 分页查询会话

- Method: `GET`
- Path: `/api/v1/sessions`
- Auth: 是

支持查询参数：

- `page`
- `size`

### 6.2 分页获取会话消息列表

- Method: `GET`
- Path: `/api/v1/sessions/:id`
- Auth: 是

说明：

- `:id` 为会话 ID
- 返回该会话下的消息分页数据

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
  "content": "你好，这是一个测试消息",
  "type": "user"
}
```

说明：

- `session_id`：会话编号
- `content`：消息内容
- `type`：消息类型（如 `user`、`assistant` 等，取决于枚举定义）

---

## 8. 知识库相关接口

### 8.1 列出知识库

- Method: `GET`
- Path: `/api/v1/kb`
- Auth: 是

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

说明：

- `embedder_id`：嵌入模型 ID

### 8.3 获取知识库详情与文件列表

- Method: `GET`
- Path: `/api/v1/kb/:id`
- Auth: 是

返回示例：

```json
{
  "code": 200,
  "msg": "OK",
  "data": {
    "id": 1,
    "name": "技术文档库",
    "documents": []
  }
}
```

---

## 9. 文档相关接口

### 9.1 上传文档并加入指定知识库

- Method: `POST`
- Path: `/api/v1/docs`
- Auth: 是

请求体：

- 采用 `multipart/form-data`
- 需上传文件字段，字段名由前端决定，后端通过 `multipart.FileHeader` 读取

示例：

```text
POST /api/v1/docs
Content-Type: multipart/form-data

knowledge_base_id=1
name=项目说明.pdf
type=document
size=1024000
file=@/path/to/project.pdf
```

说明：

- `knowledge_base_id`：所属知识库 ID
- `name`：文档名
- `type`：文档类型
- `size`：文件大小
- `file`：上传的文件二进制内容

### 9.2 下载文档

- Method: `GET`
- Path: `/api/v1/docs/:id/download`
- Auth: 是

说明：

- 用于下载已上传文档

---

## 10. 工具中心相关接口

### 10.1 列出可用工具（卡片展示）

- Method: `GET`
- Path: `/api/v1/tools`
- Auth: 是

### 10.2 工具详情

- Method: `GET`
- Path: `/api/v1/tools/:id`
- Auth: 是

返回示例：

```json
{
  "code": 200,
  "msg": "OK",
  "data": {
    "id": 1,
    "name": "天气查询工具",
    "provider": "openapi",
    "config": {}
  }
}
```

### 10.3 启用/禁用工具

- Method: `POST`
- Path: `/api/v1/tools`
- Auth: 是

请求体：

```json
{
  "tool_id": 1,
  "status": "enabled"
}
```

说明：

- `status`：工具状态（启用/禁用）

### 10.4 用户订阅/取消订阅工具

- Method: `POST`
- Path: `/api/v1/user_tools`
- Auth: 是

请求体：

```json
{
  "tool_id": 1,
  "status": "subscribed"
}
```

说明：

- `status`：用户工具状态（订阅/取消订阅）

---

## 11. 静态资源接口

### 11.1 获取头像

- Method: `GET`
- Path: `/api/v1/static/avatars/*filepath`
- Auth: 是

说明：

- 通过 Gin `Static` 路由暴露头像资源
- 实际资源目录由 `config.toml` 中 `staticAvatarPath` 控制

例如：

```text
/api/v1/static/avatars/user_1.png
```

---

## 12. 接口调用建议

1. 先调用登录接口获取 JWT。
2. 在请求头中携带 `Authorization: Bearer <token>`。
3. 对于 POST/PUT 请求，统一使用 `application/json`，除文档上传外。
4. 需要对 `config`、`edges`、`nodes` 等结构化字段按实际图结构设计传参。
5. 对于 `multipart/form-data` 上传，确保包含文件字段和必要元数据。

---

## 13. 当前文档覆盖范围说明

当前文档已根据 `internal/route` 中已注册接口整理，覆盖范围包括：

- 认证
- 用户
- 模型
- 工作流
- 会话
- 消息
- 知识库
- 文档
- 工具
- 静态资源

如果后续路由继续补充，建议同步更新此文档。
