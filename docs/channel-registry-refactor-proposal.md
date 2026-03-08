# Channel Registry Refactor Proposal

本文是“下一步如何把新增渠道的改动面继续压缩”的方案文档，目标是让未来新增一个渠道时，尽量只新增文件，而不是修改现有核心代码。

本文不是当前必须立即实现的内容，而是面向下一阶段架构演进。

## 1. 当前架构仍然存在的剩余痛点

飞书接入完成后，当前架构已经比最初好很多，但还没有达到“新增渠道几乎只新增文件”的状态。

当前仍需要修改的现有代码主要有三类：

- 后端注册点
  - `backend/internal/api/router.go`
- 后端渠道常量
  - `backend/internal/domain/models.go`
- 前端渠道 schema 中央文件
  - `frontend/lib/channels.ts`

这说明当前已经做到：

- 业务编排层相对稳定
- 渲染层相对稳定
- 渠道差异主要集中在注册和元数据层

下一步要做的，不是继续把更多逻辑塞进现有 service，而是把“注册”和“元数据”也收敛掉。

## 2. 重构目标

目标不是做一个重型插件系统，而是做一个轻量、静态链接的 channel descriptor / registry 体系。

目标状态：

- 新增渠道时主要新增一个渠道目录
- 渠道自己声明自己的类型、target 类型、render mode、前后端 schema
- 后端只从 registry 读取渠道描述，不再手工注册一堆常量
- 前端从统一 descriptor 生成账号表单、target 表单和展示逻辑

换句话说，希望把：

- “这个渠道叫什么”
- “它有哪些 target 类型”
- “它如何校验账号和 target”
- “它前端需要哪些字段”

都变成 descriptor，而不是散落的条件分支。

## 3. 推荐的 descriptor 结构

### 3.1 后端 descriptor

建议为每个渠道定义：

```go
type ChannelDescriptor struct {
    ChannelType string
    TargetTypes []string
    RenderModes []string

    Driver Driver

    DefaultAccountConfig map[string]any
    DefaultTargetConfig  map[string]any
}
```

更进一步可以拆成：

```go
type ChannelDescriptor struct {
    Meta        ChannelMeta
    Driver      Driver
    AccountSpec AccountSpec
    TargetSpec  TargetSpec
}
```

其中：

- `Meta`
  - 渠道名、展示名、描述
- `AccountSpec`
  - 账号配置所需字段
- `TargetSpec`
  - target 配置所需字段

### 3.2 前端 descriptor

前端可以复用同构思想：

```ts
type ChannelDescriptor = {
  channelType: string
  label: string
  accountFields: FieldSchema[]
  targetFields: FieldSchema[]
  buildAccountPayload(formData: FormData): unknown
  buildTargetPayload(formData: FormData, account: ChannelAccount): unknown
  describeTarget(target: ChannelTarget): TargetDescription
}
```

当前 `frontend/lib/channels.ts` 已经有一部分这个方向，但还不够彻底。

## 4. 后端 registry 方案

### 4.1 当前状态

当前注册是：

- 在 `router.go` 手工 new driver
- 再传给 `channel.NewRegistry(...)`

问题：

- 每新增一个渠道都要改 router
- token provider 依赖组装也散落在入口

### 4.2 目标状态

建议引入：

```go
type DescriptorRegistry struct {
    descriptors map[string]ChannelDescriptor
}
```

以及统一装配函数：

```go
func BuildChannelRegistry(deps Dependencies) *DescriptorRegistry
```

由这个构建器统一：

- new Telegram driver
- new Feishu driver
- 未来 new X driver

这样 `router.go` 只依赖一个构建函数，而不是关心每个渠道的具体初始化。

### 4.3 更进一步的静态自注册

如果要进一步降低入口修改成本，可以让每个渠道包暴露：

```go
func Descriptor(deps Dependencies) ChannelDescriptor
```

然后在集中构建器中统一收集：

