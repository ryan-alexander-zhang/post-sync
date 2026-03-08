# 内容分发器 MVP 决策记录

本文记录当前仓库首版 MVP 方案的关键架构决策、依赖引入计划与保守假设。由于仓库当前为空仓库，本文件同时承担首版 ADR 集合的角色。

## 决策 1：发布流程采用轻量异步任务模型

### 标题

使用“创建任务后后台执行”的轻量异步模型，而不是请求全程同步阻塞。

### 背景

一次发布可能会向多个 Telegram / Feishu target 发送消息，外部 API 时延、网络抖动、限流和失败重试都会拉长请求时间。若完全同步处理，前端接口超时和用户体验都会变差；若引入完整消息队列，又超出 MVP 需要。

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

MVP 需要支持多渠道并行发送和查看每个渠道状态，当前已实现 Telegram 与 Feishu，目标规模仍可预期较小。

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

## 决策 3A：Telegram Topic 作为独立 target 建模

### 标题

将 Telegram 群组 root 和群组下每个 topic 都建模为独立 `ChannelTarget`。

### 背景

Telegram 群组可能启用 Topics。同一群组下需要配置多个 topic，并且同一正文允许发到同一群组的不同 topic。

### 备选方案

- 方案 A：一个 group target 下挂 topic 列表，发布时再二次选择 topic。
- 方案 B：每个 topic 作为独立 target，group root 也是独立 target。

### 最终选择

选择方案 B。

### 原因

- 复用现有 target、delivery、去重、历史模型，不引入第二层目标选择。
- 同一群组下多个 topic 可以自然并行发送。
- 去重粒度可直接落在 `target_key`。

### 影响

- `telegram_group` target 的 `target_key=chat_id`
- `telegram_topic` target 的 `target_key=chat_id:topic:topic_id`
- `config_json` 需要保存 `chatId`、`topicId`、`topicName`

## 决策 3B：渠道驱动负责目标规范化与发送鉴权

### 标题

把目标规范化、凭证解析和渠道发送细节下沉到 `ChannelDriver`，而不是放在 `ChannelService` / `PublishService`。

### 背景

在仅有 Telegram 时，业务层里直接读 `os.Getenv(secretRef)`、直接分支处理 Telegram target 还能工作；一旦加入 Feishu 这类需要 `app_id + app_secret -> tenant_access_token` 的渠道，这种写法会快速失控。

### 备选方案

- 方案 A：继续在业务层增加渠道分支。
- 方案 B：把 target 规范化和发送鉴权下沉到 driver。
- 方案 C：引入外部插件系统。

### 最终选择

选择方案 B。

### 原因

- 新增渠道时不需要修改发布编排核心。
- 驱动可以自由实现静态 token、动态换 token、目标 key 规范化。
- 前后端都可以围绕渠道 schema 扩展，而不是围绕硬编码平台分支扩展。

### 影响

- `ChannelDriver` 接口从“只校验 + 发送”扩展为“校验 + NormalizeTarget + Send”。
- `PublishService` 不再直接读取环境变量。
- `ChannelService` 不再内置 Telegram 专属 target 规范化逻辑。

## 决策 3C：Feishu 首版采用应用凭证换取 tenant_access_token，静态 token 作为调试兜底

### 标题

Feishu 默认使用 `app_id + app_secret` 获取并缓存 `tenant_access_token`，允许本地通过静态 token 环境变量覆盖。

### 背景

新增 `docs/feishu-token.md` 后，仓库已经具备实现标准鉴权流程所需的信息。若继续只支持手动提供 `tenant_access_token`，会增加运维负担，也不利于后续新增类似鉴权模型的渠道。

### 备选方案

- 方案 A：只接受静态 `FEISHU_TENANT_ACCESS_TOKEN`
- 方案 B：默认走应用凭证换 token，允许静态 token 覆盖
- 方案 C：把 token 刷新完全交给外部代理服务

### 最终选择

选择方案 B。

### 原因

- 更接近飞书的标准用法。
- 能验证当前渠道抽象是否足以承载“动态 token 获取”。
- 本地调试仍然保留快捷路径。

### 影响

- 新增 `FeishuTokenProvider`
- 无需引入第三方 SDK，仍使用标准库 `net/http`
- Feishu 账号配置除 `secret_ref` 外，还需要 `appIdEnv`，可选 `tokenEnv`

## 决策 3D：Feishu 拆分为 Enterprise Feishu 与 Personal Feishu 两类渠道

### 标题

保留 `feishu` 作为企业版飞书渠道，新增 `personal_feishu` 作为基于 webhook 的个人 / 自定义机器人渠道。

### 背景

