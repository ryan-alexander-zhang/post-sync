# Backend API 文档

本文档描述当前后端实际暴露的全部接口，基于 `backend/internal/api` 与对应 service/driver 实现整理。

## 1. 基本约定

- API 基础前缀：`/api/v1`
- 内容类型：
  - JSON 接口使用 `Content-Type: application/json`
  - 文件上传接口使用 `multipart/form-data`
- 时间字段统一为服务端返回的时间戳字符串，字段名采用 camelCase，例如 `createdAt`
- ID 字段均为字符串
- 当前接口未实现认证与权限控制
- 发布任务采用“同步创建，异步执行”模型：`POST /api/v1/publish-jobs` 返回成功后，真正发送在后台进行

## 2. 通用错误响应

除 `204 No Content` 外，失败响应统一为：

```json
{
  "code": "VALIDATION_ERROR",
  "message": "validation error: xxx"
}
```

通用错误码：

| code | HTTP 状态码 | 说明 |
|---|---:|---|
| `INVALID_JSON` | 400 | 请求体不是合法 JSON |
| `INVALID_FILE` | 400 | 上传文件缺失、无法打开或无法读取 |
| `VALIDATION_ERROR` | 400 | 参数校验失败、业务校验失败 |
| `CONFIG_ERROR` | 400 | 渠道配置错误 |
| `CONTENT_NOT_FOUND` | 404 | 内容不存在 |
| `CHANNEL_ACCOUNT_NOT_FOUND` | 404 | 渠道账号不存在 |
| `CHANNEL_TARGET_NOT_FOUND` | 404 | 渠道目标不存在 |
| `PUBLISH_JOB_NOT_FOUND` | 404 | 发布任务不存在 |
| `DELIVERY_TASK_NOT_FOUND` | 404 | 投递任务不存在 |
| `RESOURCE_NOT_FOUND` | 404 | 发布时引用的 content 或 target 不存在 |
| `INTERNAL_ERROR` | 500 | 未归类的服务端内部错误 |

说明：

- `VALIDATION_ERROR.message` 会直接带上具体失败原因，适合在前端直接展示
- 发送阶段的渠道错误不会直接出现在创建发布任务接口中，而会记录在 `delivery.errorCode` 和 `delivery.errorMessage`

## 3. 数据模型

### 3.1 Content

```json
{
  "id": "01abc...",
  "sourceFilename": "example.md",
  "originalMarkdown": "---\ntitle: Demo\n---\nbody",
  "frontmatterJson": "{\"title\":\"Demo\"}",
  "title": "Demo",
  "bodyMarkdown": "body",
  "bodyPlain": "body",
  "bodyHash": "sha256...",
  "createdAt": "2026-03-08T10:00:00Z"
}
```

### 3.2 ChannelAccount

```json
{
  "id": "01abc...",
  "channelType": "telegram",
  "name": "Telegram Bot",
  "enabled": true,
  "secretRef": "TELEGRAM_BOT_TOKEN",
  "configJson": "{}",
  "createdAt": "2026-03-08T10:00:00Z",
  "updatedAt": "2026-03-08T10:00:00Z"
}
```

### 3.3 ChannelTarget

```json
{
  "id": "01abc...",
  "channelAccountId": "01account...",
  "targetType": "telegram_topic",
  "targetKey": "-100123456:topic:8",
  "targetName": "Release Notes",
  "enabled": true,
  "configJson": "{\"chatId\":\"-100123456\",\"topicId\":8,\"topicName\":\"Release Notes\"}",
  "createdAt": "2026-03-08T10:00:00Z",
  "updatedAt": "2026-03-08T10:00:00Z"
}
```

### 3.4 PublishJob

```json
{
  "id": "01job...",
  "contentId": "01content...",
  "requestId": "01req...",
  "triggerSource": "manual",
  "status": "PROCESSING",
  "totalDeliveries": 2,
  "successCount": 1,
  "failedCount": 0,
  "skippedCount": 1,
  "createdAt": "2026-03-08T10:00:00Z",
  "startedAt": "2026-03-08T10:00:01Z",
  "finishedAt": null
}
```

### 3.5 DeliveryTask

