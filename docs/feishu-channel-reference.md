# Feishu Channel Reference

该文件是基于以下原始文档提取的轻量参考，供后续实现飞书渠道时直接使用：

- `docs/feishu-message-send.md`
- `docs/feishu-message-content.md`

目标是减少后续反复加载大文档的成本，只保留当前项目接入飞书渠道最关键的信息。

## 1. MVP 接入边界

当前项目如果新增飞书渠道，建议先只支持以下最小能力：

- 渠道类型：`feishu`
- 目标类型：`feishu_chat`
- 发送身份：应用机器人
- 接收者类型：群聊 `chat_id`
- 消息类型：`text`
- 可选扩展：`post`

当前不建议首版实现：

- 自定义机器人 webhook
- 单聊发送
- 卡片 `interactive`
- 图片、文件、音视频
- 回复消息、编辑消息、话题回复

原因：

- 当前仓库的内容分发链路更适合先做“纯文本/富文本直发”
- 群聊 `chat_id` 是最接近当前 Telegram group target 的模型
- 卡片、素材上传、模板变量类型校验会显著增加 MVP 复杂度

## 2. 发送前提

根据原文档，飞书发送消息接口的前提条件是：

- 应用需要开启机器人能力
- 能力开启后需要发布版本才能生效
- 给群组发送消息时，机器人必须已经在群里
- 机器人在群内需要有发言权限

接口限制：

- 同一群组内机器人共享 5 QPS
- 发送消息接口本身还有平台频率限制：1000 次/分钟、50 次/秒
- 原始文档明确说明：`im/v1/messages` 仅支持开发者后台创建的应用机器人，不支持群自定义机器人

## 3. 鉴权与配置

发送接口使用：

- Header: `Authorization: Bearer <tenant_access_token>`
- Header: `Content-Type: application/json; charset=utf-8`

对当前项目的配置映射建议：

- `ChannelAccount.channelType = "feishu"`
- `ChannelAccount.secretRef = FEISHU_APP_SECRET`
- `ChannelAccount.configJson` 保存非敏感配置：
  - `appIdEnv`
  - 可选 `tokenEnv`
  - 可选 `baseUrl`

更稳妥的本地环境变量方案：

- `FEISHU_APP_ID`
- `FEISHU_APP_SECRET`
- `FEISHU_TENANT_ACCESS_TOKEN`

说明：

- 当前仓库已补充 `docs/feishu-auth-reference.md`
- 飞书更适合通过 `app_id + app_secret` 获取并缓存 `tenant_access_token`
- 本地调试时也可以允许显式提供 `FEISHU_TENANT_ACCESS_TOKEN`

结论：

- 飞书渠道建议直接按 token provider 方式实现
- 这样后续新增需要自动换 token 的渠道时，不需要改动发布编排层

## 4. Target 模型建议

首版目标模型建议只支持群聊：

- `ChannelTarget.targetType = "feishu_chat"`
- `ChannelTarget.targetKey = <chat_id>`
- `ChannelTarget.targetName = 群名或人工备注`
- `ChannelTarget.configJson` 可选字段：
  - `receiveIdType: "chat_id"`
  - `chatId: "<oc_xxx>"`

原因：

- 原始发送接口支持多种 `receive_id_type`，但群聊最适合当前项目
- 对接群聊只需要 `chat_id`
- 去重逻辑可直接复用现有 `channel_type + target_key + body_hash`

后续扩展时再考虑：

- `feishu_user_open_id`
- `feishu_user_id`
- `feishu_email`

## 5. 发送接口

发送接口：

- `POST https://open.feishu.cn/open-apis/im/v1/messages`

查询参数：

- `receive_id_type=chat_id`

最小请求体：

```json
{
  "receive_id": "oc_xxx",
  "msg_type": "text",
  "content": "{\"text\":\"test content\"}",
  "uuid": "dedup-uuid-optional"
}
```

当前项目映射建议：

- `receive_id` <- `ChannelTarget.targetKey`
- `msg_type` <- 渠道渲染结果决定，首版固定 `text`
- `content` <- 渲染后的飞书消息 JSON 再序列化为字符串
- `uuid` <- 可复用当前 delivery 的幂等键，建议截断到 50 字符以内

## 6. 文本消息格式

飞书 `text` 消息的 `content` JSON 结构是：

```json
{
  "text": "plain text"
}
```

支持的能力：

- `\n` 换行
- `@` 用户
- `@all`
- 部分样式标签
- Markdown 风格超链接 `[text](url)`

适合作为当前项目首版的原因：