```go
descriptors := []ChannelDescriptor{
    telegram.Descriptor(deps),
    feishu.Descriptor(deps),
}
```

这仍然不是运行时插件系统，但已经足够把“新增渠道”的改动压缩到 descriptor 层。

## 5. 渠道类型常量的演进方案

当前 `domain/models.go` 里维护：

- `ChannelTypeTelegram`
- `ChannelTypeFeishu`
- `TargetTypeTelegramGrp`
- `TargetTypeTelegramTopic`
- `TargetTypeFeishuChat`
- `RenderModeTelegram`
- `RenderModeFeishuPost`

问题：

- 每新增渠道都得改这个文件
- 渠道元数据和领域模型耦合在一起

建议演进方向：

- 保留数据库字段仍然是 string
- 不再强依赖中心化常量
- descriptor 自己声明自己的 `channelType` / `targetTypes` / `renderModes`

保守做法：

- 先保留已有常量，兼容旧代码
- 新逻辑改成优先从 descriptor 读取
- 等所有调用点切换完成后，再考虑删除部分中心化常量

## 6. 前端 schema 化演进方案

### 6.1 当前状态

当前 `channels.ts` 已经集中了一部分：

- 渠道选项
- 默认 `secretRef`
- payload 映射
- target 展示解析

但表单字段本身仍部分写在组件里。

### 6.2 目标状态

建议引入声明式字段 schema：

```ts
type FieldSchema = {
  name: string
  label: string
  kind: "input" | "select" | "checkbox"
  placeholder?: string
  required?: boolean
  defaultValue?: string
  visible?: (ctx: FormContext) => boolean
}
```

然后每个渠道 descriptor 提供：

- `accountFields`
- `targetFields`

表单组件只做：

- 读取 descriptor
- 渲染字段
- 调用 payload builder

这样新增一个渠道时，前端主要是新增 schema，而不是改页面组件。

## 7. 分阶段迁移建议

### Phase 1：收敛后端 descriptor

目标：

- 新增 `ChannelDescriptor`
- 新增集中构建器
- `router.go` 改为只拿 registry

收益：

- 新增渠道不再直接改 router 逻辑

### Phase 2：收敛前端字段 schema

目标：

- `channels.ts` 升级为真正 schema + builder
- 表单组件只做通用渲染

收益：

- 新增渠道主要改 `channels.ts` 或独立 descriptor 文件

### Phase 3：弱化中心常量

目标：

- `channel_type` / `target_type` / `render_mode` 逐步由 descriptor 提供

收益：

- 新增渠道时更少改动 `domain/models.go`

### Phase 4：补统一开发工具

可选能力：

- 自动生成渠道骨架
- 渠道开发 checklist 模板
- 渠道测试模板

## 8. 不建议当前就做的事

为了避免过度设计，当前不建议立刻做：

- 运行时动态插件加载
- 独立插件包热插拔
- 前后端 descriptor 完全共享协议生成
- 元编程式代码生成器

原因：

- 当前渠道数量还少
- 复杂度会明显超过收益
- 先把静态 descriptor / registry 做好已经足够

## 9. 推荐实施顺序

如果下一步要继续把新增渠道的改动压缩到最小，建议按这个顺序做：

1. 后端引入 `ChannelDescriptor`
2. 后端引入集中 registry builder
3. 让 Telegram / Feishu 迁到 descriptor 模式
4. 前端引入 `FieldSchema`
5. 让 Telegram / Feishu 前端配置迁到 descriptor 模式
6. 再评估是否还需要进一步移除中心常量

## 10. 成功标准

完成这轮重构后，新增一个新渠道时应尽量满足：

- 后端主要新增一个渠道目录
- 前端主要新增一个 descriptor / schema 文件
- 不修改发布编排核心
- 不修改历史核心
- 不修改去重核心
- 只在极少数入口做注册性修改，或完全由集中 builder 接管

这就是当前阶段“最小最小幅度不修改现有代码，而是以新增为主”的现实可达目标。
