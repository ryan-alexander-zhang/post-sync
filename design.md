# 内容分发器 MVP 技术方案

## 1. 项目目标与范围

本项目用于接收本地 Markdown 内容，并将同一篇内容发布到一个或多个外部渠道。MVP 目标是先打通从 Markdown 上传、内容解析、模板渲染、渠道投递、历史查询到 Docker 部署的最小闭环，同时保证后续新增渠道时不需要重写业务主流程。

当前明确约束如下：

- 内容源为 Markdown 文本文件，允许包含 frontmatter 元信息。
- 系统内部需要统一抽象内容、渠道账号、渠道目标、发布任务、单渠道投递结果。
- 去重逻辑必须基于“去除 Meta 后的正文内容”，而不是整篇原始文件。
- MVP 当前实现 Telegram 群组 / Topic 与 Feishu 群聊富文本投递，但所有业务层接口按多渠道设计。
- 数据存储通过配置切换 SQLite 或 PostgreSQL，不允许业务层分叉。

本方案目标不是一次性设计完整内容平台，而是为后续编码提供足够具体、可拆分实现的落地蓝图。

## 2. MVP 范围与非目标

### 2.1 MVP 当前实现

- 上传单个 Markdown 文件并持久化。
- 解析 frontmatter 与正文，生成统一 Content Model。
- 管理 Telegram 渠道账号、群组 root target 与群组 topic target。
- 管理 Feishu 渠道账号与 chat target。
- 选择一个内容并发布到一个或多个 Telegram / Feishu target。
- 多 target 并行投递，并记录每个投递项状态。
- 基础模板渲染：支持系统字段、frontmatter 字段、渠道参数注入。
- 历史记录查询：按发布任务和单投递结果查看。
- SQLite / PostgreSQL 双兼容。
- Dockerfile、docker-compose、本地开发说明。

### 2.2 非目标

- X、RedNote、微信公众号、博客平台接入。
- 定时发布、审批流、草稿箱、多人协作。
- 分布式队列、复杂调度中心、事件总线。
- 高级模板 DSL、可视化模板编辑器。
- 大规模多租户隔离、权限系统、SSO。

### 2.3 预留扩展点

- `ChannelDriver` 接口支持新增平台驱动。
- `channel_accounts` / `channel_targets` 通用化支持更多渠道目标。
- `render_mode`、`template_name`、`config_json` 支持更多渲染和投递策略。
- `publish_jobs` / `delivery_tasks` 拆分后可平滑升级到异步 worker。

## 3. 总体架构设计

系统采用单体应用 + 前后端分层的 MVP 架构：

1. 前端：Next.js 管理界面，用于上传内容、配置渠道、发起发布、查看历史。
2. 后端 API：Gin 提供 REST API，负责内容管理、渠道管理、发布编排、历史查询。
3. 领域服务层：封装内容解析、模板渲染、去重计算、发布状态聚合。
4. 渠道适配层：通过统一接口调用具体渠道实现，当前提供 Telegram 与 Feishu adapter。
5. 持久化层：Gorm + repository，兼容 SQLite / PostgreSQL。
6. 部署层：Docker 多阶段构建，`docker-compose` 提供本地一键启动。

建议目录结构：

```text
backend/
  cmd/server/
  internal/
    api/
    app/
    domain/
    repository/
    service/
    channel/
      feishu/
      telegram/
    render/
    parser/
    config/
    db/
frontend/
  app/
  components/
  lib/
deploy/
  docker/
scripts/
docs/
```

前后端解耦，前端只依赖 REST API，不直接感知具体渠道实现。后端业务主流程只依赖渠道接口，不直接 import 某个平台 SDK 到业务服务层。

## 4. 核心模块划分

### 4.1 Backend 模块

- `content` 模块：上传 Markdown、解析 frontmatter、正文标准化、内容入库。
- `channel` 模块：管理渠道账号、target、参数校验、驱动发现。
- `publish` 模块：创建发布任务、生成投递项、并行执行、状态聚合。
- `render` 模块：构造模板上下文并执行模板；具体渠道格式转换由各 driver 负责。
- `history` 模块：查询发布任务与投递历史。
- `audit/log` 模块：记录最小审计字段与结构化日志。
- `config` 模块：装载数据库、服务、渠道密钥等配置。

