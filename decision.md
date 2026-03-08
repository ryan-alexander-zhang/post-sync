# 内容分发器 MVP 决策记录

本文记录当前仓库首版 MVP 方案的关键架构决策、依赖引入计划与保守假设。由于仓库当前为空仓库，本文件同时承担首版 ADR 集合的角色。

## 决策 1：发布流程采用轻量异步任务模型

### 标题

使用“创建任务后后台执行”的轻量异步模型，而不是请求全程同步阻塞。

### 背景

一次发布可能会向多个 Telegram target 发送消息，外部 API 时延、网络抖动、限流和失败重试都会拉长请求时间。若完全同步处理，前端接口超时和用户体验都会变差；若引入完整消息队列，又超出 MVP 需要。

### 备选方案

- 方案 A：HTTP 请求内同步完成全部发送。
- 方案 B：落库后由应用内 goroutine 异步执行。
- 方案 C：引入 Redis / MQ / 专用 worker 异步执行。

### 最终选择

选择方案 B。

### 原因

- 明显优于同步阻塞请求。
- 不需要在 MVP 引入额外基础设施。
- `publish_jobs` / `delivery_tasks` 已经具备后续迁移到 worker 的数据基础。

### 影响

- API 需要返回 `jobId`，前端轮询任务状态。
- 应用重启时，未完成任务可能需要在启动阶段补偿扫描。
- 后续切换到外部队列时，业务模型和状态机可直接复用。

## 决策 2：多渠道并行发送采用单机 goroutine 并发

### 标题

使用单机内 goroutine + 有界并发控制执行同一任务下的多 target 投递。

### 背景

MVP 需要支持多渠道并行发送和查看每个渠道状态，但当前只实现 Telegram，目标规模可预期较小。

### 备选方案

- 方案 A：串行投递。
- 方案 B：goroutine + `errgroup`/`WaitGroup` + semaphore。
- 方案 C：独立 worker pool 服务。

### 最终选择

选择方案 B。

### 原因

- 并发实现简单，足够覆盖少量 target。
- 失败隔离优于串行。
- 比独立 worker 更易调试和部署。

### 影响

- 需要增加 `PUBLISH_MAX_PARALLELISM` 配置。
- 需要处理共享状态汇总和并发日志。
- 后续扩展到分布式队列时，当前并发模型会被替换，但不会影响外部 API。

## 决策 3：Telegram 通过统一驱动接口接入

### 标题

用 `ChannelDriver` 统一抽象渠道能力，Telegram 只是其中一个实现。

### 背景

项目未来会扩展飞书、X、博客平台等渠道。如果在发布服务里直接耦合 Telegram SDK 或 Telegram 专属字段，后续新增渠道成本会迅速上升。

### 备选方案

- 方案 A：业务服务直接调用 Telegram 客户端。
- 方案 B：定义统一驱动接口，业务层只依赖接口。
- 方案 C：直接按插件系统动态加载渠道。

### 最终选择

选择方案 B。

### 原因

- 满足“业务层不得直接耦合具体渠道实现”的约束。
- 复杂度远低于插件系统。
- 便于测试：可对 `ChannelDriver` 做 mock。

### 影响

- 需要定义账号配置、target 配置、渲染结果、发送结果等统一对象。
- 某些平台特有能力需要通过 `config_json` 或扩展字段承载。

## 决策 4：去重算法基于去除 Meta 后的标准化正文哈希

### 标题

去重使用“移除 frontmatter 后的标准化正文”的 SHA-256。

### 背景

需求明确要求重复计算不能包含 Markdown Meta 信息，且同一正文允许发到不同渠道或同一渠道不同 target。

### 备选方案

- 方案 A：对整篇原始 Markdown 计算哈希。
- 方案 B：对去除 frontmatter 后的正文做标准化再计算哈希。
- 方案 C：使用模糊相似度算法近似去重。

### 最终选择

选择方案 B。

### 原因

- 精确满足需求。
- 规则透明，可调试，可重现。
- 实现成本低，不需要引入复杂文本相似度算法。

### 影响

- frontmatter 修改不会触发重新发布资格。
- 正文轻微变更会产生新哈希，允许再次发布。
- 需要把标准化规则固定下来并测试覆盖。

## 决策 5：模板引擎采用 Go 标准库 `text/template`

### 标题

MVP 模板渲染优先使用 Go 标准库 `text/template`，配合少量自定义函数。

### 背景

系统需要基础模板渲染能力，但当前不需要复杂 DSL 或可视化编辑。

### 备选方案

- 方案 A：`text/template`
- 方案 B：第三方模板引擎，如 `pongo2`
- 方案 C：自定义 DSL

### 最终选择

选择方案 A。

### 原因

- 无新增运行时依赖，维护成本最低。
- 已支持条件、循环、函数注入，足够覆盖 MVP。
- 安全边界更清晰，不会引入模板语言膨胀。

