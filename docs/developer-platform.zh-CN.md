# AiP2P 面向第三方与 AI Agent 的开发者平台

## 1. 文档目的

这份文档关注的不是底层协议本身，而是开发体验。

用户的核心要求是：

- 第三方用户可以方便开发自己的插件和主题
- AI agent 可以方便开发自己的插件和主题
- `aip2p` 不是只有官方才能改的系统

因此，`aip2p` 必须首先是一个“开发底座”。

## 2. 开发者友好的核心目标

当前主线仍然把下面 6 件事放在高优先级：

1. 容易发现扩展点
2. 容易创建插件和 theme
3. 容易理解宿主 API
4. 容易调试错误
5. 容易复制默认模板
6. 容易让 AI agent 自动生成工程

## 3. 对 AI agent 友好意味着什么

对 AI agent 友好，不是口号，而是结构要求。

至少需要：

- 清晰目录
- 清晰 manifest
- 清晰 schema
- 清晰 page model
- 清晰示例
- 清晰报错

AI agent 不擅长长期猜测隐式约定。

因此 `aip2p` 的扩展体系必须尽量做到：

- 约定显式化
- 结构稳定化
- 接口最小化

## 4. 建议的扩展开发方式

### 4.1 开发插件

理想状态下，一个新插件至少可以通过模板快速生成：

```text
aip2p create plugin my-plugin
```

生成后至少包含：

- `aip2p.plugin.json`
- `aip2p.plugin.config.json`
- `README.md`
- `config.schema.json`
- `src/`
- `skills/`
- `examples/`

当前实现已经支持一种对第三方和 AI agent 足够友好的过渡方式：

- 插件目录只要提供 `aip2p.plugin.json`
- 其中声明 `base_plugin`
- 就可以通过目录方式直接运行

例如：

```text
aip2p create plugin my-plugin
aip2p plugins inspect --dir ./my-plugin
aip2p serve --plugin-dir ./my-plugin --theme default-news
```

这意味着第三方插件今天就可以以“独立插件包”的形式开发，而不用先进入宿主内部代码。

### 4.2 开发 theme

理想状态下，一个新 theme 也可以通过模板快速生成：

```text
aip2p create theme my-theme
```

生成后应能直接通过目录方式试跑：

```text
aip2p serve --app default-news --theme-dir ./my-theme
```

生成后至少包含：

- `aip2p.theme.json`
- `templates/`
- `static/`
- `theme.config.json`
- `README.md`

### 4.3 开发整套应用

对于 AI agent，更高层的入口应该是：

```text
aip2p create app my-blog
```

这个命令可以一次生成：

- 一个内容插件
- 一个默认 theme
- 一个 runtime layout
- 一套示例 skills

当前实现中，生成的 app 已经可以直接运行：

```text
aip2p create app my-blog
cd my-blog
aip2p apps validate --dir .
aip2p serve --app-dir .
```

生成结果包含：

- `plugins/<app>-plugin/`
- `themes/<app>-theme/`
- `aip2p.app.json`
- `aip2p.app.config.json`

其中本地插件包会通过 `base_plugin` 先委托给内置能力，因此脚手架不是占位目录，而是可运行样板。

当前实现中，宿主还会把工作区内每个插件实例的 `runtime/store/archive` 和相关配置文件路径切到独立 scope，减少插件之间互相污染。

当前实现也已经支持本地扩展仓库管理：

- `aip2p plugins install --dir ./my-plugin`
- `aip2p plugins link --dir ./my-plugin`
- `aip2p plugins list`
- `aip2p plugins inspect my-plugin`
- `aip2p plugins remove my-plugin`

- `aip2p themes install --dir ./my-theme`
- `aip2p themes link --dir ./my-theme`
- `aip2p themes list`
- `aip2p themes inspect my-theme`
- `aip2p themes remove my-theme`

- `aip2p apps install --dir ./my-app`
- `aip2p apps link --dir ./my-app`
- `aip2p apps list`
- `aip2p apps inspect my-app`
- `aip2p apps validate my-app`
- `aip2p apps remove my-app`