当前仓库原先把 `feishu` 默认实现为 `app_id + app_secret -> tenant_access_token -> im/v1/messages`。这实际上对应的是飞书企业应用机器人，而不是群自定义机器人 webhook。用户明确要求区分“企业 Feishu Channel”与“Personal Feishu Channel”，避免在产品和配置层混淆两类完全不同的发送模型。

### 备选方案

- 方案 A：继续沿用单一 `feishu` 渠道，在账号配置里用字段区分企业版与 webhook 版。
- 方案 B：保留 `feishu` 作为企业版，新增独立 `personal_feishu` 渠道。
- 方案 C：把原 `feishu` 直接重命名并做数据迁移。

### 最终选择

选择方案 B。

### 原因

- 对现有数据最安全：历史 `feishu` 账号无需迁移类型值。
- 前后端语义更清晰：企业版和个人版的鉴权、target 规范化、发送接口都不同。
- 发布编排层无需感知渠道细节，仍只依赖 `ChannelDriver` 抽象。

### 影响

- `feishu` 默认表示 Enterprise Feishu，继续使用 `appIdEnv + secretRef + token provider`。
- 新增 `personal_feishu`，账号配置改为敏感信息环境变量引用，而不是直接保存 webhook URL。
- `ChannelService` 不再强制所有渠道账号都必须提供 `secretRef`，改为由各自 driver 校验。

## 决策 3E：Personal Feishu 使用环境变量引用保存 webhook，并启用签名校验

### 标题

`personal_feishu` 不直接保存 webhook URL 到前端可见配置中，而是通过环境变量引用保存 webhook 地址，并在发送时按飞书自定义机器人签名规则补 `timestamp + sign`。

### 背景

飞书企业应用机器人走 `im/v1/messages`，而个人 / 自定义机器人走 webhook。根据 `docs/feishu-custom-robot.md`，自定义机器人可以启用签名校验；同时 webhook 地址本身属于敏感信息，不应返回给前端界面。

### 备选方案

- 方案 A：继续在账号配置中直接保存 webhook URL。
- 方案 B：通过环境变量引用保存 webhook URL，并为签名秘钥也使用环境变量引用。
- 方案 C：把 webhook URL 直接存为 `target_key`。

### 最终选择

选择方案 B，并且对 `target_key` 使用 webhook 引用的稳定哈希。

### 原因

- 符合自定义机器人安全配置要求，发送端可以补签名。
- webhook URL 和签名秘钥都不需要返回给前端。
- 哈希化 `target_key` 可以继续支持去重，同时避免在界面和历史中直接暴露 webhook 地址。

### 影响

- `personal_feishu.secretRef` 改为 webhook URL 的环境变量名。
- `personal_feishu.config.signSecretRef` 保存签名秘钥的环境变量名。
- `target_key` 使用 webhook 引用的稳定哈希，继续支撑去重和历史展示。
- API 返回前端时需要对 legacy `webhookUrl` / `signSecret` 做脱敏处理。

## 决策 3F：Personal Feishu webhook 使用自定义机器人专用 post 富文本协议

### 标题

`personal_feishu` 使用 `docs/feishu-custom-robot.md` 定义的 webhook `msg_type=post` 富文本协议，但实现必须独立于企业 Feishu 的 `post + md` 结构。

### 背景

Personal Feishu 对应的是飞书自定义机器人 webhook，而不是企业应用消息接口。`docs/feishu-custom-robot.md` 的“发送富文本消息”章节已经定义了 webhook 专用 `post` 协议，其 `content.post.zh_cn.content` 由段落数组组成，支持 `text`、`a`、`at`、`img` 等节点。用户当前明确要求支持超链接标签，且强调不能复用企业 Feishu 的消息转换逻辑。

### 备选方案

- 方案 A：继续保持 webhook `text`，只保留纯文本可读性。
- 方案 B：复用企业 Feishu 的 `post + md` 结构。
- 方案 C：按自定义机器人 `post` 协议单独把 Markdown 转成 webhook 段落节点。

### 最终选择

选择方案 C。

### 原因

- 与 `docs/feishu-custom-robot.md` 的自定义机器人协议一致。
- 企业 Feishu 的 `md` 节点结构不适用于 webhook，自定义机器人必须单独转换。
- Markdown 链接可以稳定映射为 webhook `a` 节点，满足当前“支持超链接标签”的核心需求。

### 关键假设

- 当前模板输出主要仍是 Markdown 文本，可在 driver 内做轻量协议转换。
- 首版只保证段落文本和 Markdown 链接 `[text](url)` -> `a` 标签稳定可用。
- 更复杂的 Markdown 语法不在这一轮强制支持范围内。