### 影响

- 模板表达能力有限。
- 若未来需要更复杂模板能力，可在渲染层做兼容升级。

## 决策 6：Markdown 处理引入 `goldmark`

### 标题

使用 `github.com/yuin/goldmark` 处理 Markdown 到 HTML 的转换。

### 背景

Telegram 推荐发送 HTML 或 MarkdownV2。MarkdownV2 需要大量转义，MVP 中直接对通用 Markdown 做可靠转换成本高。

### 备选方案

- 方案 A：直接把 Markdown 原文当纯文本发送。
- 方案 B：使用 `goldmark` 转 HTML，再清洗为 Telegram 支持标签。
- 方案 C：自研 Markdown 子集转换器。

### 最终选择

选择方案 B。

### 原因

- 格式保留优于纯文本。
- 可靠性高于自研转换器。
- `goldmark` 是 Go 社区常见选择。

### 影响

- 新增依赖：`github.com/yuin/goldmark`
- 需要补一层 Telegram HTML sanitizer。
- 某些复杂 Markdown 语法可能被降级或忽略。

## 决策 7：frontmatter 解析引入 YAML 解析能力

### 标题

使用 YAML 解析 frontmatter，并将其持久化为 JSON / 文本字段。

### 背景

Markdown 内容通常来自 Obsidian 等工具，frontmatter 默认是 YAML。系统需要读取 tags、summary 等元信息。

### 备选方案

- 方案 A：只保留 frontmatter 原文，不结构化解析。
- 方案 B：解析 YAML frontmatter，转换成 map 保存。
- 方案 C：要求用户上传纯 Markdown，不支持 frontmatter。

### 最终选择

选择方案 B。

### 原因

- 满足统一 Meta 管理需求。
- 对模板渲染和前端展示更友好。
- 技术实现直接，成本低。

### 影响

- 新增依赖：`gopkg.in/yaml.v3`
- 需要定义不合法 YAML 的错误返回。

## 决策 8：SQLite / PostgreSQL 通过 Gorm 同一数据模型兼容

### 标题

使用 Gorm 统一实体模型和 repository，数据库切换只通过配置完成。

### 背景

需求要求同一套业务逻辑兼容 SQLite 和 PostgreSQL，且业务层不能感知数据库差异。

### 备选方案

- 方案 A：分别实现两套 repository。
- 方案 B：Gorm 单模型 + 双 driver。
- 方案 C：只支持 PostgreSQL。

### 最终选择

选择方案 B。

### 原因

- 与技术栈约束一致。
- MVP 中查询模型较简单，Gorm 可以覆盖。
- 开发、测试、部署切换成本低。

### 影响

- 新增依赖：`gorm.io/gorm`
- 新增依赖：`gorm.io/driver/sqlite`
- 新增依赖：`gorm.io/driver/postgres`
- 避免使用明显依赖单一数据库的高级特性，如 PG 专属 JSON 查询、部分索引作为强依赖。

## 决策 9：保留任务表与投递表拆分

### 标题

使用 `publish_jobs` 和 `delivery_tasks` 两张表，而不是单表混合。

### 背景

业务需要展示“本次发布任务”的聚合结果，也需要展示“每个渠道/目标的投递明细”。

### 备选方案

- 方案 A：单表，每个目标一行，前端按 job id 聚合。
- 方案 B：任务表 + 投递表拆分。

### 最终选择

选择方案 B。

### 原因

- 聚合视图和明细视图边界清晰。
- 未来更容易扩展重试、补偿、统计。
- 更贴合“一个任务对应多个投递”的真实业务关系。

### 影响

- 表结构多一层，但可维护性更高。
- 需要维护 job 状态聚合逻辑。

## 决策 10：敏感配置只通过环境变量注入

### 标题

Bot Token 等敏感信息不入库，数据库仅保存 `secret_ref`。

### 背景

Telegram Bot Token 属于高敏感凭证。MVP 不实现专门的密钥管理系统。

### 备选方案

- 方案 A：数据库明文存储。
- 方案 B：数据库加密存储。
- 方案 C：环境变量存储，数据库保存引用名。

### 最终选择

选择方案 C。

### 原因

- 实现最简单且安全性高于明文入库。
- Docker / CI / 本地开发均易配置。
- 后续如接入 Vault，只需替换 `secret_ref` 解析器。

### 影响

- 部署时必须正确配置环境变量。
- 前端只能看到 `secret_ref`，不能回显实际 token。

## 决策 11：Telegram SDK 优先使用直接 HTTP 调用

### 标题

MVP 优先使用标准库 `net/http` 调用 Telegram Bot API，而不是引入重型 SDK。

### 背景

Telegram Bot API 很稳定，MVP 只需要 `sendMessage` 等少量接口。

### 备选方案

- 方案 A：直接 `net/http`
- 方案 B：引入第三方 Telegram SDK

