# AiP2P Theme Manifest 与当前实现

## 1. 文档目的

这份文档定义 `aip2p` theme manifest 的当前实现边界，以及仍可继续扩展的方向。

目标是让宿主在不解析 theme 内部实现细节的前提下，先知道：

- 这个 theme 是什么
- 它适合哪些页面模型
- 它支持哪些插件
- 它从哪里加载模板和静态资源
- 它是否兼容当前宿主和插件输出模型

## 2. 为什么 theme 也要 manifest

如果 theme 没有 manifest，宿主会遇到这些问题：

- 不知道 theme 入口
- 不知道 theme 支持哪些页面
- 不知道缺模板时如何报错
- 不能提前做兼容性检查
- AI agent 难以自动创建和派生 theme

因此 theme 也必须是 manifest-first。

## 3. 建议的文件名

当前实现统一使用：

- `aip2p.theme.json`

theme 目录至少包含：

- `aip2p.theme.json`
- `templates/`
- `static/`
- `README.md`

可选：

- `theme.config.json`
- `examples/`

## 4. 设计目标

theme manifest 应满足：

1. 清楚声明 theme 身份
2. 清楚声明模板入口
3. 清楚声明支持的页面模型
4. 清楚声明支持的插件
5. 清楚声明兼容版本
6. 清楚声明是否为某个插件的默认 theme

## 5. 建议的最小字段

- `id`
- `name`
- `version`
- `description`
- `theme_api_version`
- `templates`
- `static`
- `supports`
- `supported_plugins`
- `required_plugins`
- `default_for_plugins`

## 6. 字段解释

### 6.1 `id`

theme 唯一标识。

例如：

- `default-news`
- `default-blog`
- `magazine-news`

### 6.2 `name`

用户可读名字。

### 6.3 `version`

theme 版本。

### 6.4 `description`

一句话说明视觉方向和适用范围。

### 6.5 `theme_api_version`

声明 theme 依赖的宿主 theme API 版本。

### 6.6 `templates`

模板目录入口。

### 6.7 `static`

静态资源目录入口。

### 6.8 `supports`

声明 theme 支持哪些页面模型。

例如：

- `NewsFeedPageModel`
- `NewsPostPageModel`
- `ArchiveDayPageModel`
- `NodeStatusPageModel`

### 6.9 `supported_plugins`

声明这个 theme 可以用于哪些插件。

它的作用不是把 theme 变成功能插件，而是明确兼容边界。

例如：

- `news-content`
- `news-archive`
- `forum-content`
- `shop-content`

如果宿主后续支持更复杂的应用组合，也可以扩展为对一组插件组合做校验。

### 6.10 `default_for_plugins`

声明哪些插件或应用组合默认使用这个 theme。

例如：

- `news-content`
- `news-app`

### 6.11 `required_plugins`

声明这个 theme 运行时依赖哪些功能插件。

这和 `supported_plugins` 不一样：

- `supported_plugins` 表示兼容边界
- `required_plugins` 表示缺了就不应该加载

例如一个论坛主题可以声明：

- 支持 `forum-content`
- 支持 `forum-moderation`
- 但要求至少启用 `forum-content`

如果一个主题依赖多个插件共同提供页面或导航，也可以直接把这些依赖列出来。

## 7. 建议的扩展字段

后续仍可扩展这些字段：

- `author`
- `homepage`
- `license`
- `config_schema`
- `fallback_theme`
- `features`
- `preview`

## 8. `config_schema`

theme 也可以有配置 schema，但只应作用于显示层。

例如：

- 首页布局模式
- 主色调
- logo URL
- 列表卡片样式
- 是否显示某些 side panel

不应通过 theme config 控制：

- 同步规则
- writer policy
- 节点治理逻辑

## 9. `fallback_theme`

后续可以支持 theme fallback。

如果当前 theme 缺某个模板，可以选择：

- 使用 fallback theme 对应模板
- 明确报错

这样能降低 theme 开发初期的门槛。

## 10. 当前 manifest 示例

```json
{
  "id": "default-news",
  "name": "Default News",
  "version": "0.1.0",
  "description": "The default AiP2P News theme based on the current aip2p-news UI.",
  "theme_api_version": "0.1",
  "templates": "./templates",
  "static": "./static",
  "supported_plugins": ["news-content", "news-archive", "news-governance", "news-ops"],
  "required_plugins": ["news-content"],
  "supports": [
    "NewsFeedPageModel",
    "NewsPostPageModel",
    "NewsDirectoryPageModel",
    "ArchiveIndexPageModel",
    "ArchiveDayPageModel",
    "ArchiveMessagePageModel",
    "NodeStatusPageModel",
    "WriterPolicyViewModel"
  ],
  "default_for_plugins": ["news-content"],
  "config_schema": "./theme.config.schema.json"
}
```

## 11. 宿主如何使用 theme manifest

宿主读取 theme manifest 后，应按这个顺序处理：

1. 校验 manifest 是否存在
2. 校验 JSON 结构
3. 校验 `theme_api_version`
4. 校验模板目录和静态目录是否存在
5. 读取 `supported_plugins`
6. 读取 `required_plugins`
7. 读取 `supports`
8. 判断是否和当前插件输出模型兼容
9. 再将其注册为可用 theme

## 12. 对第三方和 AI agent 的价值

theme manifest 之所以重要，是因为它让下列事情更容易自动化：

- 自动创建 theme 模板
- 自动检查模板缺失
- 自动判断与插件兼容性
- 自动做 theme 派生
- 自动安装 theme

## 13. 当前阶段结论

当前阶段已经确认：

1. theme 必须有独立 manifest
2. theme manifest 必须声明支持的 page model
3. theme manifest 可以声明支持哪些插件，并可通过 `required_plugins` 声明运行依赖，但仍然和插件保持分离
4. theme config 只作用于显示层
5. 当前 `aip2p-news` UI 已经收敛为 `default-news` theme 的实现来源

## 14. 当前实现状态

当前主线已经支持：

- 内置 theme 注册
- 目录 theme 加载
- `required_plugins` 与 `supported_plugins` 校验
- `templates/` 与 `static/` 路径校验
- `themes inspect --dir`
- `themes install/link/list/remove`

默认 `default-news` theme 已作为正式内置 theme 工作，而不是临时迁移资源目录。