```json
{
  "id": "01delivery...",
  "publishJobId": "01job...",
  "contentId": "01content...",
  "channelAccountId": "01account...",
  "channelTargetId": "01target...",
  "channelType": "feishu",
  "targetKey": "oc_xxx",
  "status": "FAILED",
  "attemptCount": 1,
  "idempotencyKey": "sha256...",
  "bodyHash": "sha256...",
  "templateName": "default",
  "renderMode": "feishu_post",
  "renderedTitle": "Demo",
  "renderedBody": "body",
  "externalMessageId": "",
  "errorCode": "CHANNEL_SEND_FAILED",
  "errorMessage": "feishu send failed: code=99991663 msg=app not found",
  "providerResponseJson": "{\"code\":99991663}",
  "createdAt": "2026-03-08T10:00:00Z",
  "startedAt": "2026-03-08T10:00:01Z",
  "finishedAt": "2026-03-08T10:00:02Z"
}
```

状态取值：

- `PublishJob.status`：`PENDING`、`PROCESSING`、`SUCCESS`、`PARTIAL_SUCCESS`、`FAILED`
- `DeliveryTask.status`：`PENDING`、`PROCESSING`、`SUCCESS`、`FAILED`、`SKIPPED_DUPLICATE`

## 4. 系统接口

### 4.1 健康检查

`GET /healthz`

用途：容器或反向代理健康检查。

成功响应：

```json
{
  "status": "ok"
}
```

### 4.2 系统信息

`GET /api/v1/system/info`

成功响应：

```json
{
  "database": "sqlite",
  "status": "ready"
}
```

字段说明：

- `database`：当前 Gorm 连接的数据库驱动名
- `status`：固定返回 `ready`

## 5. Content 接口

### 5.1 上传内容

`POST /api/v1/contents/upload`

请求类型：`multipart/form-data`

表单字段：

| 字段 | 必填 | 说明 |
|---|---|---|
| `file` | 是 | Markdown 文件 |

成功响应：`201 Created`

返回体：`Content`

业务规则：

- 文件名不能为空
- 文件内容不能为空
- 服务端会解析 frontmatter、正文、纯文本和 `bodyHash`
- 若标准化后的正文已存在，会返回 `400 VALIDATION_ERROR`

常见失败：

- `400 INVALID_FILE`：未上传文件、文件打不开、文件读失败
- `400 VALIDATION_ERROR`：Markdown 解析失败、文件为空、正文重复

示例：

```bash
curl -X POST http://localhost:8080/api/v1/contents/upload \
  -F "file=@./demo.md"
```

### 5.2 获取内容列表

`GET /api/v1/contents`

成功响应：`200 OK`

```json
{
  "items": [
    {
      "id": "01content...",
      "sourceFilename": "demo.md",
      "originalMarkdown": "....",
      "frontmatterJson": "{\"title\":\"Demo\"}",
      "title": "Demo",
      "bodyMarkdown": "body",
      "bodyPlain": "body",
      "bodyHash": "sha256...",
      "createdAt": "2026-03-08T10:00:00Z"
    }
  ]
}
```

### 5.3 获取内容详情

`GET /api/v1/contents/:id`

路径参数：

| 参数 | 说明 |
|---|---|
| `id` | 内容 ID |

成功响应：`200 OK`

返回体：`Content`

失败响应：

- `404 CONTENT_NOT_FOUND`

### 5.4 删除内容

`DELETE /api/v1/contents/:id`

路径参数：

| 参数 | 说明 |
|---|---|
| `id` | 内容 ID |

成功响应：`204 No Content`

业务规则：

- 已有关联发布历史的内容不能删除

失败响应：

- `404 CONTENT_NOT_FOUND`
- `400 VALIDATION_ERROR`，例如 `content with publish history cannot be deleted`

## 6. Channel Account 接口

### 6.1 获取渠道账号列表

`GET /api/v1/channel-accounts`

成功响应：`200 OK`

```json
{
  "items": [
    {
      "id": "01account...",
      "channelType": "personal_feishu",
      "name": "Personal Feishu Bot",
      "enabled": true,
      "secretRef": "PERSONAL_FEISHU_WEBHOOK_URL",
      "configJson": "{\"signSecretRef\":\"PERSONAL_FEISHU_SIGN_SECRET\"}",
      "createdAt": "2026-03-08T10:00:00Z",
      "updatedAt": "2026-03-08T10:00:00Z"
    }
  ]
}
```

安全规则：

- `personal_feishu` 账号列表会清理 `configJson` 里的 `webhookUrl` 和 `signSecret`
- 当前实现不会隐藏 `secretRef`

### 6.2 创建渠道账号

`POST /api/v1/channel-accounts`

请求体：

