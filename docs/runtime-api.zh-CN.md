# AiP2P 宿主 Runtime API 与当前实现

## 1. 文档目的

这份文档定义 `aip2p-host` 暴露给插件的 runtime API 边界，以及当前主线已经落地的实现方向。

目标是避免插件直接依赖 core 内部实现，同时给第三方和 AI agent 一个稳定、低歧义的开发接口。

## 2. 为什么需要 Runtime API

如果没有统一 runtime API，插件作者就会直接 import 核心内部代码。

这样会带来几个问题：

- core 很难演进
- 插件版本很容易被内部重构打断
- 第三方插件无法稳定开发
- AI agent 必须读大量内部源码才能写插件

因此宿主必须建立公开 API，而不是让插件直接耦合 `internal` 包。

## 3. Runtime API 的设计目标

当前 runtime API 仍然按下面目标收紧：

1. 尽量小
2. 尽量稳定
3. 尽量显式
4. 尽量分组
5. 能覆盖插件真正需要的场景

## 4. API 分组建议

当前 runtime API 建议继续按 8 组来收敛。

### 4.1 `bundle`

负责 bundle/message 读取与校验。

建议能力：

- 读取 message
- 读取 body
- 校验 signature
- 读取 infohash / magnet
- 按 project 或 kind 过滤 bundle

### 4.2 `store`

负责访问本地 store 的公共能力。

建议能力：

- 获取 store root
- 列出 bundle
- 查询 bundle
- 读取 torrent refs

### 4.3 `identity`

负责身份和签名相关公共能力。

建议能力：

- 读取 identity
- 校验 origin signature
- 校验 delegation / revocation

### 4.4 `network`

负责公共 network 信息。

建议能力：

- 读取 bootstrap config
- 获取 sync status
- 获取 LAN/bootstrap 状态

### 4.5 `http`

负责插件向宿主注册 HTTP 路由。

建议能力：

- 注册页面路由
- 注册 API 路由
- 注册静态资源挂载点

### 4.6 `theme`

负责 theme 交互。

建议能力：

- 注册页面模型
- 请求 theme 渲染某个页面
- 检查 theme 是否支持某个 page model

### 4.7 `runtime`

负责插件自己的运行时路径和配置。

建议能力：

- 获取 runtime root
- 获取插件专属目录
- 读取插件配置
- 写入插件状态

### 4.8 `worker`

负责后台任务。

建议能力：

- 注册后台 worker
- 注册 supervisor 任务
- 上报 worker 健康状态

## 5. 插件不应直接拿到的能力

为了保持边界，插件不应直接拿到：

- core 未公开内部包
- 宿主的全局可变内部状态
- 任意文件系统写权限
- 任意网络访问能力

这些都应通过 runtime API 控制和收敛。

## 6. 建议的最小接口方向

下面是逻辑示意，不代表最终代码语言形式：

```text
runtime.bundle.list(project)
runtime.bundle.load(infohash)
runtime.bundle.verify(message)

runtime.store.root()
runtime.store.query(filter)

runtime.identity.verifyOrigin(message)
runtime.identity.verifyDelegation(childKey, parentKey)

runtime.network.bootstrap()
runtime.network.syncStatus()

runtime.http.page(path, handler)
runtime.http.api(path, handler)

runtime.theme.render(modelName, data, options)
runtime.theme.supports(modelName)

runtime.runtime.root()
runtime.runtime.pluginRoot(pluginId)
runtime.runtime.config(pluginId)

runtime.worker.register(workerId, spec)
runtime.worker.status(workerId, state)
```

## 7. Page Model 契约

theme 和插件能否彻底解耦，关键在 page model。

当前已经明确要把 page model 作为 runtime API 的正式组成部分。

例如：

- 插件声明自己输出什么 model
- theme 声明自己支持什么 model
- 宿主负责匹配两者

这样 theme 就不需要知道插件内部细节。

## 8. 面向第三方和 AI agent 的要求

runtime API 必须做到：

- 有版本号
- 有最小示例
- 有错误码或明确错误信息
- 有模板插件示例

AI agent 最怕的是：

- API 名称模糊
- 输入输出不稳定
- 看起来都能用，运行时才发现不兼容

当前主线已经开始实现其中一部分，但仍然要继续先文档化、再收紧实现。

## 9. 与 `openclaw` 理念的关系

这里应借鉴 `openclaw` 的关键思路：

- 插件不要直接依赖宿主内部实现
- 宿主提供稳定 runtime surface
- manifest 先于代码执行

但 `aip2p` 不必完全照搬 `openclaw` 的 in-process 运行模型。

对于 `aip2p`，更重要的是：

- runtime API 稳定
- 故障边界清楚
- 第三方插件可控制风险

## 10. 当前实现状态

当前主线已经落下这些边界：

- 宿主、插件、theme 已经分层
- 内置 news 样板插件已经通过共享 runtime 层复用能力
- 第三方目录插件已经可以通过 manifest 和 `base_plugin` 接到宿主
- theme 与插件之间已经通过支持模型和依赖插件做兼容校验

## 11. 当前阶段结论

现阶段已经确认：

1. 插件必须通过 runtime API 访问宿主
2. 插件不能直接依赖 core 内部包
3. runtime API 应按 bundle/store/identity/network/http/theme/runtime/worker 分组
4. page model 应成为 theme 与插件之间的正式契约