### 影响

- `personal_feishu` 渲染模式改为独立的 `personal_feishu_post`。
- `PersonalDriver` 发送时构造 webhook `content.post.zh_cn.title/content`，不再发送纯 `text` 消息。
- `PersonalDriver` 内需要维护一套独立的 Markdown -> webhook 节点转换逻辑，至少覆盖段落与超链接标签。

## 决策 4：去重算法基于去除 Meta 后的标准化正文哈希

### 标题

上传去重与发布去重统一使用“移除 frontmatter 后的标准化正文”的 SHA-256。

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

- 上传阶段若发现相同 `body_hash` 已存在，则拒绝创建新的 `content` 记录。
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

## 决策 5A：模板层只产出中性 Markdown，渠道格式转换下沉到 driver

### 标题

模板执行阶段只输出中性 Markdown 文本，不再直接输出 Telegram 或其他渠道专属格式。

### 背景

在只有 Telegram 时，把模板执行、Markdown 转 HTML、Telegram HTML 清洗串在同一个渲染层里还能工作；但接入 Feishu 后，这种设计会把 Telegram 专属输出泄漏给所有新渠道。

### 备选方案

- 方案 A：模板层继续直接产出 Telegram HTML，再由其他渠道兼容处理。
- 方案 B：模板层只产出中性 Markdown，各渠道在各自 driver 内完成最终格式转换。
- 方案 C：为每个渠道维护完全独立的模板体系。

### 最终选择

选择方案 B。

### 原因

- 避免渠道间的格式耦合。
- 保持统一模板上下文和统一模板能力。
- 新增渠道时只需要关注自身的最终渲染格式。

### 影响

- `TemplateRenderer` 从“模板执行 + Telegram 格式化”收敛为“模板执行”。
- Telegram 在 driver 内执行 Markdown -> HTML -> Telegram 支持标签子集的转换。
- Feishu 等后续渠道可以直接把模板结果封装成自身 payload，而不需要兼容 Telegram HTML。

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

## 决策 6A：Feishu MVP 消息类型采用 `post + md`

### 标题

Feishu 渠道在 MVP 阶段采用 `msg_type=post`，并把统一模板产出的 Markdown 封装进 `md` 节点；当前仍不实现 `interactive`、图片上传和手工细粒度富文本节点编排。

### 背景

飞书消息接口支持 `text`、`post`、`interactive` 等多种格式。当前系统的统一模板已经产出 Markdown，而 `docs/feishu-message-content.md` 明确说明 `post` 支持 `md` 标签，可以直接承载 Markdown 子集。继续停留在 `text` 会让标题、列表、引用、代码块等结构全部退化成普通文本，实际效果明显弱于 Telegram。

### 备选方案

- 方案 A：继续使用 `text`
- 方案 B：使用 `post`，并把正文包进 `md` 节点
- 方案 C：直接实现 `post` 的细粒度节点映射
- 方案 D：直接上 card

### 最终选择

选择方案 B。

### 原因

- 与当前统一模板结果天然兼容，不需要重写模板层。
- 比 `text` 明显更接近用户对“Markdown 富文本”的预期。
- 比手工拆 AST/节点树更稳，能先复用飞书官方支持的 Markdown 子集。
- 保持平台差异继续封装在 Feishu driver 内，不扩散到发布编排层。

### 影响

- Feishu driver 当前发送 `msg_type=post`，`content` 结构为 `{"zh_cn":{"title":"...","content":[[{"tag":"md","text":"..."}]]}}`。
- `SendRequest` 补充了通用 `Title` 字段，为后续其他富文本渠道复用标题能力预留统一入口。
- 为避免标题重复，Feishu driver 会剥离默认模板中与 `post.title` 重复的首行 `# title`。
- 当前仍不支持飞书卡片、图片上传、文件上传和自定义节点级样式控制。

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

## 决策 11A：前端渠道表单采用集中 schema / 映射驱动

### 标题

前端渠道配置不再在页面组件里硬编码 Telegram 逻辑，而是集中在 `frontend/lib/channels.ts` 管理渠道 schema 与 payload 映射。

### 背景

在只有 Telegram 时，账号表单、target 表单和列表展示都写死 Telegram 字段还能接受；但新增 Feishu 后，继续在页面里堆叠 `if channelType === ...` 会迅速失控。

### 备选方案

- 方案 A：继续在各页面组件内按渠道堆条件分支。
- 方案 B：把渠道选项、配置默认值、payload 构造、展示解析集中到一个前端渠道配置模块。
- 方案 C：引入完整动态表单引擎。

### 最终选择

选择方案 B。

### 原因