### 4.2 Frontend 模块

- 内容上传页
- 内容列表 / 详情页
- 发布配置页
- 渠道管理页
- 发布历史页
- 发布任务详情页

### 4.3 模块边界原则

- Handler 只负责请求解析和响应。
- Service 负责业务编排，不直接操作数据库实现细节。
- Repository 只暴露领域友好的查询和写入接口。
- Channel adapter 只负责渠道协议转换与外部 API 调用。
- Render 不感知具体 HTTP 请求，仅接收内容和渠道上下文。

## 5. 领域模型 / 数据模型

### 5.1 Content

统一表示一次上传后的内容快照。

- `id`
- `source_filename`
- `original_markdown`
- `frontmatter_json`
- `title`
- `body_markdown`
- `body_plain`
- `body_hash`
- `created_at`

说明：

- `original_markdown` 保留原始文件，便于排错和重渲染。
- `body_markdown` 为移除 frontmatter 且标准化后的正文。
- `body_plain` 为纯文本摘要或降级发送使用。
- `body_hash` 为去重核心字段。

### 5.2 ChannelAccount

表示一个可用的渠道账号或连接配置。

- `id`
- `channel_type`，例如 `telegram`、`feishu`
- `name`
- `enabled`
- `secret_ref`，例如 `TELEGRAM_BOT_TOKEN`
- `config_json`
- `created_at`
- `updated_at`

### 5.3 ChannelTarget

表示账号下一个可投递的目标。对于 Telegram，target 可以是群组根节点，也可以是群组下某个 topic。

- `id`
- `channel_account_id`
- `target_type`，当前支持 `telegram_group`、`telegram_topic`、`feishu_chat`
- `target_key`，Telegram 的逻辑目标键。群组 root 为 `chat_id`，topic 为 `chat_id:topic:topic_id`；Feishu 为 `chat_id`
- `target_name`
- `enabled`
- `config_json`
- `created_at`
- `updated_at`

### 5.4 PublishJob

表示一次“将某内容发布到一组目标”的业务任务。

- `id`
- `content_id`
- `status`
- `request_id`
- `trigger_source`，MVP 固定 `manual`
- `total_deliveries`
- `success_count`
- `failed_count`
- `skipped_count`
- `created_at`
- `started_at`
- `finished_at`

### 5.5 DeliveryTask

表示一次单渠道单目标投递。

- `id`
- `publish_job_id`
- `content_id`
- `channel_account_id`
- `channel_target_id`
- `channel_type`
- `target_key`
- `status`
- `attempt_count`
- `idempotency_key`
- `body_hash`
- `template_name`
- `render_mode`
- `rendered_title`
- `rendered_body`
- `external_message_id`
- `error_code`
- `error_message`
- `provider_response_json`
- `created_at`
- `started_at`
- `finished_at`

关系说明：

- 一个 `PublishJob` 对应多个 `DeliveryTask`。
- `PublishJob.status` 由其下 Delivery 聚合得出。
- 历史展示以 `PublishJob` 为主视图，以 `DeliveryTask` 为明细视图。

## 6. 渠道抽象设计

### 6.1 核心接口

```go
type ChannelDriver interface {
    Type() string
    ValidateAccount(input AccountValidationInput) error
    NormalizeTarget(input TargetInput) (NormalizedTarget, error)
    Render(input RenderInput) (RenderedMessage, error)
    Send(ctx context.Context, req SendRequest) (SendResult, error)
}
```

### 6.2 抽象对象

- `AccountValidationInput`：账号级配置校验输入。
- `TargetInput` / `NormalizedTarget`：目标级参数规范化输入与结果。
- `RenderInput`：内容、模板、系统字段、渠道字段组成的统一渲染输入。
- `RenderedMessage`：标准化输出，包含 `title`、`body`、`render_mode`。
- `SendRequest`：驱动最终发送入参，包含账号、目标、幂等键等上下文。
- `SendResult`：外部 message id、原始响应、耗时、限流信息。

### 6.3 Telegram 在抽象中的实现

`telegram.Driver` 实现上述接口，但业务层只通过 `ChannelDriver` 调用。

