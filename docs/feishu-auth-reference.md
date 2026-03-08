# Feishu Auth Reference

该文件基于 `docs/feishu-token.md` 提取，供后续实现飞书渠道鉴权时直接使用。

## 1. 用途

自建应用通过飞书开放平台接口获取 `tenant_access_token`，后续调用发消息接口时使用：

- Header: `Authorization: Bearer <tenant_access_token>`

## 2. 获取接口

- Method: `POST`
- URL: `https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal`

请求头：

- `Content-Type: application/json; charset=utf-8`

请求体：

```json
{
  "app_id": "cli_xxx",
  "app_secret": "xxx"
}
```

字段说明：

- `app_id`
  - 飞书应用唯一标识
- `app_secret`
  - 飞书应用秘钥

## 3. 响应结构

成功时返回：

```json
{
  "code": 0,
  "msg": "ok",
  "tenant_access_token": "t-xxx",
  "expire": 7200
}
```

关键字段：

- `code`
  - `0` 表示成功
- `msg`
  - 错误或状态描述
- `tenant_access_token`
  - 后续调用飞书开放平台接口使用的 bearer token
- `expire`
  - token 过期时间，单位秒

## 4. 生命周期规则

原始文档给出的关键规则：

- `tenant_access_token` 最大有效期是 `2` 小时
- 剩余有效期 `< 30` 分钟时，再调用获取接口会返回一个新的 token
- 剩余有效期 `>= 30` 分钟时，再调用获取接口会返回原有 token

这意味着后续实现时可以采用保守缓存策略：

- 内存缓存 token
- 记录实际过期时间
- 在到期前 `30` 分钟主动刷新

## 5. 对当前项目的配置映射建议

飞书渠道建议新增这些环境变量：

- `FEISHU_APP_ID`
- `FEISHU_APP_SECRET`

如果要兼容本地临时调试，也可以允许：

- `FEISHU_TENANT_ACCESS_TOKEN`

推荐优先级：

1. 若显式提供 `FEISHU_TENANT_ACCESS_TOKEN`，则直接使用
2. 否则使用 `FEISHU_APP_ID + FEISHU_APP_SECRET` 获取并缓存 token

对 `ChannelAccount` 的建议：

- `channelType = "feishu"`
- `secretRef = FEISHU_APP_SECRET`
- `configJson` 保存：
  - `appIdEnv: "FEISHU_APP_ID"`
  - 可选 `tokenEnv: "FEISHU_TENANT_ACCESS_TOKEN"`
  - 可选 `baseUrl`

说明：

- 当前项目已有 `secretRef -> os.Getenv(secretRef)` 读取机制
- 但飞书需要同时读取 `app_id` 和 `app_secret`
- 因此飞书渠道不应完全复用 Telegram 的“只有一个 secretRef”的假设
- 更合理的做法是为渠道驱动引入独立的 credential resolver / token provider

## 6. 推荐的代码设计

为了尽量降低未来新增渠道的改动负担，飞书鉴权不要直接散落在业务层，建议引入以下抽象：

- `CredentialProvider`
  - 从环境变量或账号配置解析原始凭证
- `TokenProvider`
  - 按渠道获取可直接调用 API 的 access token
- `ChannelDriver`
  - 只关心 render / send，不直接拼业务层配置

对飞书而言：

- `FeishuCredentialProvider`
  - 负责读取 `app_id`、`app_secret`、可选静态 token
- `FeishuTokenProvider`
  - 负责获取和缓存 `tenant_access_token`
- `FeishuDriver`
  - 负责把渲染结果发送到飞书接口

## 7. 错误处理建议

获取 token 时至少记录：

- HTTP status
- 响应 `code`
- 响应 `msg`

建议错误分类：

- 配置缺失
  - `CONFIG_ERROR`
- token 获取失败
  - `CHANNEL_AUTH_FAILED`
- 飞书消息发送失败
  - `CHANNEL_SEND_FAILED`

## 8. 当前实现建议

新增飞书渠道时，推荐按这个顺序落地：

1. 先新增 `FeishuTokenProvider`
2. 再让 `FeishuDriver` 依赖 token provider
3. 再把业务层中的“直接读取 secretRef 环境变量”收敛成可扩展接口

这样后续再新增：

- 飞书卡片
- 企业微信
- X
- 其它 OAuth / token 刷新型渠道

都不需要继续在 `PublishService` 里增加渠道分支。
