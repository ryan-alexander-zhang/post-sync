# Obsidian QuickAdd 集成

本目录提供一套最小集成方案，用于在 Obsidian 中通过 QuickAdd 触发脚本，将当前打开的 Markdown 上传到 `post-sync`，并根据 frontmatter 自动发布到指定 target。

目录说明：

- `publish-current-note.js`
  - QuickAdd User Script
- `target-aliases.json`
  - QuickAdd 默认读取的 alias 映射文件，需由你填写真实 target ID
- `target-aliases.example.json`
  - alias 映射样例
- `frontmatter-example.md`
  - 推荐的笔记 frontmatter 示例

## 1. Frontmatter 规范

推荐使用顶层 `post_` 字段，便于直接在 Obsidian Properties 中维护。

```yaml
---
post_title: Weekly Update
post_publish: true
post_targets:
  - telegram.release-notes
  - feishu.team-news
post_template: default
---
```

字段说明：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `post_title` | string | 否 | 发布标题。脚本会把它映射为上传内容中的 `title` |
| `post_publish` | boolean | 是 | 必须为 `true` 才允许发布 |
| `post_targets` | string[] | 是 | target alias 列表 |
| `post_template` | string | 否 | 发布模板名，默认 `default` |

说明：

- 脚本会在上传前移除 `post_publish`、`post_targets`、`post_template`
- 脚本会保留 `post_title`，并额外写入 `title`，保证后端当前的标题提取逻辑可直接工作

## 2. Alias 映射

由于当前后端 target 只有 `id`，没有稳定的 alias 字段，推荐在 Obsidian vault 中维护一份本地 JSON 映射表。

默认文件：

```text
obsidian-quick-add/target-aliases.json
```

格式：

```json
{
  "telegram.release-notes": ["target_id_1"],
  "feishu.team-news": ["target_id_2"],
  "all.announce": ["target_id_1", "target_id_2"]
}
```

规则：

- key 是写入笔记 frontmatter 的 alias
- value 是后端 `channel-target` 的 ID 数组
- 一个 alias 可映射多个 target ID

## 3. QuickAdd 设置

本脚本基于 QuickAdd 官方 User Script 机制，脚本文件必须放在 vault 内且不能放到 `.obsidian/` 或隐藏目录下。

参考 QuickAdd 官方文档：

- User Scripts: https://quickadd.obsidian.guide/docs/next/UserScripts
- Macros: https://quickadd.obsidian.guide/docs/Choices/MacroChoice/

### 3.1 放置文件

把整个 `obsidian-quick-add/` 目录放进你的 Obsidian vault 根目录，确保脚本路径类似：

```text
<your-vault>/obsidian-quick-add/publish-current-note.js
```

### 3.2 创建 QuickAdd Macro

1. 打开 Obsidian
2. 打开 `Settings -> QuickAdd`
3. 点击 `Add Choice`
4. 选择 `Macro`
5. 命名为 `Publish Current Note`
6. 点击该 Choice 右侧的配置按钮
7. 在 Macro Builder 里添加一个 `User Script`
8. 选择 `obsidian-quick-add/publish-current-note.js`

### 3.3 配置脚本参数

为该脚本设置以下参数：

| 参数 | 推荐值 | 说明 |
|---|---|---|
| `API Base URL` | `http://localhost:8080/api/v1` | `post-sync` 后端 API 地址 |
| `Alias Mapping Path` | `obsidian-quick-add/target-aliases.json` | alias 映射文件路径 |
| `Default Template` | `default` | 默认模板 |
| `Show Notice` | `true` | 是否显示执行结果通知 |

### 3.4 绑定命令

可选做法：

- 在 QuickAdd 里给该 Choice 设置 hotkey
- 或在 Obsidian `Hotkeys` 中搜索 `QuickAdd: Choice: Publish Current Note`

## 4. 脚本执行流程

脚本执行逻辑：

1. 获取当前活动 Markdown 文件
2. 从 Obsidian metadata cache 读取 frontmatter
3. 校验：
   - `post_publish === true`
   - `post_targets` 是非空数组
4. 读取本地 alias 映射 JSON
5. 将 `post_targets` 解析为后端 `targetIds`
6. 生成用于上传的 Markdown：
   - 去掉 frontmatter 中的 `post_publish`
   - 去掉 frontmatter 中的 `post_targets`
   - 去掉 frontmatter 中的 `post_template`
   - 若存在 `post_title`，则写入 `title`
7. 调用 `POST /contents/upload`
8. 调用 `POST /publish-jobs`
9. 在 Obsidian 中显示 `contentId` 和 `jobId`

## 5. 常见失败

### 5.1 `post_publish must be true`

说明当前笔记没有显式允许发布。

修复：

```yaml
post_publish: true
```

### 5.2 `Missing target alias mapping`

说明 `post_targets` 中至少有一个 alias 没有在本地 JSON 映射文件中定义。

修复：

- 检查 `obsidian-quick-add/target-aliases.json`
- 检查 alias 是否拼写一致

### 5.3 `duplicate content upload`

说明后端检测到正文重复，当前项目会拒绝重复上传。

这不是 QuickAdd 脚本错误，而是后端当前内容去重策略生效。

### 5.4 `RESOURCE_NOT_FOUND`

说明 alias 虽然映射成功，但对应的 target ID 已经在后端不存在。

修复：

- 重新从 `post-sync` 后台确认 target ID
- 更新 `target-aliases.json`

## 6. 建议工作流

1. 在 `post-sync` 后台先创建 channel account 和 channel target
2. 把 target ID 填入 `target-aliases.json`
3. 在笔记 frontmatter 中填写：
   - `post_title`
   - `post_publish`
   - `post_targets`
   - `post_template`
4. 用 QuickAdd 触发 `Publish Current Note`