Telegram target 的 MVP 参数：

- `chat_id`: Telegram 群组 chat id
- `topic_id`: Telegram topic 对应的 `message_thread_id`，可选
- `topic_name`: topic 展示名，可选
- `disable_web_page_preview`
- `disable_notification`

Telegram target 建模规则：

- 群组根节点存为一个 `telegram_group` target，`target_key=chat_id`
- 群组下每个 topic 存为一个独立 `telegram_topic` target，`target_key=chat_id:topic:topic_id`
- 同一群组下可配置多个 topic target

### 6.4 Feishu 在抽象中的实现

Feishu 当前实现规则：

- `ChannelAccount.channel_type = feishu`
- `ChannelTarget.target_type = feishu_chat`
- `target_key = chat_id`
- 账号配置保存 `appIdEnv`、可选 `tokenEnv`、可选 `baseUrl`
- `secret_ref` 默认指向 `FEISHU_APP_SECRET`

鉴权策略：

- 若配置了 `tokenEnv` 且环境变量存在，则直接使用该 `tenant_access_token`
- 否则通过 `appIdEnv + secret_ref` 获取并缓存 `tenant_access_token`
- token provider 作为 driver 依赖，不暴露到业务编排层

渲染策略：

- 默认模板输出 Markdown 文本
- Feishu driver 将模板结果封装为 `msg_type=post`
- `post.content` 目前使用单个 `md` 节点承载 Markdown 正文
- 若存在 `title`，写入 `post.zh_cn.title`
- 为避免重复，driver 会去掉默认模板里与标题相同的首个 `# title` 行

### 6.5 新增渠道的扩展方式

后续新增飞书、X、博客平台时，仅需要：

1. 新增驱动实现 `ChannelDriver`
2. 在 driver 内实现账号校验、target 规范化、发送逻辑
3. 如需鉴权刷新，引入渠道专属 token provider
4. 在前端 `channels.ts` 中补充 schema 和表单映射
5. 如有需要，实现专属模板转换策略

发布编排、状态机、历史查询无需重写。

## 7. 模板渲染设计

### 7.1 渲染链路

MVP 渲染链路：

1. 上传 Markdown 文件。
2. 解析 frontmatter 和正文。
3. 生成标准化 `Content`。
4. 组装模板上下文。
5. 通过模板引擎生成渠道消息骨架。
6. 由渠道驱动把正文转换为目标平台可发送格式。
7. 得到最终 `RenderedMessage` 并投递。

### 7.2 模板上下文

模板上下文分三类字段：

- 系统字段：`content.id`、`content.title`、`content.body_markdown`、`content.body_plain`、`publish.request_id`
- frontmatter 字段：`meta.tags`、`meta.summary`、`meta.slug`
- 渠道字段：`channel.target_name`、`channel.target_key`

示例：

```gotemplate
{{ if .content.title }}# {{ .content.title }}{{ end }}

{{ .content.body_markdown }}

{{ if .meta.tags }}
Tags: {{ join .meta.tags ", " }}
{{ end }}
```

### 7.3 模板能力边界

MVP 仅支持基础变量替换、条件判断、简单函数。

- 支持：`if`、`range`、字符串拼接、`join`
- 不支持：自定义脚本、外部 HTTP、复杂 DSL、模板继承

### 7.4 Telegram / Feishu 渲染策略

默认模板先输出 Markdown 文本，再由各渠道 driver 转换为目标平台格式。

Telegram：

1. 模板产出 Markdown。
2. 使用 Markdown 渲染器转为 HTML。
3. 清洗为 Telegram 支持标签子集。
4. 发送时使用 `parse_mode=HTML`。

Feishu：

1. 模板产出 Markdown。
2. 不做 HTML 转换，也不复用 Telegram HTML。
3. 将 Markdown 正文封装为 `msg_type=post` 的 `md` 节点。
4. 标题进入 `post.zh_cn.title`，正文进入 `post.zh_cn.content`。

这样可以让模板层保持统一，而把平台差异压缩在 driver 内。

## 8. 发布流程设计

### 8.1 主流程

