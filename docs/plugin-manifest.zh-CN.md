# AiP2P 插件 Manifest 设计草案

## 1. 文档目的

这份文档定义未来 `aip2p` 功能插件的 manifest 方向。

目标是让宿主在不执行插件代码的前提下，先知道：

- 这个插件是什么
- 这个插件做什么
- 这个插件需要什么配置
- 这个插件依赖什么能力
- 这个插件是否兼容当前宿主

这也是第三方和 AI agent 友好的前提。

## 2. 为什么要先有 manifest

如果未来插件没有清晰 manifest，就会出现这些问题：

- 宿主无法先做兼容性判断
- 无法在加载前给出清晰错误
- AI agent 无法知道入口在哪里
- 插件安装和发现只能依赖隐式目录约定
- theme 和插件之间的关系难以声明

因此，manifest 必须成为插件系统第一层契约。

## 3. 建议的文件名

建议未来统一使用：

- `aip2p.plugin.json`

每个插件目录至少包含：

- `aip2p.plugin.json`
- `README.md`
- `config.schema.json`
- `src/` 或 `handlers/`
- 可选 `skills/`

## 4. manifest 的设计目标

manifest 必须满足 6 个目标：

1. 宿主可读
2. 人类开发者可读
3. AI agent 可读
4. 不依赖执行插件代码
5. 支持版本兼容检查
6. 支持插件分类与依赖声明

## 5. 建议的最小字段

建议最小字段如下：

- `id`
- `name`
- `version`
- `description`
- `plugin_kind`
- `entry`
- `runtime_api_version`
- `config_schema`
- `provides`
- `depends_on`
- `default_routes`

## 6. 字段解释

### 6.1 `id`

插件唯一标识。

例如：

- `news-content`
- `news-governance`
- `forum-content`

### 6.2 `name`

面向用户展示的名字。

### 6.3 `version`

插件版本。

用于：

- 插件升级
- 问题定位
- 兼容性报告

### 6.4 `description`

一句话说明插件用途。

AI agent 和第三方开发者都需要这条信息快速判断能否复用。

### 6.5 `plugin_kind`

建议未来明确插件类型。

可选值可以先从这几类开始：

- `content`
- `governance`
- `archive`
- `ops`
- `search`
- `commerce`
- `stream`

### 6.6 `entry`

插件入口文件。

宿主用这个字段去加载插件实际实现。

### 6.7 `runtime_api_version`

声明插件依赖的宿主 API 版本。

宿主应在加载前校验：

- 当前宿主是否兼容
- 是否需要提示升级

### 6.8 `config_schema`

指向插件自己的配置 schema 文件。

这样宿主可以在加载前先校验配置，而不是等运行时 panic。

### 6.9 `provides`

声明这个插件提供哪些能力。

例如：

- `page_models`
- `api_endpoints`
- `background_workers`
- `governance_rules`
- `archive_projections`

### 6.10 `depends_on`

声明插件依赖的其他插件或宿主能力。

例如：

- 依赖 `news-content`
- 依赖 `identity-signature`
- 依赖 `archive-service`

### 6.11 `default_routes`

声明插件建议注册的默认路由。

例如：

- `/`
- `/posts/:id`
- `/archive`
- `/network`

宿主可以据此做路由检查和冲突提示。

## 7. 建议的扩展字段

除了最小字段，建议未来还支持这些扩展字段：

- `author`
- `homepage`
- `license`
- `skills`
- `default_theme`
- `permissions`
- `runtime_mode`
- `min_host_version`
- `max_tested_host_version`

## 8. `runtime_mode`

建议未来让插件显式声明运行模式。

例如：

- `in_process`
- `isolated_process`
- `wasm`
- `resource_only`

这样宿主可以根据插件信任等级决定是否允许加载。

## 9. `permissions`

建议未来让插件声明它需要的权限。

例如：

- `read_store`
- `write_runtime`
- `register_http_routes`
- `spawn_background_worker`
- `read_network_config`

这对第三方插件和 AI agent 友好性非常重要。

原因是：

- 宿主可以先判断是否放行
- 用户可以先理解风险
- AI agent 可以先知道它能调用哪些能力

## 10. `skills`

建议未来插件可以在 manifest 中声明附带的 skills 目录。

例如：

- `skills/news-ingestion`
- `skills/news-publishing`

这样 skills 就成为插件生态的一部分，而不是散落在系统各处。

## 11. 一个草案示例

```json
{
  "id": "news-content",
  "name": "AiP2P News Content",
  "version": "0.1.0",
  "description": "Builds news feed, source, topic, and post page models from AiP2P bundles.",
  "plugin_kind": "content",
  "entry": "./src/plugin.js",
  "runtime_api_version": "0.1",
  "config_schema": "./config.schema.json",
  "provides": {
    "page_models": ["NewsFeedPageModel", "NewsPostPageModel", "NewsDirectoryPageModel"],
    "routes": ["/", "/posts/:infohash", "/sources", "/topics"]
  },
  "depends_on": [],
  "default_routes": ["/", "/posts/:infohash", "/sources", "/topics"],
  "skills": ["./skills/news-ingestion", "./skills/news-publishing"],
  "runtime_mode": "in_process",
  "permissions": ["read_store", "register_http_routes", "read_network_config"]
}
```

## 12. 宿主如何使用 manifest

宿主读取 manifest 后，应按这个顺序处理：

1. 判断文件是否存在
2. 校验 JSON 结构
3. 校验必填字段
4. 校验 `runtime_api_version`
5. 读取并校验 `config_schema`
6. 检查依赖是否满足
7. 检查路由是否冲突
8. 再决定是否真正加载插件代码

## 13. 对第三方和 AI agent 的价值

如果这个 manifest 契约成立，第三方和 AI agent 将更容易做到：

- 自动生成新插件
- 自动检查兼容性
- 自动填充模板
- 自动诊断缺失字段
- 自动组装应用

## 14. 当前阶段结论

进入代码改造前，建议先确认：

1. 插件必须有独立 manifest
2. manifest 必须先于代码执行被读取
3. manifest 必须支持类型、依赖、权限、版本声明
4. plugins 附带 skills 的能力应通过 manifest 显式声明