- 最容易从 Markdown 正文降级得到稳定输出
- 不需要多语言结构
- 不需要构建复杂节点树

需要注意：

- `content` 字段本身是字符串，所以需要先构造 JSON，再整体序列化
- 文本消息请求体最大 150 KB

## 7. 富文本 post 结构

飞书 `post` 消息使用多语言对象，最小结构通常类似：

```json
{
  "zh_cn": {
    "title": "标题",
    "content": [
      [
        { "tag": "text", "text": "第一段" }
      ],
      [
        { "tag": "md", "text": "**bold**\n- item" }
      ]
    ]
  }
}
```

关键点：

- `post` 的 `content` 是段落数组
- 每个段落是 node 数组
- `md` 节点可承载 Markdown
- 至少需要一种语言，例如 `zh_cn`
- `post` / 卡片请求体最大 30 KB

对当前项目的结论：

- 飞书渠道首版可以只做 `text`
- 第二阶段再考虑 `post`
- 如果实现 `post`，最简单路径是把 Markdown body 映射为一个或多个 `md` 节点，而不是手工拆成细粒度 `text/a/at/...` 节点

## 8. 渲染链路建议

飞书渠道建议新增独立渲染模式：

- `RenderMode = "feishu_text"` 或 `RenderMode = "feishu_post"`

推荐首版渲染策略：

1. 使用现有统一模板上下文生成中间文本
2. 渠道驱动将文本封装为：
   - `msg_type = "text"`
   - `content = {"text":"..."}`
3. 发送前再做 JSON 序列化

推荐的保守转换规则：

- 保留换行
- 保留 `#tag` 文本
- 不主动输出 HTML
- 不尝试把 Telegram HTML 直接复用到飞书

原因：

- 飞书 `text` 支持的是文本 + 部分样式语法，不是 Telegram HTML
- 当前项目的 Telegram 渲染链路不能直接复用到飞书发送 payload

## 9. 错误码与排查重点

原始文档中，首版最值得优先处理的错误包括：

- `230002`
  - 机器人不在群组中
- `230006`
  - 应用未启用机器人能力
- `230013`
  - 目标用户或单聊对象不在机器人可用范围内
- `230020`
  - 应用权限不足
- `230031`
  - 消息内容超长
- `230099`
  - 内容构造失败，常见于卡片/富文本格式错误

对当前项目建议的错误映射：

- 配置问题 -> `CONFIG_ERROR`
- 渲染失败 -> `RENDER_ERROR`
- 飞书 API 返回失败 -> `CHANNEL_SEND_FAILED`

建议至少记录这些字段：

- HTTP status
- Feishu `code`
- Feishu `msg`
- `log_id`（如果响应里有）

## 10. 对当前代码库的落地建议

后端最小改动点：

- 新增 `internal/channel/feishu/driver.go`
- 在 channel registry 注册 `feishu`
- 在 `channel_service` 中新增 `feishu` account / target 校验
- 新增飞书 target config 解析器
- 新增飞书发送请求结构与响应结构

前端最小改动点：

- account 表单允许选择 `feishu`
- target 表单在 `feishu` 下只采集：
  - `targetName`
  - `chatId`
- channels 列表支持 `feishu_chat` 展示

推荐的首版数据映射：

- `ChannelAccount.secretRef`:
  - 先存 `FEISHU_APP_SECRET`
- `ChannelAccount.configJson`:
  - `{"appIdEnv":"FEISHU_APP_ID","tokenEnv":"FEISHU_TENANT_ACCESS_TOKEN"}`
- `ChannelTarget.targetKey`:
  - `oc_xxx`
- `ChannelTarget.configJson`:
  - `{"receiveIdType":"chat_id","chatId":"oc_xxx"}`

## 11. 当前文档缺口

本仓库现有飞书文档只覆盖了：

- 发消息接口
- 消息内容结构

还缺少后续实现一定会用到的文档：

- 获取 `tenant_access_token`
- 获取 `chat_id`
- 应用机器人加入群聊的操作说明
- 权限与可用范围配置说明
- 如果后续要支持卡片，还需要卡片 JSON / 模板文档

因此，后续正式开发飞书渠道前，建议再补一份：

- `docs/feishu-auth-reference.md`

## 12. 推荐的实现顺序

1. 先做 `feishu` account / `feishu_chat` target 模型支持
2. 只支持 `text` 消息直发
3. 使用环境变量提供 `tenant_access_token`
4. 跑通群聊发送、历史记录、去重
5. 再考虑 `post`
6. 最后再考虑 card / 素材上传 / token 自动刷新