1. 用户上传 Markdown 文件。
2. 后端解析并保存 `Content`。
3. 用户选择一个内容和多个 target 发起发布。
4. 后端创建一条 `PublishJob` 和 N 条 `DeliveryTask`。
5. 发布服务并行执行所有 `DeliveryTask`。
6. 每个 Delivery 在发送前先执行去重检查。
7. 未命中重复时执行模板渲染和渠道发送。
8. 汇总所有 Delivery 状态并更新 `PublishJob.status`。
9. 前端轮询或刷新查看任务结果。

### 8.2 同步接口 + 异步执行模型

MVP 采用“请求内创建任务，后台 goroutine 执行”的轻量异步方式：

- `POST /api/publish-jobs` 返回 `job_id`
- 服务端异步启动执行器
- 前端查询任务状态直到完成

这样比完全同步阻塞请求更稳妥，也比引入外部队列更符合 MVP。

### 8.3 状态流转

`PublishJob.status`：

- `PENDING`
- `PROCESSING`
- `SUCCESS`
- `PARTIAL_SUCCESS`
- `FAILED`

`DeliveryTask.status`：

- `PENDING`
- `PROCESSING`
- `SUCCESS`
- `FAILED`
- `SKIPPED_DUPLICATE`

聚合规则：

- 全部 `SUCCESS` => `SUCCESS`
- 存在 `FAILED` 且存在 `SUCCESS` / `SKIPPED_DUPLICATE` => `PARTIAL_SUCCESS`
- 全部 `FAILED` => `FAILED`
- 全部 `SKIPPED_DUPLICATE` => `SUCCESS`，但 job 的 `skipped_count` > 0

## 9. 并行发送与状态机设计

### 9.1 并行策略

MVP 采用单机内 goroutine + `errgroup` / `WaitGroup` 并行投递，不引入独立 worker 服务。

执行策略：

- 一个 `PublishJob` 内最多并行执行 `min(目标数, maxParallelism)` 个投递。
- 默认 `maxParallelism=5`，通过配置覆盖。
- 每个 Delivery 使用独立 context 和超时。

推荐原因：

- 足够支撑 MVP 的少量渠道目标。
- 实现简单，便于调试。
- 后续可无缝切换为数据库轮询 worker 或消息队列消费者。

### 9.2 Delivery 状态机

```text
PENDING -> PROCESSING -> SUCCESS
PENDING -> PROCESSING -> FAILED
PENDING -> SKIPPED_DUPLICATE
FAILED  -> PROCESSING -> SUCCESS   (手动重试时)
FAILED  -> PROCESSING -> FAILED
```

### 9.3 重试策略

MVP 默认在单次发布中自动重试最多 2 次，仅针对瞬时错误：

- 网络超时
- Telegram 5xx
- 明确可重试的限流错误

不自动重试：

- 认证失败
- target 不存在
- 模板渲染错误
- 内容超长且无法拆分

## 10. 去重与幂等设计

### 10.1 正文标准化规则

去重前对 Markdown 执行如下标准化：

1. 去除 YAML frontmatter。
2. 将 `\r\n` / `\r` 统一为 `\n`。
3. 删除文首文尾空白。
4. 每行去除尾部空格。
5. 连续 3 个及以上空行折叠为 2 个空行。

标准化后得到 `body_markdown`。

### 10.2 去重哈希

对 `body_markdown` 计算 `sha256`，保存为 `body_hash`。

去重粒度为：

- 相同 `channel_type`
- 相同 `target_key`
- 相同 `body_hash`
- 且历史上存在 `SUCCESS` 状态投递

则本次 Delivery 直接标记为 `SKIPPED_DUPLICATE`。

说明：

- Telegram 群组 root 与同一群组下不同 topic 的 `target_key` 不同
- 因此同一正文允许在同一个群组的不同 topic 各发一次，不会互相去重

### 10.3 幂等键

`idempotency_key = sha256(channel_type + ":" + target_key + ":" + body_hash)`

用途：

- 标准化表示一次“正文向某渠道某目标投递”的逻辑唯一性。
- 用于日志检索和历史追踪。
- 后续若引入分布式队列，可复用为幂等消费键。

### 10.4 为什么不使用数据库唯一约束阻断

MVP 不直接用唯一索引阻止插入，因为：

