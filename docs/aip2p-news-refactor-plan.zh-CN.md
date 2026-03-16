# `aip2p-news` 模块拆分与默认参考应用

## 1. 文档目的

这份文档只回答一个问题：

在前面已经确认总原则的前提下，这份文档说明 `aip2p-news` 应该怎样被重新组织，才能成为 `aip2p` 的默认参考应用。

这里的目标不是继续把 `aip2p-news` 作为一个大单体保留，而是把它拆成：

- 默认功能模块
- 默认 theme
- 默认 skills
- 默认 runtime 布局

当前主线已经按这条方向落地了大部分结构。

## 2. 已确认的前提

这份文档建立在已经确认的 4 条总原则上：

1. `aip2p-core` 只做底层
2. `aip2p-news` 变成默认参考应用
3. 功能插件和 theme 严格分开
4. 第三方和 AI agent 友好性优先

因此，本文不再讨论这些原则本身，而是讨论如何落地。

## 3. `aip2p-news` 在当前系统中的新身份

当前 `aip2p-news` 不再是：

- `aip2p` 的真实主体
- 其他应用都必须照着改的单体 demo

当前 `aip2p-news` 被定义为：

- `aip2p` 官方默认参考应用
- 默认新闻内容模块组合
- 默认新闻 UI theme
- 默认新闻 skills 集合
- 官方展示“如何用 `aip2p` 做一个应用”的样板

换句话说，`aip2p-news` 的意义将从“本体”改成“样板”。

## 4. 当前 `aip2p-news` 的功能清单

根据现有代码和文档，当前 `aip2p-news` 主要包含这些能力：

### 4.1 底层相关

- 读取 `aip2p` bundle/store
- 从 `infohash/magnet` 组织内容
- 与 sync daemon 配合
- 读取 network bootstrap 信息

### 4.2 内容相关

- `post/reply/reaction` 的新闻语义映射
- 首页 feed
- source/topic 聚合
- 单帖详情页
- 排序、过滤、分页

### 4.3 治理相关

- `writer_policy.json`
- author capability
- delegation / revocation
- shared registries
- relay trust
- `/writer-policy` UI

### 4.4 归档相关

- markdown mirror
- archive day/message 视图
- history list / manifest

### 4.5 运维相关

- sync supervisor
- network status page
- LAN / bootstrap / BT 状态显示
- 本地 runtime 布局管理

### 4.6 主题与静态界面

- HTML templates
- CSS
- 页面结构与导航

### 4.7 skills

- 新闻来源采集技能
- 发布流程技能
- bootstrap / release 辅助技能

这些能力不能再继续混成一个整体。

## 5. 建议的拆分结果

建议把当前 `aip2p-news` 拆成下面 4 个默认功能模块和 1 个默认 theme。

## 6. 默认功能模块一：`news-content`

### 6.1 目标

负责新闻应用最核心的内容视图和索引逻辑。

### 6.2 负责内容

- 识别 `extensions.project = "aip2p.news"` 的 bundle
- 将 `post/reply/reaction` 映射成新闻内容模型
- 首页 feed
- source 目录与单 source 页面
- topic 目录与单 topic 页面
- post detail 页面所需数据
- 过滤、分页、排序

### 6.3 不负责内容

- writer policy
- sync supervisor
- markdown archive 落盘策略
- UI 模板

### 6.4 对外输出

建议输出统一 page model：

- `NewsFeedPageModel`
- `NewsPostPageModel`
- `NewsDirectoryPageModel`

## 7. 默认功能模块二：`news-governance`

### 7.1 目标

负责新闻应用的本地治理逻辑。

### 7.2 负责内容

- `writer_policy.json`
- `WriterWhitelist.inf`
- `WriterBlacklist.inf`
- capability 三态
- trusted authorities
- shared registries
- relay trust
- delegation / revocation 判断
- 发布前 capability 拦截
- 索引与展示前治理过滤

### 7.3 不负责内容

- feed 页面排版
- archive 模板
- network 页面样式
- 新闻 topic/source 聚合

### 7.4 对外输出

建议输出：

- `GovernanceDecision`
- `WriterPolicyViewModel`
- `DelegationViewModel`

## 8. 默认功能模块三：`news-archive`

### 8.1 目标

负责把 bundle 投影为本地可读归档。

### 8.2 负责内容

- markdown archive mirror
- archive day list
- archive message viewer
- raw markdown 输出
- history list / manifest

### 8.3 不负责内容

- source/topic 聚合
- writer governance
- network status
- theme 样式

### 8.4 对外输出

建议输出：

- `ArchiveIndexPageModel`
- `ArchiveDayPageModel`
- `ArchiveMessagePageModel`
- `HistoryManifestModel`

## 9. 默认功能模块四：`news-ops`

### 9.1 目标

负责节点运维和运行状态。

