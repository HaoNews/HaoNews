# AiP2P 第一阶段实施清单

## 1. 文档目的

这份文档把前面的路线图进一步落实成“真正开始改代码前的执行清单”。

目标是明确：

- 第一阶段具体改哪些地方
- 哪些地方暂时不动
- 每一步完成后怎么验证

## 2. 第一阶段目标

第一阶段只做一件事：

- 建立 `aip2p-host` 的基础边界，并让当前 `aip2p-news` UI 具备转成默认 theme 的入口条件

这一阶段不追求把所有业务都拆完。

## 3. 第一阶段范围

### 3.1 要动的范围

- `aip2p` 仓库内新增 host 层目录
- `aip2p` 仓库内新增 plugin/theme 注册入口
- `aip2p` 仓库内新增 manifest 和 runtime API 的初始骨架
- `aip2p-news` 现有 UI 资源的迁移准备

### 3.2 暂时不动的范围

- `aip2p-core` 协议格式
- `aip2p` 现有 message/bundle/store/sync 主要逻辑
- `aip2p-news` 的治理细节
- `aip2p-news` 的 archive 细节
- 第三方脚手架命令

## 4. 第一阶段建议改造顺序

## 5. Step 1：建立宿主目录和命名

### 5.1 目标

先把宿主层目录定下来，避免后续代码无处安放。

### 5.2 建议目录

- `aip2p/internal/host/`
- `aip2p/internal/plugins/`
- `aip2p/internal/themes/`

### 5.3 完成标准

- 宿主代码不再继续写进 `internal/aip2p`
- 插件和 theme 有明确归属位置

## 6. Step 2：建立插件注册入口

### 6.1 目标

让宿主具备“注册内置插件”的能力。

### 6.2 需要的最小能力

- 注册插件 ID
- 注册插件 manifest
- 校验插件基本信息
- 根据配置启用或禁用插件

### 6.3 当前第一批内置插件占位

- `news-content`
- `news-governance`
- `news-archive`
- `news-ops`

注意：

第一阶段可以先只接一个最小 `news` 占位实现，但目录和契约必须按最终拆分方向设计。

## 7. Step 3：建立 theme 注册入口

### 7.1 目标

让宿主具备“注册内置 theme”的能力。

### 7.2 需要的最小能力

- 注册 theme ID
- 注册 theme manifest
- 判断 theme 是否支持某个 page model
- 为默认应用选择默认 theme

### 7.3 第一批默认 theme

- `default-news`

## 8. Step 4：定义 page model 通道

### 8.1 目标

先把业务输出和 theme 渲染之间的接口立住。

### 8.2 第一阶段建议先定义的 page model

- `NewsFeedPageModel`
- `NewsPostPageModel`
- `NewsDirectoryPageModel`
- `ArchiveIndexPageModel`
- `NodeStatusPageModel`

不要求第一阶段全部实现完成，但名字和边界应先固定。

## 9. Step 5：把当前 `aip2p-news` UI 迁移为默认 theme 资源

### 9.1 目标

把 templates/static 从现有单体逻辑中抽离成 theme 资源。

### 9.2 需要迁移的内容

当前主要来源：

- `aip2p-news/internal/latestapp/web/templates/`
- `aip2p-news/internal/latestapp/web/static/`

### 9.3 第一阶段只要求

- 资源位置独立
- theme 能被宿主识别
- 页面渲染有正式入口

第一阶段不要求：

- 所有页面都立刻完全脱离旧逻辑

## 10. Step 6：建立 runtime API 最小骨架

### 10.1 目标

先把宿主对插件的公开面固定下来。

### 10.2 第一阶段建议先提供的分组

- `bundle`
- `store`
- `network`
- `http`
- `theme`
- `runtime`

### 10.3 可以后补的分组

- `identity`
- `worker`

不是不重要，而是第一阶段可以先把最小宿主跑起来。

## 11. Step 7：建立默认应用入口

### 11.1 目标

让 `aip2p` 宿主可以以“默认应用组合”的方式启动，而不是继续直接依赖老的单体入口。

### 11.2 目标效果

逻辑上类似：

- 宿主启动
- 宿主装载默认 news 插件组合
- 宿主装载 `default-news`
- 由宿主统一对外提供 HTTP 服务

## 12. 第一阶段暂不处理的代码拆分

以下内容建议明确放到第二阶段以后：

- `writer_policy` 彻底拆出
- delegation / revocation 单独模块化
- archive mirror 单独模块化
- sync supervisor 单独模块化
- skills 装载体系正式化

原因：

- 第一阶段核心目标是先立边界，不是一次拆完所有逻辑

## 13. 第一阶段要重点看的现有代码位置

开始真正改造时，建议优先审这几处：

### 13.1 `aip2p-core`

- `aip2p/internal/aip2p/`
- `aip2p/cmd/aip2p/main.go`

### 13.2 现有 `aip2p-news` 单体入口

- `aip2p-news/cmd/latest/main.go`
- `aip2p-news/internal/latestapp/server.go`

### 13.3 现有 UI 资源

- `aip2p-news/internal/latestapp/web/templates/`
- `aip2p-news/internal/latestapp/web/static/`

### 13.4 现有 runtime/path 和 network 逻辑

- `aip2p-news/internal/latestapp/runtimepaths.go`
- `aip2p-news/internal/latestapp/network.go`
- `aip2p-news/internal/latestapp/sync_supervisor.go`

## 14. 第一阶段验证清单

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

## 15. 第二阶段的进入条件

只有当第一阶段满足下面条件，才建议进入下一阶段：

1. 宿主边界已经存在
2. 默认 theme 已有正式身份
3. page model 通道已建立
4. 默认应用入口已从单体入口中松动出来

## 16. 当前阶段后的建议动作

在真正进入代码改造前，建议最后再做一次确认：

1. 是否接受第一阶段只先立宿主边界，不急着拆完全部业务
2. 是否接受先把当前 UI 资源独立成 `default-news` theme
3. 是否接受 `news-content` 先作为第一批主要业务模块落地

如果这 3 点确认，就可以正式进入代码改造。