- 需要保留“本次请求被跳过”的历史记录。
- 失败后需要允许再次投递。
- SQLite / PostgreSQL 对带条件唯一索引的兼容和迁移复杂度更高。

因此采用“查询历史成功记录 + 生成 `SKIPPED_DUPLICATE` 明细”的业务层方案。

## 11. 数据库表设计

### 11.1 `contents`

| 字段 | 类型建议 | 说明 |
|---|---|---|
| id | uuid / string | 主键 |
| source_filename | varchar(255) | 原始文件名 |
| original_markdown | text | 原始 Markdown |
| frontmatter_json | json/text | frontmatter 原文解析结果 |
| title | varchar(255) | 标题 |
| body_markdown | text | 标准化正文 |
| body_plain | text | 纯文本版本 |
| body_hash | char(64) | 正文哈希 |
| created_at | timestamp | 创建时间 |

索引：

- `idx_contents_body_hash`
- `idx_contents_created_at`

### 11.2 `channel_accounts`

| 字段 | 类型建议 | 说明 |
|---|---|---|
| id | uuid / string | 主键 |
| channel_type | varchar(50) | 渠道类型 |
| name | varchar(100) | 展示名 |
| enabled | bool | 是否启用 |
| secret_ref | varchar(100) | 环境变量名 |
| config_json | json/text | 非敏感账号配置 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |

索引：

- `idx_channel_accounts_type`

### 11.3 `channel_targets`

| 字段 | 类型建议 | 说明 |
|---|---|---|
| id | uuid / string | 主键 |
| channel_account_id | uuid / string | 所属账号 |
| target_type | varchar(50) | `telegram_group`、`telegram_topic`、`feishu_chat` |
| target_key | varchar(255) | 逻辑目标键，topic 目标为复合键 |
| target_name | varchar(100) | 显示名称 |
| enabled | bool | 是否启用 |
| config_json | json/text | 目标参数，Telegram 包含 `chatId`、`topicId`、`topicName`；Feishu 包含 `chatId`、`receiveIdType` |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |

索引：

- `idx_channel_targets_account_id`
- `idx_channel_targets_target_key`

### 11.4 `publish_jobs`

| 字段 | 类型建议 | 说明 |
|---|---|---|
| id | uuid / string | 主键 |
| content_id | uuid / string | 内容 ID |
| request_id | varchar(100) | 请求追踪 ID |
| trigger_source | varchar(50) | 触发来源 |
| status | varchar(30) | 聚合状态 |
| total_deliveries | int | 总投递数 |
| success_count | int | 成功数 |
| failed_count | int | 失败数 |
| skipped_count | int | 跳过数 |
| created_at | timestamp | 创建时间 |
| started_at | timestamp | 开始时间 |
| finished_at | timestamp | 完成时间 |

索引：

- `idx_publish_jobs_content_id`
- `idx_publish_jobs_status`
- `idx_publish_jobs_created_at`

### 11.5 `delivery_tasks`

| 字段 | 类型建议 | 说明 |
|---|---|---|
| id | uuid / string | 主键 |
| publish_job_id | uuid / string | 所属任务 |
| content_id | uuid / string | 内容 ID |
| channel_account_id | uuid / string | 渠道账号 |
| channel_target_id | uuid / string | 渠道目标 |
| channel_type | varchar(50) | 渠道类型冗余 |
| target_key | varchar(255) | 目标标识冗余 |
| status | varchar(30) | 投递状态 |
| attempt_count | int | 已尝试次数 |
| idempotency_key | char(64) | 幂等键 |
| body_hash | char(64) | 正文哈希 |
| template_name | varchar(100) | 模板名 |
| render_mode | varchar(30) | 渲染模式 |
| rendered_title | text | 渲染标题 |
| rendered_body | text | 渲染结果 |
| external_message_id | varchar(255) | 外部消息 ID |
| error_code | varchar(100) | 错误码 |
| error_message | text | 错误信息 |
| provider_response_json | json/text | 外部响应快照 |
| created_at | timestamp | 创建时间 |
| started_at | timestamp | 开始时间 |
| finished_at | timestamp | 完成时间 |

索引：

