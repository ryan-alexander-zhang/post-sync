# post-sync

`post-sync` 是一个面向 Markdown 内容源的内容分发器。用户可以将本地 Markdown 文件上传到系统，统一解析 frontmatter 与正文，再按模板发布到一个或多个外部渠道。

当前版本定位为 MVP：先打通 Markdown -> 内容模型 -> 模板渲染 -> Telegram 群组 / Topic 与 Feishu 群聊富文本投递 -> 历史查询 -> Docker 部署的最小闭环，并为后续新增渠道保留清晰扩展点。

## 功能清单

- 上传 Markdown 文本文件
- 解析 YAML frontmatter 与正文
- 生成统一 Content Model
- 配置 Telegram 与 Feishu 渠道账号
- 配置 Telegram 群组 root target / topic target 与 Feishu chat target
- 选择一个内容发布到一个或多个 target
- 并行投递并记录每个投递项状态
- 基于正文标准化哈希进行去重
- 查看发布任务历史和投递明细
- 兼容 SQLite / PostgreSQL
- 支持 Docker 化部署

## MVP 范围

当前实现目标：

- 内容上传与解析
- Telegram 群组 / Topic 投递
- Feishu 群聊富文本投递
- 基础模板渲染
- 发布历史查询
- Docker / docker-compose 部署

当前不做：

- X、RedNote、微信公众号、博客平台
- 定时发布、草稿、审批流
- 分布式队列、复杂调度系统
- 高级模板 DSL

详细设计见 [design.md](/Users/erpang/GitHubProjects/post-sync/design.md)，关键决策见 [decision.md](/Users/erpang/GitHubProjects/post-sync/decision.md)。

## 协作约束

仓库级协作约束见 [AGENTS.md](/Users/erpang/GitHubProjects/post-sync/AGENTS.md)。

当前强制要求：

- 核心决策必须记录到 [decision.md](/Users/erpang/GitHubProjects/post-sync/decision.md)
- [design.md](/Users/erpang/GitHubProjects/post-sync/design.md) 必须包含模块划分、数据模型、API 设计、状态设计、去重设计、模板设计、部署设计
- [decision.md](/Users/erpang/GitHubProjects/post-sync/decision.md) 必须明确记录新增依赖、关键架构决策、关键假设
- 提交必须按“小模块、低耦合、易 review”原则拆分

## 技术栈

### Backend

- Golang
- Gin
- Gorm
- SQLite / PostgreSQL

### Frontend

- Next.js
- React
- Tailwind CSS
- shadcn/ui
- lucide-react

### Deployment

- Docker
- docker-compose

## 目录结构

当前仓库是设计起步阶段，建议按如下结构实现：

```text
.
├── backend
│   ├── cmd/server
│   └── internal
│       ├── api
│       ├── app
│       ├── channel
│       │   ├── feishu
│       │   └── telegram
│       ├── config
│       ├── db
│       ├── domain
│       ├── parser
│       ├── render
│       ├── repository
│       └── service
├── frontend
│   ├── app
│   ├── components
│   └── lib
├── deploy
│   └── docker
├── scripts
├── design.md
├── decision.md
└── README.md
```

## 快速开始

### 1. 准备环境

建议版本：

- Go `1.23+`
- Node.js `20+`
- npm `10+`
- Docker `24+`

### 2. 配置环境变量

创建 `.env`：

```env
APP_ENV=development
SERVER_ADDR=:8080
DB_DRIVER=sqlite
DB_DSN=./data/post-sync.db

TELEGRAM_BOT_TOKEN=your_bot_token
FEISHU_APP_ID=cli_xxx
FEISHU_APP_SECRET=app_secret_xxx
FEISHU_TENANT_ACCESS_TOKEN=
PERSONAL_FEISHU_WEBHOOK_URL=https://open.feishu.cn/open-apis/bot/v2/hook/xxxx
PERSONAL_FEISHU_SIGN_SECRET=custom_bot_sign_secret

PUBLISH_MAX_PARALLELISM=5
PUBLISH_TIMEOUT_SECONDS=20
HTTP_READ_TIMEOUT_SECONDS=10
HTTP_WRITE_TIMEOUT_SECONDS=30

NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1
```

### 3. 启动本地开发环境

后端：

```bash
make dev-backend
```

前端：

```bash
make dev-frontend
```

如果前端开发态出现 `.next` 缓存损坏、样式丢失或 `ENOENT app-build-manifest.json` 这类问题，可直接重启：

```bash
make restart-frontend
```

如果后端需要释放 `8080` 端口并重启：

```bash
make restart-backend
```

默认访问：

- API: `http://localhost:8080`
- Web: `http://localhost:3000`

## 环境变量说明

| 变量 | 必填 | 说明 |
|---|---|---|
| `APP_ENV` | 否 | 运行环境，如 `development` |
| `SERVER_ADDR` | 否 | 后端监听地址 |
| `DB_DRIVER` | 是 | `sqlite` 或 `postgres` |
| `DB_DSN` | 是 | SQLite 文件路径或 PostgreSQL DSN |
| `TELEGRAM_BOT_TOKEN` | 否 | Telegram Bot Token |
| `FEISHU_APP_ID` | 否 | Feishu 应用 app id |
| `FEISHU_APP_SECRET` | 否 | Feishu 应用 app secret |
| `FEISHU_TENANT_ACCESS_TOKEN` | 否 | Feishu 调试时可直接提供的 tenant access token |
| `PERSONAL_FEISHU_WEBHOOK_URL` | 否 | Personal Feishu 自定义机器人的 webhook URL |
| `PERSONAL_FEISHU_SIGN_SECRET` | 否 | Personal Feishu 自定义机器人的签名秘钥 |
| `PUBLISH_MAX_PARALLELISM` | 否 | 单任务最大发送并发数 |
| `PUBLISH_TIMEOUT_SECONDS` | 否 | 单次投递超时 |
| `HTTP_READ_TIMEOUT_SECONDS` | 否 | HTTP 读超时 |
| `HTTP_WRITE_TIMEOUT_SECONDS` | 否 | HTTP 写超时 |
| `NEXT_PUBLIC_API_BASE_URL` | 是 | 前端调用 API 地址 |

