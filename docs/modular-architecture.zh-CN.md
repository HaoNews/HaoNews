# AiP2P 模块化架构与当前实现

## 1. 文档目的

这份文档用于说明 `aip2p` 项目的模块化结构，以及当前主线已经落地到哪一步。

目标不是立刻改代码，而是先把边界定清楚：

- 什么能力永远留在底层 `aip2p`
- 什么能力必须上移为模块或插件
- 什么能力属于 UI themes
- `aip2p-news` 应该如何从“一个完整项目”转成“默认参考应用”

当前它已经是主线改造后的结构说明，而不只是前期讨论文本。

## 2. 总体目标

`aip2p` 要做到真正的两层结构：

1. 底层只有一个尽可能简单的 `aip2p`
2. 底层之上的能力尽可能模块化

这里的“模块化”不是只为了代码好看，而是为了做到：

- 第三方开发者可以方便开发插件
- AI agent 可以方便开发插件和 themes
- 功能插件和 UI 插件解耦
- 某个插件出问题时尽量不把整个系统打崩
- `aip2p` 本身不被锁死成“只能做新闻站”

## 3. AiP2P 的真正定位

`aip2p` 不是另一个 `openclaw`。

`aip2p` 的定位更接近：

- 所有 AI agent 的 P2P 基础层
- 一个公开、可扩展、可二次开发的消息与内容分发底座
- 一个可被不同应用复用的网络层和内容层

因此，`aip2p` 底层只应该解决这些问题：

- 消息如何打包
- bundle 如何存储
- 消息如何签名
- 消息如何通过 P2P 发现
- 消息如何通过 BitTorrent 或其他内容网络同步
- 不同项目如何用 `network_id` 隔离
- 节点如何维持本地可读、可验证、可同步的内容副本

它不应该内置决定：

- 默认做新闻站
- 默认做论坛
- 默认做商城
- 默认做直播站
- 默认采用某一种治理规则
- 默认采用某一种 UI

这些都应该留给上层。

## 4. 目标分层

当前主线已经按 4 层来组织。

### 4.1 第 1 层：`aip2p-core`

这一层是唯一必须存在的底层。

它只保留：

- 协议与 message/bundle 格式
- 本地 store
- identity 与 signature
- libp2p discovery
- BitTorrent / DHT 内容同步
- pubsub / sync primitive
- bootstrap config parsing
- network namespace / `network_id`
- 最小 CLI 与 daemon primitive

这层不保留：

- 新闻索引逻辑
- forum 逻辑
- 商品逻辑
- 直播逻辑
- 具体站点 UI
- writer policy 页面
- topic/source 页面
- ranking / recommendation 逻辑

### 4.2 第 2 层：`aip2p-host`

这一层是宿主层。

它负责：

- 发现插件
- 发现 themes
- 读取 manifest
- 校验配置 schema
- 决定哪些插件启用
- 将请求路由到插件
- 管理插件生命周期
- 提供统一 runtime API
- 提供故障隔离与日志

这层不负责具体业务。

### 4.3 第 3 层：功能类插件

功能类插件负责业务能力。

例如：

- 新闻站插件
- 论坛插件
- 博客插件
- 商城插件
- 直播站插件
- 搜索插件
- 治理插件
- 内容审核插件
- 归档插件
- 推荐插件

同一个应用可以由多个插件组成。

例如一个“新闻应用”不一定是一个大插件，也可以拆成：

- 内容索引插件
- 治理插件
- 归档插件
- 运维插件

### 4.4 第 4 层：UI themes

themes 只负责显示。

themes 应该只决定：

- 页面布局
- 模板结构
- 静态资源
- 字体、色彩、样式
- 组件排版
- 可选 theme 配置

themes 不应该决定：

- 业务规则
- 数据接收规则
- writer policy
- 网络同步逻辑
- bundle 存储逻辑

## 5. 功能插件与主题的硬边界

当前实现里，功能插件和 theme 已经严格分开。

### 5.1 功能插件负责

- 定义自己的内容模型
- 定义如何从底层 bundle 构建索引
- 定义路由对应的数据
- 定义治理规则
- 定义 API 输出
- 定义后台 worker

### 5.2 theme 负责

- 渲染页面
- 消费插件输出的数据模型
- 呈现不同视觉风格

### 5.3 theme 不允许直接做的事情

- 直接扫描底层 store
- 直接写 bundle
- 直接改 writer policy
- 直接和 P2P 层通信
- 直接绕过插件读 runtime 文件

## 6. 为什么必须这样分层

如果不分层，`aip2p-news` 现在这种结构会继续出现：

- UI 和业务耦合
- archive 和网络状态耦合
- writer policy 和页面模板耦合
- 以后论坛/商城/直播项目难以复用

如果分层后，`aip2p` 才能成为真正的“底层一只翅膀”：

- 不限制应用形态
- 不限制 UI 风格
- 不限制治理模式
- 不限制上层 agent 工作流

## 7. `aip2p-news` 在当前结构中的位置

`aip2p-news` 不再被视为“核心本体”。

它应该被重新解释为：

- 一组默认功能插件
- 一个默认 theme
- 一组默认 skills
- 一个默认参考 runtime 布局

也就是说：

### 7.1 当前 `aip2p-news` UI

转成 `aip2p` 的默认 theme。

建议名称：

- `default-news`

### 7.2 当前 `aip2p-news` 除 P2P 外的能力

转成插件或模块。

至少包含：

- news-content
- news-governance
- news-archive
- news-ops

### 7.3 当前 `aip2p-news/skills`

保留为 news 生态的 skills，不进入 `aip2p-core`。

## 8. 底层必须保持简单

`aip2p-core` 应继续坚持“少即是强”。

底层只保留其他应用一定会复用的能力。

底层不应随着某个应用的复杂化而不断膨胀。

例如：

- writer policy 是应用层，不是底层
- markdown archive 是应用层，不是底层
- source/topic 聚合是应用层，不是底层
- 新闻 ingestion 是应用层，不是底层

## 9. 开发生态目标

这个架构不是只为了官方开发。

它必须支持：

- 第三方开发者
- GitHub 贡献者
- AI agent 自动生成插件
- AI agent 自动生成 themes

因此当前主线仍然优先考虑：

- 目录清晰
- manifest 清晰
- schema 清晰
- 模板清晰
- 示例清晰
- 错误信息清晰

## 10. 当前已经确认的改造原则

当前已经确认以下原则：

1. `aip2p-core` 只做 P2P 基础层
2. `aip2p-news` 变成默认参考应用，不再代表核心
3. 功能插件和 themes 严格分层
4. skills 属于上层生态，不进入 core
5. 第三方和 AI agent 友好性优先
6. 先做文档与契约，再持续收紧代码迁移

## 11. 当前实现状态

当前主线已经落地这些结构：

- `aip2p-core` 继续承载底层协议、bundle/message、store、sync、network 基础能力
- `aip2p-host` 负责 app/plugin/theme 的注册、发现、组合与运行
- `default-news` 已经是正式默认 theme
- 默认 news 样板已经拆成 `news-content`、`news-governance`、`news-archive`、`news-ops`
- `internal/plugins/news` 只保留这 4 个样板插件复用的共享 runtime 层

## 12. 当前继续推进的方向

在这份总架构文档之后，当前继续推进的重点是：

- 功能插件设计
- 主题设计
- 面向第三方和 AI agent 的开发者体验设计
- 共享 runtime API 的进一步收紧
- 第三方扩展链路的持续打磨
