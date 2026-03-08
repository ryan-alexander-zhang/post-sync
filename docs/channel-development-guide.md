# Channel Development Guide

本文面向后续开发者，说明在当前代码结构下如何新增一个渠道，并尽量把改动控制在“新增为主，少量注册修改”的范围内。

适用范围：

- 当前仓库的现有实现
- 当前已经支持的渠道：Telegram、Feishu
- 当前前端渠道表单实现：集中 schema / payload 映射

不覆盖：

- 更激进的插件化 / descriptor 自动注册方案
- 分布式队列、多租户、模板 DSL 等未来能力

## 1. 当前架构的渠道扩展原则

当前新增渠道的目标不是“零修改现有代码”，而是：

- 不修改发布编排核心
- 不修改去重核心
- 不修改历史模型
- 主要通过新增 driver、token provider、前端 schema 来扩展

当前已经收敛好的边界：

- 后端通过 `ChannelDriver` 抽象渠道差异
- target 规范化由各渠道 driver 负责
- 账号鉴权与 token 获取由各渠道 driver / token provider 负责
- 模板层只产出中性 Markdown
- 各渠道在自身 driver 内完成最终格式转换
- 前端账号/target 表单通过 `frontend/lib/channels.ts` 做集中 schema / payload 映射

这意味着新增一个渠道时，优先考虑“新增一个完整渠道目录”，而不是在业务层堆条件分支。

## 2. 当前代码中的关键扩展点

### 2.1 后端入口

新增渠道最常用的几个文件：

- `backend/internal/channel/driver.go`
  - 渠道抽象接口
- `backend/internal/api/router.go`
  - 渠道 driver 注册点
- `backend/internal/service/channel_service.go`
  - 调用 driver 做账号校验和 target 规范化
- `backend/internal/service/publish_service.go`
  - 调用 driver 做最终发送
- `backend/internal/domain/models.go`
  - 渠道常量、target 类型、render mode 常量

### 2.2 前端入口

新增渠道最常用的几个文件：

- `frontend/lib/channels.ts`
  - 渠道元数据、默认值、payload 构造、展示解析
- `frontend/components/forms/channel-account-form.tsx`
  - 账号表单通用渲染入口
- `frontend/components/forms/channel-target-form.tsx`
  - target 表单通用渲染入口
- `frontend/app/channels/page.tsx`
  - 渠道列表展示

当前前端已经把大部分渠道差异收敛到了 `channels.ts`，所以新增渠道时应优先修改该文件，而不是直接改页面。

## 3. 新增一个渠道的推荐开发流程

建议按以下顺序进行，避免把多个边界混在一个提交里。

### 第一步：确认渠道边界

在动手前先明确：

- 渠道类型名是什么
- target 类型有哪些
- 发送身份是什么
- 是否需要动态 token
- 首版支持哪些消息类型
- 是否需要素材上传
- 去重粒度是否沿用 `channel_type + target_key + body_hash`

如果首版边界不收敛，代码很容易提前过度设计。

### 第二步：新增渠道后端 driver

建议新增目录：

```text
backend/internal/channel/<channel_name>/
  driver.go
  driver_test.go
  token_provider.go      // 可选
  payload.go             // 可选
```

实现内容通常包括：

- `Type()`
- `ValidateAccount(...)`
- `NormalizeTarget(...)`
- `Render(...)`
- `Send(...)`

要求：

- 账号校验只校验该渠道自己的必要条件
- target 规范化必须输出稳定的 `target_type` 和 `target_key`
- `Send(...)` 里不要依赖业务层分支

### 第三步：注册渠道

目前仍需要改一个注册点：

- `backend/internal/api/router.go`

在这里把新 driver 放进 `channel.NewRegistry(...)`。

如果新渠道需要 token provider，也在这里组装依赖。

### 第四步：补领域常量

目前仍需要改：

- `backend/internal/domain/models.go`

新增：

- `ChannelTypeXXX`
- `TargetTypeXXX`
- `RenderModeXXX`

要求：

- 常量命名和数据库取值保持稳定
- 一旦落库，不要轻易改名

### 第五步：补前端渠道 schema

优先修改：

- `frontend/lib/channels.ts`

通常需要补：