```json
{
  "channelType": "telegram",
  "name": "Telegram Bot",
  "enabled": true,
  "secretRef": "TELEGRAM_BOT_TOKEN",
  "config": {}
}
```

字段说明：

| 字段 | 必填 | 说明 |
|---|---|---|
| `channelType` | 是 | `telegram`、`feishu`、`personal_feishu` |
| `name` | 是 | 展示名称 |
| `enabled` | 否 | 默认 `true` |
| `secretRef` | 视渠道而定 | 敏感信息对应的环境变量名 |
| `config` | 否 | 渠道专属配置 |

成功响应：`201 Created`

返回体：`ChannelAccount`

失败响应：

- `400 INVALID_JSON`
- `400 VALIDATION_ERROR`

### 6.3 更新渠道账号

`PATCH /api/v1/channel-accounts/:id`

路径参数：

| 参数 | 说明 |
|---|---|
| `id` | 渠道账号 ID |

请求体为局部更新，所有字段可选：

```json
{
  "name": "Telegram Main Bot",
  "enabled": false,
  "secretRef": "TELEGRAM_BOT_TOKEN",
  "config": {}
}
```

成功响应：`200 OK`

返回体：`ChannelAccount`

失败响应：

- `400 INVALID_JSON`
- `400 VALIDATION_ERROR`
- `404 CHANNEL_ACCOUNT_NOT_FOUND`

说明：

- `config` 是整体覆盖，不是深度 merge
- 更新后会再次按渠道驱动做账号校验

### 6.4 删除渠道账号

`DELETE /api/v1/channel-accounts/:id`

成功响应：`204 No Content`

失败响应：

- `404 CHANNEL_ACCOUNT_NOT_FOUND`

## 7. Channel Target 接口

### 7.1 获取渠道目标列表

`GET /api/v1/channel-targets`

成功响应：`200 OK`

```json
{
  "items": [
    {
      "id": "01target...",
      "channelAccountId": "01account...",
      "targetType": "feishu_chat",
      "targetKey": "oc_xxx",
      "targetName": "Dev Group",
      "enabled": true,
      "configJson": "{\"receiveIdType\":\"chat_id\",\"chatId\":\"oc_xxx\"}",
      "createdAt": "2026-03-08T10:00:00Z",
      "updatedAt": "2026-03-08T10:00:00Z"
    }
  ]
}
```

安全规则：

- `personal_feishu_webhook` target 列表会清理 `configJson` 里的 `webhookUrl`

### 7.2 创建渠道目标

`POST /api/v1/channel-targets`

请求体示例：Telegram 群组

```json
{
  "channelAccountId": "01account...",
  "targetType": "telegram_group",
  "targetKey": "-1001234567890",
  "targetName": "Main Group",
  "enabled": true,
  "config": {
    "chatId": "-1001234567890",
    "disableNotification": false,
    "disableWebPagePreview": false
  }
}
```

请求体示例：Telegram Topic

```json
{
  "channelAccountId": "01account...",
  "targetType": "telegram_topic",
  "targetKey": "-1001234567890",
  "targetName": "Release Notes",
  "config": {
    "chatId": "-1001234567890",
    "topicId": 8,
    "topicName": "Release Notes"
  }
}
```

请求体示例：Enterprise Feishu

```json
{
  "channelAccountId": "01account...",
  "targetType": "feishu_chat",
  "targetKey": "oc_xxx",
  "targetName": "Dev Group",
  "config": {
    "receiveIdType": "chat_id",
    "chatId": "oc_xxx"
  }
}
```

请求体示例：Personal Feishu

```json
{
  "channelAccountId": "01account...",
  "targetType": "personal_feishu_webhook",
  "targetKey": "",
  "targetName": "Personal Bot",
  "config": {}
}
```

字段说明：

| 字段 | 必填 | 说明 |
|---|---|---|
| `channelAccountId` | 是 | 所属渠道账号 ID |
| `targetType` | 否但强烈建议传 | 目标类型，默认值由驱动决定 |
| `targetKey` | 视渠道而定 | 逻辑目标键 |
| `targetName` | 是 | 展示名称 |
| `enabled` | 否 | 默认 `true` |
| `config` | 否 | 渠道专属目标配置 |

成功响应：`201 Created`

返回体：`ChannelTarget`

失败响应：

- `400 INVALID_JSON`
- `400 VALIDATION_ERROR`
- `404 CHANNEL_ACCOUNT_NOT_FOUND`

