# AiP2P 模块化改造路线图 v1

## 1. 文档目的

这份文档把前面的设计草案收敛成一个正式改造路线图。

目标是让后续代码改造有明确顺序、明确阶段、明确验收标准。

这份路线图默认建立在已经确认的前提上：

- `aip2p-core` 只做底层
- `aip2p-news` 变默认参考应用
- 功能插件和 theme 严格分开
- 第三方和 AI agent 友好性优先

## 2. 改造总目标

改造完成后，项目应达到下面状态：

1. `aip2p-core` 只保留 P2P 基础能力
2. `aip2p-host` 成为统一宿主
3. `aip2p-news` 不再是核心本体，而是默认应用组合
4. 当前 `aip2p-news` UI 成为默认 theme
5. 功能能力通过插件组合
6. 第三方和 AI agent 可以基于 manifest、theme、runtime API 开发扩展

## 3. 总体阶段

建议把改造拆成 6 个阶段。当前主线已经完成了前半段的大部分基础工作。

## 4. Phase 0：设计冻结

### 4.1 目标

冻结总原则和基础契约，避免边写代码边改规则。

### 4.2 输入

当前已经完成的文档：

- 模块化总架构
- `aip2p-news` 拆分设计
- 插件系统
- theme 系统
- 开发者平台
- 插件 manifest
- theme manifest
- runtime API

### 4.3 完成标准

- 总原则不再反复修改
- `aip2p-news` 拆分方向确认
- plugin manifest 方向确认
- theme manifest 方向确认
- runtime API 分组方向确认

### 4.4 当前状态

可视为基本完成。

## 5. Phase 1：建立宿主边界

### 5.1 目标

在代码层把 `aip2p-core` 和上层应用彻底分开。

### 5.2 主要工作

- 建立 `aip2p-host`
- 统一插件发现与注册入口
- 统一 theme 注册入口
- 统一 runtime API 入口
- 建立 page model 渲染通道

### 5.3 不做的事

- 不先重写所有业务
- 不先改视觉
- 不先做论坛/商城新应用

### 5.4 当前状态

- 宿主已经可以在不依赖 `aip2p-news` 单体入口的情况下启动
- 宿主已经可以识别内置插件、内置 theme、目录扩展和已安装扩展
- 宿主已经可以基于 manifest 做加载、校验和组合判断

## 6. Phase 2：先拆 theme

### 6.1 目标

先把当前 `aip2p-news` UI 从业务逻辑里分离出来。

### 6.2 主要工作

- 把 templates/static 独立成 `default-news`
- 让页面渲染通过 theme 层完成
- 让 theme 只接收 page model，不直接读业务内部结构

### 6.3 为什么先拆 theme

因为当前最大耦合点之一就是 UI 和业务代码混在一起。

theme 先拆出来后：

- 页面边界会清楚很多
- 后面拆业务模块的阻力会更小
- 也能更早给第三方和 AI agent 一个可复制 theme 模板

### 6.4 当前状态

- 当前 `aip2p-news` UI 已以 `default-news` theme 身份运行
- theme 不直接读 store 或 runtime 私有文件
- theme 已支持 manifest、required_plugins 和目录方式加载

## 7. Phase 3：拆 `news-content`

### 7.1 目标

把新闻内容索引和展示数据模型从单体中拆出。

### 7.2 主要工作

- 抽出 `post/reply/reaction` 到新闻模型的映射
- 抽出 feed/source/topic/post 的 page model
- 抽出过滤、排序、分页逻辑
- 让这些逻辑成为 `news-content` 插件

### 7.3 当前状态

- `news-content` 已独立提供内容页和内容 API
- `default-news` theme 只消费对应 view model
- 内容逻辑已不再以整包 `news` 插件形式对外暴露

## 8. Phase 4：拆 `news-governance`

### 8.1 目标

把治理逻辑从内容逻辑和页面逻辑中拆出来。

### 8.2 主要工作

- 抽出 `writer_policy.json` 处理
- 抽出 capability 逻辑
- 抽出 delegation / revocation 逻辑
- 抽出 shared registry / relay trust 逻辑
- 统一输出治理 view model

### 8.3 当前状态

- `news-governance` 已独立提供治理页面
- 治理逻辑不再作为默认暴露的大插件入口存在
- 当前剩余工作主要是继续压缩共享 runtime 层

## 9. Phase 5：拆 `news-archive` 与 `news-ops`

### 9.1 目标

把 archive 和 runtime/ops 从主内容逻辑里彻底拆出。

### 9.2 主要工作

- 抽出 markdown mirror
- 抽出 archive page model
- 抽出 history manifest
- 抽出 sync supervisor
- 抽出 network status 与 node health 逻辑

### 9.3 当前状态

- `news-archive` 已独立提供归档相关页面和 API
- `news-ops` 已独立提供 network/status 页面和 API
- 当前剩余工作主要是继续清理共享层中仍然保留的公共 helper

## 10. Phase 6：开放扩展生态

### 10.1 目标

在内置参考应用稳定后，再正式开放第三方开发体验。

### 10.2 主要工作

- 提供插件模板
- 提供 theme 模板
- 提供示例 skills
- 提供开发文档
- 提供 manifest/schema 校验工具
- 提供 `aip2p create plugin/theme/app` 脚手架

### 10.3 完成标准

- 第三方开发者可以独立创建插件
- AI agent 可以根据模板自动生成插件和 theme
- 错误信息与兼容性信息足够清楚

## 11. 当前建议的实施顺序

后续正式动代码时，建议严格按这个顺序：

1. host 边界
2. default theme
3. news-content
4. news-governance
5. news-archive
6. news-ops
7. 第三方脚手架与模板

不建议直接跳到：

- 先做论坛插件
- 先做商城插件
- 先做更多 UI

因为底座没稳定前，越往上堆越容易返工。

## 12. 每个阶段都要检查的风险

### 12.1 核心膨胀风险

每次拆分时都要检查：

- 这个能力是否真的必须留在 core

如果不是，就不应重新塞回 core。

### 12.2 插件与 theme 重新耦合风险

每次改动都要检查：

- theme 是否偷偷依赖了业务内部结构

### 12.3 单体回潮风险

每次改动都要检查：

- `aip2p-news` 是否又被做回一个大单体

### 12.4 AI agent 不友好风险

每次改动都要检查：

- 新增结构是否仍然可通过 manifest、schema、模板和文档理解

## 13. 当前阶段后的直接下一步

当前直接下一步已经变成：

- 继续收紧 `internal/plugins/news` 共享运行时层
- 继续清理过时文档表述
- 继续打磨第三方开发和安装链路

## 14. 结论

这份路线图的核心意思很简单：

- 先定边界
- 再定宿主
- 再拆 theme
- 再拆业务
- 最后开放生态

只要顺序不乱，`aip2p` 才有机会从一个 demo 演变成真正可扩展的 P2P 应用底座。