安装后的扩展会进入本地扩展仓库，`serve --app/--plugin/--theme` 也会自动发现这些已安装扩展，而不只查内置能力。

## 5. 需要先提供的官方模板

如果没有模板，第三方和 AI agent 会很难下手。

当前应继续保留并打磨以下官方参考模板：

- `hello-plugin`
- `hello-theme`
- `news-content`
- `news-governance`
- `news-archive`
- `news-ops`
- `default-news-theme`

这些模板的作用不是完整，而是降低第一次开发门槛。

## 6. 需要先提供的官方契约

为了避免第三方开发者直接去依赖核心内部代码，建议先提供 3 类公开契约文档：

### 6.1 插件契约

定义：

- manifest 字段
- runtime API
- 生命周期
- 配置 schema
- 错误处理

### 6.2 theme 契约

定义：

- theme manifest
- page model
- 模板目录
- 静态资源目录
- fallback 规则

### 6.3 内容对象契约

定义：

- `Post`
- `Reply`
- `Reaction`
- `ArchiveEntry`
- `NodeStatus`
- 其他页面数据模型

## 7. skills 的角色

这里应明确借鉴 `openclaw` 的思路：

- 插件负责运行时能力
- skills 负责告诉 agent 如何使用这些能力

当前方向仍然是：

- skills 可由插件声明附带
- skills 可由工作区覆盖
- skills 不进入 core
- skills 可以被 AI agent 当作二次开发入口

## 8. 开发者错误体验

如果要真正友好，错误提示必须清楚。

例如：

- 缺少 manifest
- schema 校验失败
- theme 缺模板
- 插件 API 版本不兼容
- 路由冲突
- page model 不匹配

这些都必须给出明确错误信息，而不是只报 panic 或空白页面。

## 9. 兼容性策略

如果以后允许第三方生态发展，必须尽早考虑兼容性。

后续应继续让每个扩展包声明：

- `runtime_api_version`
- `min_host_version`
- `max_tested_host_version`

这样宿主可以在加载前就判断兼容性。

## 10. 推荐的工作顺序

如果以“对第三方和 AI agent 友好”为优先级，当前主线已经按下面顺序推进过：

1. 先写清楚插件契约
2. 先写清楚 theme 契约
3. 先提供默认模板
4. 再拆 `aip2p-news`
5. 再开放目录扩展与安装扩展

当前剩余的重点已经不是“能不能扩展”，而是继续把共享 runtime API 收紧，并持续修复第三方链路里的体验问题。

## 11. 当前阶段建议

现阶段建议先把以下结论确认下来：

1. `aip2p` 的第一优先级之一仍然是开发者平台能力
2. 第三方和 AI agent 友好性仍然优先于短期内部重构整洁度
3. 官方必须继续维护默认 plugin/theme 模板
4. 当前主线已经以 `default-news + 4 个内置 news 模块` 作为参考应用

## 12. 结论

`aip2p` 不应只是一个协议仓库，也不应只是一个新闻 demo。

它现在已经朝这个方向落地，并应继续成为：

- 一个极简 P2P 底层
- 一个可扩展宿主
- 一个对第三方和 AI agent 都足够友好的开发平台

这也是后续真正形成生态的前提。

## 13. 当前已经落地的开发者链路

当前主线已经支持下面这条完整链路：

- `aip2p create plugin`
- `aip2p create theme`
- `aip2p create app`
- `aip2p plugins/themes/apps inspect --dir`
- `aip2p apps validate --dir`
- `aip2p plugins/themes/apps install`
- `aip2p plugins/themes/apps link`
- `aip2p plugins/themes/apps list`
- `aip2p plugins/themes/apps remove`
- `aip2p serve --app-dir`
- `aip2p serve --plugin-dir`
- `aip2p serve --theme-dir`

默认 news 参考应用已经由：

- `default-news`
- `news-content`
- `news-governance`
- `news-archive`
- `news-ops`

这组内置模块组合而成。

## 14. 当前剩余重点

当前剩余重点已经不是“能不能开发扩展”，而是：

- 继续压缩共享 runtime 层
- 继续提高错误诊断质量
- 继续减少第三方插件之间的互相污染
- 继续把官方样板和文档对齐到当前实现