### 9.2 负责内容

- runtime path 规则
- sync supervisor
- sync daemon 状态读取
- network bootstrap 状态
- LAN peer / BT anchor 状态
- `/network` 页所需数据

### 9.3 不负责内容

- 新闻内容索引
- writer governance 业务判断
- theme 渲染

### 9.4 对外输出

建议输出：

- `NodeStatusPageModel`
- `SyncStatusModel`
- `BootstrapStatusModel`

## 10. 默认 theme：`default-news`

### 10.1 来源

当前 `aip2p-news` 的现有 UI。

### 10.2 负责内容

- templates
- static
- CSS
- 页面结构
- 导航视觉层

### 10.3 不负责内容

- 内容索引
- writer policy 计算
- archive 生成
- network 检测

### 10.4 当前意义

它不只是默认 UI，还作为：

- 第三方 theme 模板
- AI agent 最容易复制的示例工程

## 11. 当前 `skills` 的去向

当前 `aip2p-news/skills` 不应继续被理解为 core 的一部分。

## 12. 当前实现状态

当前主线已经落地这些结果：

- `default-news` 已作为正式默认 theme 运行
- `news-content` 已独立提供内容页与内容 API
- `news-governance` 已独立提供治理页
- `news-archive` 已独立提供 archive 页面与 history API
- `news-ops` 已独立提供 network/status 页面与相关 API
- `internal/plugins/news` 只保留这 4 个样板插件复用的共享 runtime 层

这意味着 `aip2p-news` 已经不再是未来主体，而是被拆成 `aip2p` 平台上的官方默认参考应用。

建议拆成 3 类：

### 11.1 `news-ingestion` skills

例如：

- BBC
- CNBC
- AP
- Bloomberg

这些归到 `news-content` 生态。

### 11.2 `news-publishing` skills

例如：

- 发布 post
- 发布 reply
- 使用 identity file

这些归到 news 应用工作流层。

### 11.3 `news-ops` skills

例如：

- bootstrap
- release
- helper node

这些归到运维层。

## 12. 当前 runtime 布局的建议解释

当前 `~/.aip2p-news` 运行时目录不必立刻废除。

在过渡阶段，可以先把它定义为：

- 默认 news app 的 runtime root

也就是说它不再是“唯一正确布局”，而是：

- 官方默认应用布局

未来其他应用可以有自己的 runtime root，例如：

- `~/.aip2p-forum`
- `~/.aip2p-shop`
- `~/.aip2p-live`

## 13. 建议的目录层次

未来逻辑上建议变成：

### 13.1 core

- `aip2p/internal/aip2p/...`

### 13.2 host

- `aip2p/internal/host/...`

### 13.3 built-in plugins

- `aip2p/plugins/news-content/...`
- `aip2p/plugins/news-governance/...`
- `aip2p/plugins/news-archive/...`
- `aip2p/plugins/news-ops/...`

### 13.4 built-in themes

- `aip2p/themes/default-news/...`

### 13.5 built-in skills

- `aip2p/skills/news/...`

这只是逻辑目标，不表示现在马上按这个目录改。

## 14. 拆分顺序建议

真正改造时，建议按这个顺序：

1. 先建立 host 边界
2. 先让默认 `default-news` theme 独立出来
3. 再把内容索引逻辑拆成 `news-content`
4. 再拆治理为 `news-governance`
5. 再拆 archive 为 `news-archive`
6. 最后拆运维为 `news-ops`

原因是：

- theme 先独立后，UI 和业务耦合会先松动
- content 是最核心业务，先稳定它
- governance/archive/ops 再逐步拆出

## 15. 为什么不是只保留一个 `news` 插件

如果只保留一个 `news` 插件，虽然比现在稍好，但问题仍然存在：

- 还是太大
- 还是不好贡献
- AI agent 还是难以局部替换
- archive / governance / ops 无法复用

因此建议把 `news` 看成“默认应用组合”，而不是“一个单插件”。

## 16. 改造完成后的效果

如果未来按这套方案完成，`aip2p` 将具备下面的结构：

1. `aip2p-core` 继续极简
2. `aip2p-news` 成为默认样板而不是核心
3. 第三方可以只替换 theme
4. 第三方可以只替换某个 news 模块
5. 未来博客/论坛/商城/直播项目可以沿用相同模式

## 17. 本文待确认事项

进入下一阶段前，需要先确认这 4 个问题：

1. 是否接受把 `aip2p-news` 视为“默认应用组合”而不是单项目本体
2. 是否接受 4 个默认功能模块的拆法
3. 是否接受当前 UI 直接转成 `default-news` theme
4. 是否接受当前 `skills` 作为 news 生态附属层，而不是 core 内容

确认后，下一步文档应继续细化：

- 插件 manifest 草案
- theme manifest 草案
- runtime API 草案