- 保持当前 MVP 复杂度可控。
- 相比完整动态表单引擎，实现成本更低。
- 新增渠道时主要改一个集中模块，而不是散改多个页面。

### 影响

- 新增 `frontend/lib/channels.ts`
- 账号表单、target 表单、渠道列表统一复用该模块
- 后续新增第三个渠道时，前端主要在 schema / 映射层补充配置

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
- 同一群组下多个 topic 通过多个独立 target 表达，而不是在单个 target 内嵌套数组。

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

## 决策 15：Obsidian QuickAdd 通过本地 alias 映射驱动自动发布

### 标题

Obsidian 集成采用 “frontmatter 写业务 alias + QuickAdd 脚本本地映射到 target id” 的方式，而不是把后端 target id 直接写进笔记。

### 背景

用户需要在 Obsidian 中对当前打开 Markdown 直接触发发布，同时希望利用 Obsidian frontmatter / Properties 面板维护发布元信息。当前后端 `channel_targets` 只有数据库 `id`，没有稳定 alias 字段；如果直接把 target id 写入笔记，内容可读性和可迁移性都会较差。

### 备选方案

- 方案 A：frontmatter 直接写后端 target id
- 方案 B：frontmatter 写业务 alias，QuickAdd 脚本读取 vault 本地 JSON 做 alias -> targetIds 映射
- 方案 C：frontmatter 写 alias，脚本运行时调用后端接口按名称模糊查找

### 最终选择

选择方案 B。

### 原因

- 笔记可读性最好，适合在 Obsidian Properties 中直接编辑。
- target 重建或迁移环境时，只需要改本地映射文件，不需要批量改历史笔记。
- 不要求后端新增 alias 字段，也不依赖名称匹配的稳定性。

### 影响

- vault 里需要维护一份 `target-aliases.json`。
- QuickAdd 脚本需要负责 alias 解析、上传、创建 publish job。
- 发布控制字段统一使用顶层 `post_` 前缀：
  - `post_title`
  - `post_publish`
  - `post_targets`
  - `post_template`

## 决策 16：QuickAdd 上传前剥离发布控制字段，并把 `post_title` 映射为 `title`

### 标题

QuickAdd 脚本在上传 Markdown 到后端前，去除 frontmatter 中的发布控制字段，并把 `post_title` 写入上传内容的 `title` 字段。

### 背景

后端当前的 Markdown 解析逻辑以 frontmatter 中的 `title` 作为内容标题来源，而用户在 Obsidian 中明确希望用 `post_title` 维护发布标题。如果直接把原始笔记上传，后端不会自动识别 `post_title`。

### 备选方案

- 方案 A：修改后端解析器，让它优先识别 `post_title`
- 方案 B：由 QuickAdd 脚本在上传前转换 frontmatter，兼容现有后端

### 最终选择

选择方案 B。

### 原因

- 不改动现有后端接口和内容模型。
- 风险更小，影响范围只在 Obsidian 集成脚本。
- `post_publish`、`post_targets`、`post_template` 属于发布控制字段，本身也不应进入上传后的内容元数据。

### 影响

- QuickAdd 上传的是“变换后的 Markdown 内容”，不是 vault 中原始文件的字节流。
- 上传后的 `frontmatter_json` 不包含发布控制字段。
- `post_title` 仍可保留在原笔记中供 Obsidian 使用，但后端最终读取的是脚本写入的 `title`。

## 决策 17：CORS 改为环境变量驱动的来源白名单

### 标题

后端 CORS 不再硬编码单一 `http://localhost:3000`，改为通过 `CORS_ALLOW_ORIGINS` 环境变量配置允许来源列表。

### 背景

项目最初只服务本地前端页面，因此 CORS 中间件只允许 `http://localhost:3000`。在引入 Obsidian QuickAdd 后，桌面端用户脚本会以 Obsidian 自身的 origin 发起请求，固定来源会导致浏览器环境直接报 `Failed to fetch`，请求无法到达业务逻辑。

### 备选方案

- 方案 A：通过 `CORS_ALLOW_ORIGINS` 配置来源白名单
- 方案 B：直接放开为 `*`
- 方案 C：继续硬编码多个来源

### 最终选择

选择方案 A。

### 原因

- 同时满足本地前端与 Obsidian QuickAdd 的访问需求。
- 比 `*` 更保守，不会把所有来源都放开。
- 后续新增其它本地工具来源时只改环境变量，不改代码。

### 影响

- 新增环境变量 `CORS_ALLOW_ORIGINS`
- 配置格式为逗号分隔，例如 `http://localhost:3000,app://obsidian.md`
- CORS 中间件按请求 `Origin` 命中白名单后再回写允许头