- `idx_delivery_tasks_publish_job_id`
- `idx_delivery_tasks_status`
- `idx_delivery_tasks_lookup_dedup (channel_type, target_key, body_hash, status)`
- `idx_delivery_tasks_created_at`

## 12. API 设计

API 前缀统一为 `/api/v1`。

### 12.1 Content API

- `POST /contents/upload`
  - `multipart/form-data`
  - 参数：`file`
  - 返回：内容 ID、标题、body hash、frontmatter 摘要

- `GET /contents`
  - 查询内容列表
  - 支持分页、按标题模糊查询

- `GET /contents/:id`
  - 返回内容详情、frontmatter、正文预览

### 12.2 Channel API

- `GET /channel-accounts`
- `POST /channel-accounts`
- `PATCH /channel-accounts/:id`
- `GET /channel-targets`
- `POST /channel-targets`
- `PATCH /channel-targets/:id`

`POST /channel-accounts` 请求示例：

```json
{
  "channelType": "telegram",
  "name": "main-bot",
  "secretRef": "TELEGRAM_BOT_TOKEN",
  "config": {
    "parseMode": "HTML"
  }
}
```

### 12.3 Publish API

- `POST /publish-jobs`

请求示例：

```json
{
  "contentId": "cnt_123",
  "targetIds": ["tgt_1", "tgt_2"],
  "templateName": "default"
}
```

返回示例：

```json
{
  "jobId": "job_123",
  "status": "PENDING"
}
```

- `GET /publish-jobs`
  - 列表查询，支持按状态、时间、内容 ID 筛选

- `GET /publish-jobs/:id`
  - 返回任务详情及 delivery 明细

- `POST /delivery-tasks/:id/retry`
  - 手动重试单条失败投递

### 12.4 错误响应格式

```json
{
  "code": "TARGET_NOT_FOUND",
  "message": "channel target not found",
  "requestId": "req_xxx"
}
```

## 13. 前端页面设计

### 13.1 页面最小清单

- `/`：总览页，展示最近发布任务和成功/失败统计。
- `/contents`：内容列表页。
- `/contents/[id]`：内容详情页，展示 frontmatter、正文预览、body hash、发布按钮。
- `/publish/new`：发起发布页，选择内容、目标、模板。
- `/channels`：渠道账号和 target 管理页，支持配置 Telegram Group root / Topic 与 Feishu Chat target。
- `/history`：发布历史页。
- `/history/[jobId]`：任务详情页，查看单 target 状态、错误信息、外部消息 ID。

### 13.2 组件建议

优先使用 shadcn/ui：

- `Table`
- `Card`
- `Badge`
- `Dialog`
- `Form`
- `Input`
- `Textarea`
- `Select`
- `Checkbox`
- `Tabs`
- `Alert`

图标使用 `lucide-react`。

### 13.3 交互重点

- 上传成功后立即展示解析结果和 `body_hash`。
- 发起发布时按 target 多选，并清晰提示可能被去重跳过。
- 渠道页需要清楚展示某个 target 的渠道类型、Group root / Topic / Chat 子类型，以及所属逻辑 ID。
- 历史页需要区分任务状态和单投递状态。
- 失败投递提供“重试”按钮。

## 14. 错误处理与重试策略

### 14.1 错误分类

- `VALIDATION_ERROR`：文件格式、参数缺失、模板变量错误。
- `CONFIG_ERROR`：缺少 bot token、secretRef 未映射。
- `RENDER_ERROR`：模板执行失败、Markdown 转换失败。
- `CHANNEL_ERROR`：Telegram / Feishu API 4xx / 5xx。
- `DB_ERROR`：数据库读写失败。

### 14.2 API 层原则

- 对外返回稳定错误码，不直接暴露底层异常堆栈。
- 日志中保留原始错误链和 request id。
- 对用户可修复问题给出明确 message。

### 14.3 重试原则

- 自动重试：最多 2 次，指数退避 `1s -> 3s`
- 手动重试：只允许针对 `FAILED` 的 `DeliveryTask`
- `SKIPPED_DUPLICATE` 不允许重试，除非内容变化后创建新任务

### 14.4 Telegram topic 投递规则

- `telegram_group` target 发送到群组根节点，不携带 `message_thread_id`
- `telegram_topic` target 发送到指定 topic，并携带 `message_thread_id=topic_id`
- 同一个群组下不同 topic 的投递彼此独立，状态、去重、重试均分别计算