### 7.3 更新渠道目标

`PATCH /api/v1/channel-targets/:id`

请求体为局部更新：

```json
{
  "targetName": "Release Notes",
  "enabled": true,
  "config": {
    "chatId": "-1001234567890",
    "topicId": 9,
    "topicName": "Release Notes V2"
  }
}
```

成功响应：`200 OK`

返回体：`ChannelTarget`

失败响应：

- `400 INVALID_JSON`
- `400 VALIDATION_ERROR`
- `404 CHANNEL_TARGET_NOT_FOUND`

说明：

- `config` 也是整体覆盖
- 更新后后端会重新做 target 规范化，可能导致 `targetType` 或 `targetKey` 被重算

### 7.4 删除渠道目标

`DELETE /api/v1/channel-targets/:id`

成功响应：`204 No Content`

失败响应：

- `404 CHANNEL_TARGET_NOT_FOUND`

## 8. Publish 接口

### 8.1 创建发布任务

`POST /api/v1/publish-jobs`

请求体：

```json
{
  "contentId": "01content...",
  "targetIds": ["01targetA...", "01targetB..."],
  "templateName": "default"
}
```

字段说明：

| 字段 | 必填 | 说明 |
|---|---|---|
| `contentId` | 是 | 要发布的内容 ID |
| `targetIds` | 是 | 目标 ID 数组，不能为空 |
| `templateName` | 否 | 不传或空字符串时自动使用 `default` |

成功响应：`201 Created`

```json
{
  "jobId": "01job...",
  "status": "PENDING"
}
```

业务规则：

- 接口只负责创建 `PublishJob` 和 `DeliveryTask`
- 创建完成后后台异步执行
- `DeliveryTask.idempotencyKey = sha256(channel_type + \":\" + target_key + \":\" + body_hash)`
- 若历史上同一 `channelType + targetKey + bodyHash` 已成功发送，则该投递会在执行阶段变成 `SKIPPED_DUPLICATE`

失败响应：

- `400 INVALID_JSON`
- `404 RESOURCE_NOT_FOUND`
- `400 VALIDATION_ERROR`

### 8.2 获取发布任务列表

`GET /api/v1/publish-jobs`

成功响应：`200 OK`

```json
{
  "items": [
    {
      "id": "01job...",
      "contentId": "01content...",
      "requestId": "01req...",
      "triggerSource": "manual",
      "status": "SUCCESS",
      "totalDeliveries": 2,
      "successCount": 1,
      "failedCount": 0,
      "skippedCount": 1,
      "createdAt": "2026-03-08T10:00:00Z",
      "startedAt": "2026-03-08T10:00:01Z",
      "finishedAt": "2026-03-08T10:00:02Z"
    }
  ]
}
```

### 8.3 获取发布任务详情

`GET /api/v1/publish-jobs/:id`

成功响应：`200 OK`

```json
{
  "job": {
    "id": "01job...",
    "contentId": "01content...",
    "requestId": "01req...",
    "triggerSource": "manual",
    "status": "PARTIAL_SUCCESS",
    "totalDeliveries": 2,
    "successCount": 1,
    "failedCount": 1,
    "skippedCount": 0,
    "createdAt": "2026-03-08T10:00:00Z",
    "startedAt": "2026-03-08T10:00:01Z",
    "finishedAt": "2026-03-08T10:00:03Z"
  },
  "deliveries": [
    {
      "id": "01delivery...",
      "publishJobId": "01job...",
      "contentId": "01content...",
      "channelAccountId": "01account...",
      "channelTargetId": "01target...",
      "channelType": "telegram",
      "targetKey": "-1001234567890",
      "status": "SUCCESS",
      "attemptCount": 1,
      "idempotencyKey": "sha256...",
      "bodyHash": "sha256...",
      "templateName": "default",
      "renderMode": "telegram_html",
      "renderedTitle": "Demo",
      "renderedBody": "<b>hello</b>",
      "externalMessageId": "321",
      "errorCode": "",
      "errorMessage": "",
      "providerResponseJson": "{\"message_id\":321}",
      "createdAt": "2026-03-08T10:00:00Z",
      "startedAt": "2026-03-08T10:00:01Z",
      "finishedAt": "2026-03-08T10:00:02Z"
    }
  ]
}
```

失败响应：

- `404 PUBLISH_JOB_NOT_FOUND`

### 8.4 重试单个失败投递

