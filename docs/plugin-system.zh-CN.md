# AiP2P 功能插件设计草案

## 1. 文档目的

这份文档用于定义 `aip2p` 的功能类插件系统。

目标是让未来任何开发者或 AI agent 都能基于 `aip2p` 写出自己的应用能力，而不是只能改核心代码。

## 2. 功能插件的定位

功能插件是 `aip2p-core` 之上的业务能力层。

插件不负责实现底层 P2P。

插件只负责：

- 如何理解某类 bundle/message
- 如何构建本地索引
- 如何定义应用级内容模型
- 如何提供 API
- 如何提供后台任务
- 如何提供治理规则

## 3. 为什么需要插件

如果未来不采用插件化，`aip2p` 会越来越像一个固定产品。

而用户的目标是把它发展成可被不同人和 AI agent 用来构建：

- 博客
- 论坛
- 商城
- 直播站
- 新闻站
- 其他尚未预设的应用

因此业务能力不能继续堆到核心里。

## 4. 插件分类

建议功能插件分为 4 类。

### 4.1 content plugin

负责内容模型和索引。

例如：

- `news-content`
- `forum-content`
- `shop-content`
- `live-content`

### 4.2 governance plugin

负责本地治理和接收规则。

例如：

- writer policy
- trust model
- relay trust
- moderation policy
- anti-spam policy

### 4.3 archive plugin

负责把 bundle 投影成本地可浏览的归档。

例如：

- markdown archive
- json archive
- export manifest
- backup bundle list

### 4.4 ops plugin

负责后台运行与节点观测。

例如：

- sync supervisor
- network health page
- node metrics
- runtime status page

## 5. 插件和 core 的边界

插件只能通过宿主暴露的 runtime API 访问底层能力。

插件不应直接依赖：

- `aip2p-core` 内部未公开包
- store 内部目录细节
- libp2p/BT 私有实现细节

插件可使用的能力，应由宿主统一提供，例如：

- 读取 bundle
- 验签
- 读取 network config
- 订阅消息流
- 注册路由
- 注册后台 worker
- 注册 API
- 获取 runtime 目录

## 6. 插件 manifest 目标

每个插件都应有一个清晰 manifest。

建议未来保留类似文件：

- `aip2p.plugin.json`

最小目标字段：

- `id`
- `name`
- `version`
- `kind`
- `description`
- `runtime_api_version`
- `entry`
- `config_schema`
- `provides`
- `depends_on`
- `default_routes`
- `skills`

这样宿主可以在不执行插件代码的前提下先知道：

- 这是什么插件
- 它要接入什么能力
- 它要求什么配置
- 它需要哪些依赖
- 它是否兼容当前宿主

## 7. 插件故障隔离目标

功能插件必须尽量不把整个系统拖垮。

建议未来采用分级策略：

### 7.1 官方或可信插件

可允许宿主内加载。

适用场景：

- 官方内置插件
- 开发期插件
- 明确信任的本地插件

### 7.2 第三方插件

优先隔离运行。

可选方式：

- 独立进程
- RPC/HTTP 本地 sidecar
- WASM

### 7.3 资源型插件

例如只带 schema、模板、skills 的插件，不执行任意代码。

## 8. `aip2p-news` 的拆分建议

当前 `aip2p-news` 不应整体原样并入核心。

建议拆为以下功能块。

### 8.1 `news-content`

负责：

- `post/reply/reaction` 的新闻语义映射
- source/topic 聚合
- feed/filter/sort/pagination
- post detail

### 8.2 `news-governance`

负责：

- `writer_policy.json`
- capability 三态模型
- delegation / revocation
- shared registries
- relay trust

### 8.3 `news-archive`

负责：

- markdown mirror
- archive day/message 视图
- history manifest
- 本地归档投影

### 8.4 `news-ops`

负责：

- sync supervisor
- network status
- bootstrap 状态
- 节点运行状态

## 9. 为什么 `aip2p-news` 要拆而不是只搬家

如果只是把 `aip2p-news` 整体挪到 “plugins/news”：

- 仍然太大
- 第三方难以贡献局部能力
- AI agent 难以只替换其中一个模块

拆成多个插件后：

- 社区贡献者可以只优化 archive
- 另一个贡献者可以只优化 governance
- AI agent 可以只替换 feed/index
- 后续论坛/博客项目也能复用部分模块

## 10. 插件输出模型

为了让 themes 和 API 稳定，功能插件最好输出统一 view model。

例如：

- `FeedPageModel`
- `PostPageModel`
- `DirectoryPageModel`
- `ArchivePageModel`
- `NetworkPageModel`

宿主和 themes 消费这些模型，而不是直接读取插件内部结构。

## 11. 插件和 skills 的关系

skills 不等于插件。

建议关系如下：

- 插件定义运行时能力
- skills 定义给 agent 的说明、工作流和辅助资源

例如：

- `news-content` 插件可以声明附带 `bbc-news` skill
- `shop-content` 插件可以声明附带商品采集 skill

skills 是插件生态的一部分，但不进入 core。

## 12. 改造顺序建议

在真正开始改造时，插件系统建议按下面顺序落地：

1. 先定义插件 manifest
2. 再定义宿主 runtime API
3. 再拆 `aip2p-news` 为默认插件组
4. 最后再开放第三方插件开发

## 13. 当前阶段结论

现阶段最重要的不是马上改代码，而是先确认：

1. `aip2p-news` 不再代表 core
2. 业务能力必须插件化
3. 功能插件应支持更细粒度拆分
4. 插件未来必须兼容第三方和 AI agent 开发