## 本地开发方式

### 使用 SQLite

适合本地开发和单机试用：

```env
DB_DRIVER=sqlite
DB_DSN=./data/post-sync.db
```

优点：

- 零外部依赖
- 启动快

### 使用 PostgreSQL

适合验证生产形态兼容性：

```env
DB_DRIVER=postgres
DB_DSN=host=127.0.0.1 user=postgres password=postgres dbname=post_sync port=5432 sslmode=disable TimeZone=UTC
```

要求业务代码始终通过统一 repository 访问数据库，不直接写数据库方言分支。

## Docker / docker-compose 启动方式

### Docker 镜像构建

建议提供多阶段 Dockerfile，并通过 `buildx` 处理不同 CPU 架构：

```bash
./scripts/build-image.sh post-sync:local
```

建议脚本支持：

- `linux/amd64`
- `linux/arm64`
- `--load`
- `--push`

### docker-compose 启动

SQLite 模式：

```bash
docker compose up -d
```

PostgreSQL 模式建议通过 profile 或单独服务启用：

```bash
docker compose --profile postgres up -d
```

建议在 `docker-compose.yaml` 中包含：

- `app`
- `postgres`（可选 profile）
- 应用数据卷

## 示例使用流程

1. 在 `/channels` 配置一个 Telegram、Enterprise Feishu 或 Personal Feishu channel account。
2. Telegram 账号一般使用 `secretRef=TELEGRAM_BOT_TOKEN`。
3. Enterprise Feishu 账号一般使用 `secretRef=FEISHU_APP_SECRET`，并在配置里填写 `FEISHU_APP_ID`。
4. Personal Feishu 账号使用：
   - `secretRef=PERSONAL_FEISHU_WEBHOOK_URL`
   - `signSecretRef=PERSONAL_FEISHU_SIGN_SECRET`
3. 在 `/channels` 新增 target：
   - Telegram：填写群组 `chat_id`；若要发到某个 topic，再补 `topic_id`
   - Enterprise Feishu：填写群聊 `chat_id`
   - Personal Feishu：只填写 target 名称，复用账号里的 webhook 配置
4. 在 `/contents` 上传 Markdown 文件。
5. 系统解析 frontmatter，生成内容记录和 `body_hash`。
6. 在 `/publish/new` 选择内容、target、模板，发起发布。
7. 系统创建 `PublishJob` 和多个 `DeliveryTask` 并并行发送。
8. 在 `/history` 查看任务状态，在 `/history/[jobId]` 查看每个 target 的结果。
9. 若内容正文未变化，再次向同一 target 发布时会标记 `SKIPPED_DUPLICATE`。
10. 同一 Telegram 群组下不同 topic 被视为不同 target，不会互相去重。

## 架构概览

核心模型：

- `Content`：上传后的内容快照
- `ChannelAccount`：渠道账号配置
- `ChannelTarget`：可投递目标，Telegram 下既可以是 group root / topic，Feishu 下为 chat
- `PublishJob`：一次发布任务
- `DeliveryTask`：单 target 投递记录

核心流程：

1. 上传 Markdown
2. 解析 frontmatter 与正文
3. 标准化正文并计算 `body_hash`
4. 组装模板上下文
5. 通过 `ChannelDriver` 渲染并发送
6. 聚合投递状态，形成发布历史

去重原则：

- 仅基于“去除 frontmatter 后的标准化正文”
- 粒度为 `channel_type + target_key + body_hash`
- 仅跳过历史上已经 `SUCCESS` 的重复投递

Target 规则：

- Telegram group root 的 `target_key = chat_id`
- Telegram topic 的 `target_key = chat_id:topic:topic_id`
- Feishu chat 的 `target_key = chat_id`
- `config_json` 保存实际发送所需的渠道参数

## API 概览

完整接口说明见 [api.md](/Users/erpang/GitHubProjects/post-sync/api.md)。

当前主要接口：

- `POST /api/v1/contents/upload`
- `GET /api/v1/contents`
- `GET /api/v1/contents/:id`
- `DELETE /api/v1/contents/:id`
- `GET /api/v1/channel-accounts`
- `POST /api/v1/channel-accounts`
- `PATCH /api/v1/channel-accounts/:id`
- `DELETE /api/v1/channel-accounts/:id`
- `GET /api/v1/channel-targets`
- `POST /api/v1/channel-targets`
- `PATCH /api/v1/channel-targets/:id`
- `DELETE /api/v1/channel-targets/:id`
- `POST /api/v1/publish-jobs`
- `GET /api/v1/publish-jobs`
- `GET /api/v1/publish-jobs/:id`
- `POST /api/v1/delivery-tasks/:id/retry`

## 后续规划

- 新增 X、博客平台驱动
- 增加模板管理与更多渲染模式
- 增加定时发布和失败批量重试
- 将应用内异步执行升级为独立 worker
- 增加草稿、审批流和多用户权限

## 开发建议

推荐按以下顺序继续实现：

1. 初始化后端和前端工程骨架
2. 定义 Gorm 模型与数据库迁移
3. 实现 Markdown 解析、标准化、去重哈希
4. 定义 `ChannelDriver` 并实现 Telegram / Feishu adapter
5. 实现发布编排与状态聚合
6. 实现 API
7. 实现前端页面
8. 补 Dockerfile、docker-compose 和构建脚本
