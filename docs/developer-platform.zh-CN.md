# AiP2P 面向第三方与 AI Agent 的开发者平台设计草案

## 1. 文档目的

这份文档关注的不是底层协议本身，而是开发体验。

用户的核心要求是：

- 第三方用户可以方便开发自己的插件和主题
- AI agent 可以方便开发自己的插件和主题
- `aip2p` 不是只有官方才能改的系统

因此，未来的 `aip2p` 必须首先是一个“开发底座”。

## 2. 开发者友好的核心目标

未来设计必须把下面 6 件事放在高优先级：

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

因此 `aip2p` 未来的扩展体系必须尽量做到：

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

## 5. 需要先提供的官方模板

如果没有模板，第三方和 AI agent 会很难下手。

因此建议未来至少保留以下官方参考模板：

- `hello-plugin`
- `hello-theme`
- `news-plugin`
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

因此建议未来：

- skills 可由插件声明附带
- skills 可由工作区覆盖
- skills 不进入 core
- skills 可以被 AI agent 当作二次开发入口

## 8. 开发者错误体验

未来如果要真正友好，错误提示必须清楚。

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

建议未来每个扩展包声明：

- `runtime_api_version`
- `min_host_version`
- `max_tested_host_version`

这样宿主可以在加载前就判断兼容性。

## 10. 推荐的工作顺序

如果以“对第三方和 AI agent 友好”为优先级，未来改造顺序建议如下：

1. 先写清楚插件契约
2. 先写清楚 theme 契约
3. 先提供默认模板
4. 再拆 `aip2p-news`
5. 最后再开放生态

如果顺序反过来，先拆代码再补契约，最后只会得到“内部可维护”，却不一定“对外可扩展”。

## 11. 当前阶段建议

现阶段建议先把以下结论确认下来：

1. `aip2p` 的第一优先级之一是开发者平台能力
2. 第三方和 AI agent 友好性优先于短期内部重构整洁度
3. 官方必须提供默认 plugin/theme 模板
4. `aip2p-news` 未来既是默认参考应用，也是第三方学习模板

## 12. 结论

未来的 `aip2p` 不应只是一个协议仓库，也不应只是一个新闻 demo。

它应该成为：

- 一个极简 P2P 底层
- 一个可扩展宿主
- 一个对第三方和 AI agent 都足够友好的开发平台

这也是后续真正形成生态的前提。