### 最终选择

选择方案 A。

### 原因

- 减少依赖数量。
- 请求和错误处理逻辑更透明。
- 接口面较小，不值得为少量调用引入额外封装。

### 影响

- 需要自行维护 Telegram 请求/响应结构体。
- 若未来调用面扩大，再考虑引入 SDK。

## 决策 12：前端采用 Next.js App Router + shadcn/ui

### 标题

前端用 Next.js App Router 实现管理界面，UI 优先复用 shadcn/ui 组件。

### 背景

需求已经明确指定 Next.js、React、Tailwind CSS、shadcn/ui、lucide-react。

### 备选方案

- 方案 A：Next.js App Router
- 方案 B：Next.js Pages Router
- 方案 C：纯 React SPA

### 最终选择

选择方案 A。

### 原因

- 符合当前 Next.js 主流工程模式。
- 易于按页面和路由组织后台界面。
- 与 shadcn/ui 集成成熟。

### 影响

- 新增依赖：`next`
- 新增依赖：`react`
- 新增依赖：`tailwindcss`
- 新增依赖：`lucide-react`
- 新增依赖：shadcn/ui 初始化后按组件按需引入

## 决策 13：Docker 部署同时支持 SQLite 与 PostgreSQL 模式

### 标题

提供单套 Docker 化部署方案，并通过配置切换数据库模式。

### 背景

项目需要 Docker 化部署，且数据库可在 SQLite / PostgreSQL 间切换。

### 备选方案

- 方案 A：只支持 PostgreSQL 容器部署。
- 方案 B：支持 SQLite 单容器和 PostgreSQL 组合部署。

### 最终选择

选择方案 B。

### 原因

- 更贴近开发和轻量部署需求。
- SQLite 适合单机小规模场景。
- PostgreSQL 适合稍大规模和团队使用。

### 影响

- `docker-compose.yaml` 需要提供不同 profile 或示例。
- SQLite 需要挂载数据卷。

## 决策 14：补偿策略为应用启动时扫描未完成任务

### 标题

应用启动时扫描 `PENDING` / `PROCESSING` 的任务并标记或恢复执行。

### 背景

MVP 使用应用内后台 goroutine 执行任务。若应用异常退出，任务状态可能停留在处理中。

### 备选方案

- 方案 A：忽略异常中断任务。
- 方案 B：启动时扫描并补偿。

### 最终选择

选择方案 B。

### 原因

- 可以避免历史状态长期不一致。
- 不引入复杂调度系统也能做到最小可靠性闭环。

### 影响

- 启动阶段需要一段恢复逻辑。
- 恢复时要防止重复投递，可复用去重检查和幂等键。

## 计划引入依赖清单

### Backend 必需依赖

- `github.com/gin-gonic/gin`
  - 用途：REST API 框架
- `gorm.io/gorm`
  - 用途：ORM 与 repository 基础
- `gorm.io/driver/sqlite`
  - 用途：SQLite 支持
- `gorm.io/driver/postgres`
  - 用途：PostgreSQL 支持
- `github.com/yuin/goldmark`
  - 用途：Markdown 转 HTML
- `gopkg.in/yaml.v3`
  - 用途：frontmatter YAML 解析

### Frontend 必需依赖

- `next`
  - 用途：前端框架
- `react`
  - 用途：UI 基础
- `tailwindcss`
  - 用途：样式系统
- `lucide-react`
  - 用途：图标
- `shadcn/ui`
  - 用途：基础 UI 组件体系

## 关键假设

### 假设 1

MVP 默认单用户、单实例部署，不实现权限和协作。

影响：

- 不设计用户表和权限表。
- 审计字段中不包含操作者身份，只预留 `trigger_source`。

### 假设 2

同一个 Telegram Bot Token 可服务多个 Telegram 群组 target。

影响：

- 将 token 归属到 `channel_accounts`，target 只存 chat id 和目标参数。

### 假设 3

Markdown 上传以文本文件方式进行，MVP 不直接监听本地 Obsidian 目录。

影响：

- 先实现手动上传。
- 后续若要支持目录同步，可新增导入器模块，不影响当前 Content Model。

### 假设 4

frontmatter 采用 YAML 语法，且内容大小在单条 Telegram 消息可处理范围附近。

影响：

- MVP 先不做长文自动拆分发送。
- 超长内容按 `FAILED` 返回，并在后续版本补充分片策略。

### 假设 5

历史去重只阻止“已成功发送过”的重复正文，不阻止失败后的重试。

影响：

- 查询去重时只匹配 `SUCCESS` 投递。
- 用户仍然可以对失败记录执行重试。

### 假设 6

模板默认为系统内置模板，MVP 不提供模板管理后台。

影响：

- `template_name` 先只支持固定枚举值，如 `default-telegram`。
- 后续新增模板管理时无需改动发布模型。