### 14.5 Feishu 鉴权与投递规则

- 若配置 `tokenEnv` 且环境变量存在，直接使用静态 `tenant_access_token`
- 否则使用 `appIdEnv + secret_ref` 请求 `tenant_access_token`
- 当前发送 `msg_type=post`
- `feishu_chat` target 使用 `chat_id` 作为 `target_key`
- 当前富文本策略依赖飞书 `post` 的 `md` 标签支持，不手工拆分细粒度富文本节点

## 15. 配置与安全设计

### 15.1 配置分层

- 环境变量：密钥、数据库 DSN、并发数、超时
- 配置文件：可选，非敏感默认值
- 数据库：渠道 target 等业务配置

### 15.2 敏感配置存储

敏感信息不写入数据库明文：

- Telegram Bot Token 存环境变量
- Feishu App Secret / App ID / 可选静态 tenant access token 存环境变量
- 数据库中仅存 `secret_ref`
- 运行时由配置模块解析 `secret_ref -> env value`

示例：

```env
APP_ENV=development
SERVER_ADDR=:8080
DB_DRIVER=sqlite
DB_DSN=./data/post-sync.db
TELEGRAM_BOT_TOKEN=xxx
FEISHU_APP_ID=cli_xxx
FEISHU_APP_SECRET=xxx
FEISHU_TENANT_ACCESS_TOKEN=
PUBLISH_MAX_PARALLELISM=5
PUBLISH_TIMEOUT_SECONDS=20
```

### 15.3 最小日志 / 审计字段

- `timestamp`
- `level`
- `request_id`
- `job_id`
- `delivery_id`
- `content_id`
- `channel_type`
- `target_key`
- `idempotency_key`
- `status`
- `error_code`
- `duration_ms`

## 16. 部署设计

### 16.1 本地开发

- 后端：`go run ./backend/cmd/server`
- 前端：`npm run dev --prefix frontend`
- SQLite 默认用于本地开发
- 如验证 PostgreSQL 兼容性，通过 `.env` 切换 `DB_DRIVER=postgres`

### 16.2 Dockerfile

采用多阶段构建：

1. Node 阶段构建前端静态资源或 Next.js production build
2. Go 阶段编译后端二进制
3. Runtime 阶段使用轻量基础镜像运行

架构兼容策略：

- 使用 `docker buildx build`
- 在 Dockerfile 中使用 `ARG TARGETOS TARGETARCH`
- Go 编译时显式传入 `GOOS=$TARGETOS GOARCH=$TARGETARCH`

### 16.3 docker-compose

提供两个典型启动模式：

- `sqlite` 模式：仅启动应用容器，数据卷挂载 SQLite 文件
- `postgres` 模式：启动 `app + postgres`

建议文件：

```text
docker-compose.yaml
.env.example
scripts/build-image.sh
```

### 16.4 镜像构建脚本

`scripts/build-image.sh` 建议职责：

- 接收镜像名和 tag
- 默认启用 `buildx`
- 支持 `linux/amd64`、`linux/arm64`
- 可选 `--load` / `--push`

## 17. 后续扩展路线

### 17.1 近期演进

- 增加内容列表筛选和全文搜索
- 支持自定义模板管理
- 支持失败批量重试

### 17.2 中期演进

- 将后台 goroutine 升级为数据库驱动 worker
- 增加定时发布
- 增加草稿和审批状态
- 增加渠道限流与更细粒度重试策略

### 17.3 长期演进

- 分布式队列
- 多租户权限模型
- 高级模板 DSL
- 内容版本比较和回滚

## 18. 推荐实施顺序

为了便于小步提交，建议按以下顺序编码：

1. 初始化后端/前端工程骨架与配置加载
2. 建立数据模型与数据库迁移
3. 实现 Markdown 解析、标准化和去重哈希
4. 实现渠道抽象接口与 Telegram / Feishu driver
5. 实现发布任务编排与 Delivery 状态机
6. 实现 Content / Channel / Publish API
7. 实现前端上传、发布、历史页面
8. 补 Dockerfile、docker-compose、构建脚本
