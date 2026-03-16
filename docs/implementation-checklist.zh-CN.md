# AiP2P 实施清单与当前状态

## 1. 文档目的

这份文档把前面的路线图进一步落实成执行清单，并标记当前主线已经完成到哪里。

目标是明确：

- 当前主线已经落地了哪些模块
- 共享层还剩哪些收尾项
- 每次继续收紧后应该怎么验证

## 2. 已完成的基础目标

当前主线已经完成这些基础目标：

- 建立 `aip2p-host`
- 建立 plugin/theme/app 注册与发现入口
- 建立内置样板插件与默认 theme
- 建立目录扩展、安装扩展、工作区扩展链路

这些基础能力已经不是规划状态，而是当前实现。

## 3. 当前已经落地的范围

- `aip2p` 仓库内新增 host 层目录
- `aip2p` 仓库内新增 plugin/theme 注册入口
- `aip2p` 仓库内新增 manifest 和 runtime API 的初始骨架
- `aip2p-news` 现有 UI 资源迁移为 `default-news` theme 来源

### 3.1 当前仍不作为 core 范围的内容

- `aip2p-core` 协议格式
- `aip2p` 现有 message/bundle/store/sync 主要逻辑
- `aip2p-news` 的治理细节
- `aip2p-news` 的 archive 细节
- 业务模型本身以外的新产品逻辑

## 4. 当前已经固定下来的模块边界

- `aip2p-core` 继续只承载底层协议、bundle/message、store、sync、network 基础能力
- `aip2p-host` 负责 app/plugin/theme 注册、发现、校验、组合与运行
- `default-news` 已经是默认主题，而不是业务插件
- 默认 news 样板能力已经拆成 `news-content`、`news-governance`、`news-archive`、`news-ops`
- `internal/plugins/news` 现在只保留这些 news 样板插件共享的 runtime、索引、治理/归档/运维辅助逻辑

## 5. Step 1：建立宿主目录和命名

### 5.1 目标

先把宿主层目录定下来，避免后续代码无处安放。

### 5.2 当前目录

- `aip2p/internal/host/`
- `aip2p/internal/plugins/`
- `aip2p/internal/themes/`

### 5.3 当前状态

- 宿主代码不再继续写进 `internal/aip2p`
- 插件和 theme 有明确归属位置

## 6. Step 2：插件注册入口

### 6.1 目标

让宿主具备“注册内置插件”的能力。

### 6.2 当前最小能力

- 注册插件 ID
- 注册插件 manifest
- 校验插件基本信息
- 根据配置启用或禁用插件

### 6.3 当前第一批内置插件

- `news-content`
- `news-governance`
- `news-archive`
- `news-ops`

当前主线已不再保留一个默认暴露的整包 `news` 插件。

## 7. Step 3：theme 注册入口

### 7.1 目标

让宿主具备“注册内置 theme”的能力。

### 7.2 当前最小能力

- 注册 theme ID
- 注册 theme manifest
- 判断 theme 是否支持某个 page model
- 为默认应用选择默认 theme

### 7.3 当前默认 theme

- `default-news`

## 8. Step 4：page model 通道

### 8.1 目标

先把业务输出和 theme 渲染之间的接口立住。

### 8.2 当前默认样板已使用的 page model

- `NewsFeedPageModel`
- `NewsPostPageModel`
- `NewsDirectoryPageModel`
- `ArchiveIndexPageModel`
- `NodeStatusPageModel`

当前默认样板已经通过这组 page model 对应的数据结构运行。

## 9. Step 5：默认 theme 资源

### 9.1 目标

把 templates/static 从现有单体逻辑中抽离成 theme 资源。

### 9.2 主要来源

当前主要来源：

- `aip2p-news/internal/latestapp/web/templates/`
- `aip2p-news/internal/latestapp/web/static/`

### 9.3 当前已经达到

- 资源位置独立
- theme 能被宿主识别
- 页面渲染有正式入口

当前不做：

- 让 theme 承载业务逻辑或直接操作底层 store

## 10. Step 6：runtime API 最小骨架

### 10.1 目标

先把宿主对插件的公开面固定下来。

### 10.2 当前已经存在的分组

- `bundle`
- `store`
- `network`
- `http`
- `theme`
- `runtime`

### 10.3 当前仍可继续收紧的分组

- `identity`
- `worker`

当前宿主已经可跑，后续主要是继续把共享 runtime API 再按领域收紧。

## 11. 当前默认应用入口

### 11.1 目标

`aip2p` 宿主已经以“默认应用组合”的方式启动，不再依赖老的单体入口。

### 11.2 目标效果

逻辑上类似：

- 宿主启动
- 宿主装载默认 news 插件组合
- 宿主装载 `default-news`
- 由宿主统一对外提供 HTTP 服务

## 12. 当前继续收紧的代码拆分

以下内容已经从主线大块拆开，但还可以继续往更细模块收紧：

- `writer_policy` 彻底拆出
- delegation / revocation 单独模块化
- archive mirror 单独模块化
- sync supervisor 单独模块化
- skills 装载体系正式化

当前边界已经立住，重点是继续压缩共享层和完善第三方开发体验。

## 13. 当前要重点看的代码位置

继续收尾时，重点看这几处：

### 13.1 `aip2p-core`

- `aip2p/internal/aip2p/`
- `aip2p/cmd/aip2p/main.go`

### 13.2 历史样板来源

- `aip2p-news/cmd/latest/main.go`
- `aip2p-news/internal/latestapp/server.go`

### 13.3 现有 UI 资源

- `aip2p-news/internal/latestapp/web/templates/`
- `aip2p-news/internal/latestapp/web/static/`

### 13.4 现有 runtime/path 和 network 逻辑

- `aip2p-news/internal/latestapp/runtimepaths.go`
- `aip2p-news/internal/latestapp/network.go`
- `aip2p-news/internal/latestapp/sync_supervisor.go`

## 14. 当前持续验证清单

开始改代码后，每完成一步至少验证这些项：

### 14.1 编译验证

- `aip2p` 仍可编译
- 现有 `aip2p` CLI 行为不被破坏

### 14.2 路由验证

- 默认首页仍可访问
- 默认详情页仍可访问
- theme 静态资源可正常加载

### 14.3 隔离验证

- theme 不直接读取 core 内部结构
- 插件入口不直接依赖模板文件的私有路径

### 14.4 回归验证

- 现有 news demo 的核心浏览体验不应立刻退化
- `aip2p` 核心 publish/show/verify/sync 不应因宿主改造受损

## 15. 当前继续推进的收尾条件

当前至少要持续满足下面条件：

1. 宿主边界已经存在
2. 默认 theme 已有正式身份
3. page model 通道已建立
4. 默认应用入口已从单体入口中松动出来

## 16. 当前继续推进的动作

1. 继续压缩 `internal/plugins/news` 共享层
2. 继续收紧 runtime API 的分域边界
3. 持续验证第三方开发链路 `create/inspect/validate/install/link/serve`
4. 继续清理文档中会误导第三方开发者和 AI agent 的阶段性表述