`POST /api/v1/delivery-tasks/:id/retry`

路径参数：

| 参数 | 说明 |
|---|---|
| `id` | DeliveryTask ID |

成功响应：`202 Accepted`

```json
{
  "deliveryId": "01delivery...",
  "status": "PENDING"
}
```

业务规则：

- 只有 `FAILED` 状态的投递允许重试
- 重试会清空原有错误信息并异步重新执行

失败响应：

- `404 DELIVERY_TASK_NOT_FOUND`
- `400 VALIDATION_ERROR`，例如 `only failed deliveries can be retried`

## 9. 渠道配置字段说明

本节用于说明 `ChannelAccount.config` 与 `ChannelTarget.config` 的结构约束。

### 9.1 Telegram

`channelType = telegram`

账号创建：

```json
{
  "channelType": "telegram",
  "name": "Telegram Bot",
  "secretRef": "TELEGRAM_BOT_TOKEN",
  "config": {}
}
```

账号规则：

- `secretRef` 必填
- `secretRef` 必须对应已存在的环境变量

Target 配置字段：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `chatId` | string | 是 | 群组 chat id |
| `topicId` | number | 否 | Topic ID，对应 `message_thread_id` |
| `messageThreadId` | number | 否 | `topicId` 的别名，后端也接受 |
| `topicName` | string | 否 | Topic 展示名 |
| `disableNotification` | boolean | 否 | 是否静默发送 |
| `disableWebPagePreview` | boolean | 否 | 是否关闭链接预览 |

Target 规则：

- `telegram_group` 不能带 `topicId`
- `telegram_topic` 必须带 `topicId`
- `telegram_topic` 必须提供非空 `targetName`
- `targetKey` 最终会被规范化为：
  - 群组：`chat_id`
  - Topic：`chat_id:topic:topic_id`

### 9.2 Enterprise Feishu

`channelType = feishu`

账号创建：

```json
{
  "channelType": "feishu",
  "name": "Enterprise Feishu",
  "secretRef": "FEISHU_APP_SECRET",
  "config": {
    "appIdEnv": "FEISHU_APP_ID",
    "tokenEnv": "FEISHU_TENANT_ACCESS_TOKEN",
    "baseUrl": "https://open.feishu.cn"
  }
}
```

账号规则：

- `secretRef` 必填，默认表示 app secret 的环境变量名
- 若 `config.tokenEnv` 存在且其环境变量有值，则优先直接使用该 token
- 否则要求 `config.appIdEnv` 与 `secretRef` 对应的环境变量都存在

Target 配置字段：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `chatId` | string | 是 | 群聊 chat id |
| `receiveIdType` | string | 否 | 默认 `chat_id` |

Target 规则：

- 仅支持 `targetType = feishu_chat`
- `targetKey` 最终规范化为 `chatId`

### 9.3 Personal Feishu

`channelType = personal_feishu`

账号创建：

```json
{
  "channelType": "personal_feishu",
  "name": "Personal Feishu Bot",
  "secretRef": "PERSONAL_FEISHU_WEBHOOK_URL",
  "config": {
    "signSecretRef": "PERSONAL_FEISHU_SIGN_SECRET"
  }
}
```

账号规则：

- `secretRef` 优先作为 webhook URL 的环境变量名
- `config.signSecretRef` 必填，表示签名密钥的环境变量名
- 当前接口会在列表返回时清理 `configJson` 内可能出现的 `webhookUrl` 和 `signSecret`

Target 配置字段：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `webhookEnvRef` | string | 否 | 通常由账号 `secretRef` 自动注入 |

Target 规则：

- 仅支持 `targetType = personal_feishu_webhook`
- 创建 target 时即使 `targetKey` 为空也可以
- 后端会把账号的 `secretRef` 注入到目标配置中
- `targetKey` 最终会被规范化为 `webhook:` 前缀的稳定哈希，而不直接暴露 webhook URL

## 10. 推荐调用顺序

典型使用流程：

1. 调用 `POST /api/v1/channel-accounts` 创建渠道账号
2. 调用 `POST /api/v1/channel-targets` 创建账号下的目标
3. 调用 `POST /api/v1/contents/upload` 上传 Markdown 内容
4. 调用 `POST /api/v1/publish-jobs` 发起发布
5. 调用 `GET /api/v1/publish-jobs/:id` 轮询任务与投递状态
6. 如有失败投递，调用 `POST /api/v1/delivery-tasks/:id/retry` 重试