- 渠道选项
- 默认 `secretRef`
- account payload 构造
- target payload 构造
- target 展示解析

如果 schema 足够表达，就不要改页面组件。

只有在该渠道字段形态明显不同、当前通用 UI 承载不了时，才去改：

- `channel-account-form.tsx`
- `channel-target-form.tsx`
- `channels/page.tsx`

### 第六步：补文档

至少同步：

- `design.md`
- `decision.md`
- `README.md`

如果用户另外放了平台原始文档，建议像飞书一样提炼成轻量参考文件，而不是每次重新加载大文档。

## 4. 当前新增渠道通常需要改哪些文件

### 4.1 后端最小改动面

新增文件：

- `backend/internal/channel/<channel_name>/driver.go`
- `backend/internal/channel/<channel_name>/driver_test.go`
- `backend/internal/channel/<channel_name>/token_provider.go`（如需要）

通常需要修改的现有文件：

- `backend/internal/api/router.go`
- `backend/internal/domain/models.go`

理想情况下，不需要修改：

- `backend/internal/service/publish_service.go`
- `backend/internal/service/content_service.go`
- `backend/internal/repository/*`

### 4.2 前端最小改动面

新增文件通常不是必须的，当前以集中 schema 为主。

通常需要修改：

- `frontend/lib/channels.ts`

可能需要少量修改：

- `frontend/components/forms/channel-account-form.tsx`
- `frontend/components/forms/channel-target-form.tsx`
- `frontend/app/channels/page.tsx`

理想情况下，不需要修改：

- `frontend/app/publish/new/page.tsx`
- `frontend/components/forms/publish-job-form.tsx`

## 5. 渠道开发 checklist

### 5.1 后端 checklist

- 定义 `channel_type`
- 定义 `target_type`
- 定义 `render_mode`
- 实现 `ValidateAccount`
- 实现 `NormalizeTarget`
- 实现 `Render`
- 实现 `Send`
- 如需 token，补 `token_provider`
- 补单元测试
- 注册到 router

### 5.2 前端 checklist

- 补渠道 option
- 补账号默认值
- 补 account payload 映射
- 补 target payload 映射
- 补 target 展示描述
- 检查 `/channels`
- 检查 `/publish/new`

### 5.3 文档 checklist

- 设计文档补“当前实现内容”
- 决策文档补关键 ADR
- README 补环境变量与使用流程

## 6. 当前实现中要避免的反模式

### 6.1 不要在 `PublishService` 里新增渠道分支

错误方向：

- `if channelType == "feishu" { ... }`
- `if channelType == "x" { ... }`

这样会把发布核心重新耦合回具体渠道。

正确方向：

- 把差异压到 driver 内

### 6.2 不要在模板层输出渠道专属格式

错误方向：

- 模板直接输出 Telegram HTML
- 模板直接输出飞书 card JSON

正确方向：

- 模板输出统一 Markdown 文本
- 各渠道自行转换最终 payload

### 6.3 不要在前端多个页面散落渠道判断

错误方向：

- 每个表单、列表、详情页都单独加一份 `if channelType === ...`

正确方向：

- 尽量把渠道配置收敛到 `frontend/lib/channels.ts`

## 7. 提交建议

新增渠道时，推荐按以下提交顺序拆小 commit：

1. 新增渠道 driver 抽象实现
2. 新增 token provider / auth 支持
3. 注册渠道与常量
4. 前端渠道 schema / 表单支持
5. 文档更新

示例 commit message：

- `add xxx channel driver`
- `add xxx token provider`
- `register xxx channel`
- `add xxx channel frontend config`
- `document xxx channel support`

## 8. 当前新增渠道的工作量判断

以当前结构看：

- 新增一个和 Telegram / Feishu 复杂度接近的渠道
  - 中等偏小
- 新增一个需要 OAuth、多目标类型、复杂富文本的渠道
  - 中等

相比飞书接入前，当前最大的进步是：

- 发布核心基本不需要改
- 去重和历史模型不需要改
- 模板层也不需要改

剩余仍需要收敛的地方：

- driver 注册仍在 `router.go`
- 常量仍在 `domain/models.go`
- 前端 schema 还不是完全声明式字段渲染

这些将在下一份重构方案中处理。
