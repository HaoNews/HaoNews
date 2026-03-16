# AiP2P Themes 与当前实现

## 1. 文档目的

这份文档定义 `aip2p` 的 UI theme 体系，以及当前主线已经落地的实现边界。

目标是把 UI 变成可替换层，而不是把当前 `aip2p-news` 的界面永久绑定到系统行为上。

## 2. theme 的定位

theme 是上层显示层。

theme 的职责只有一个：

- 用不同视觉和页面结构展示同一批应用数据

theme 不负责：

- 定义消息语义
- 决定接收规则
- 处理同步逻辑
- 修改 writer policy
- 直接操作底层 bundle

theme 和功能插件是两种不同扩展类型。

它们必须保持分开：

- theme 不是功能插件
- 功能插件也不是 theme

但是 theme 可以声明：

- 自己兼容哪些插件
- 自己依赖哪些页面模型

也就是说：

- 两者分开
- 但 theme 可以依赖一个或多个特定插件组合来工作

## 3. 为什么要有 themes

用户明确希望：

- `aip2p-news` 当前 UI 作为默认模板存在
- 任何人都可以开发自己的 UI
- 同一个底层应用能力可以长成博客、论坛、购物网站、直播网站等不同形态

这决定了 UI 必须成为一等扩展点。

## 4. theme 的边界

theme 只消费功能插件输出的数据模型。

theme 不应该知道：

- store 在磁盘上的真实结构
- bundle 的内部目录布局
- sync daemon 的内部实现
- writer policy 具体文件如何合并

theme 应只拿到：

- 页面数据
- 组件数据
- theme 配置

## 5. 默认 theme 的来源

当前 `aip2p-news` UI 建议直接成为默认 theme。

建议名称：

- `default-news`

这个默认 theme 的意义是：

- 给用户一个立即可运行的默认界面
- 给第三方开发者一个最直接的可复制模板
- 给 AI agent 一个最容易改造的参考 theme

## 6. theme 的最小包结构

当前主线和第三方目录 theme 已按下面结构组织：

- `aip2p.theme.json`
- `templates/`
- `static/`
- `theme.config.json`
- `README.md`

如果想让 AI agent 也容易理解，还可以允许：

- `examples/`
- `screens/`
- `skills/`

但 theme 本身不应带业务执行代码。

## 7. theme manifest 目标

建议保留：

- `aip2p.theme.json`

最小字段建议：

- `id`
- `name`
- `version`
- `description`
- `theme_api_version`
- `entry_templates`
- `entry_static`
- `supported_page_models`
- `default_for_plugins`

这样宿主可以先知道：

- 这个 theme 给谁用
- 支持哪些页面
- 是否兼容当前插件输出模型

## 8. theme 应支持的页面类型

不同应用会有不同页面。

因此 theme 不应只面向 news。

当前 theme 体系已经围绕通用页面模型来组织，例如：

- 首页 feed
- 单内容页
- 聚合目录页
- 归档页
- 网络状态页
- 管理页

news 只是当前第一个参考应用。

## 9. theme 的配置

theme 可以有自己的配置，但只限显示相关。

例如：

- 配色
- 首页布局
- 列表密度
- 字体
- logo
- 卡片样式
- 是否显示某些 side panel

theme 配置不应允许直接改变：

- 接收规则
- 同步规则
- 权限规则
- 底层网络行为

## 10. theme 与功能插件的关系

theme 不应和某个插件死绑定。

更合理的方式是：

- 插件输出稳定 page model
- theme 声明自己支持哪些 page model

同时，theme 还可以声明：

- 自己支持哪些插件
- 自己是哪些插件组合的默认 theme

这样：

- 一个论坛插件可以套不同 theme
- 一个博客插件可以复用某些通用 theme
- 一个 news theme 以后也能演变成更像博客或杂志站

## 11. 对第三方和 AI agent 的要求

theme 系统必须特别友好。

这意味着：

- 页面模型名字清楚
- 模板目录清楚
- 静态资源目录清楚
- theme 报错要清楚
- 缺模板时要明确提示
- 提供可复制的最小 theme 示例

AI agent 最适合做的是：

- 从默认 theme 派生一个新 theme
- 修改模板
- 修改 CSS
- 调整页面布局

因此默认 theme 必须尽量像“模板工程”，而不是高度耦合业务代码的内部实现。

## 12. 当前实现状态

当前主线已经具备这些 theme 能力：

- `default-news` 作为正式默认 theme 运行
- 内置 theme 注册
- 目录 theme 加载
- `aip2p.theme.json` manifest 发现
- `supported_plugins` / `required_plugins` 校验
- `themes inspect --dir`
- `themes install/link/list/remove`

这意味着 theme 已经不再只是规划中的概念，而是宿主的一等扩展点。

## 13. 后续主题方向

等系统稳定后，可以出现多个官方参考 theme：

- `default-news`
- `default-blog`
- `default-forum`
- `default-shop`
- `default-stream`

这些 theme 不一定一开始都实现，但架构上要允许它们存在。

## 14. 当前阶段结论

现阶段已经确认：

1. 当前 `aip2p-news` UI 会变成默认 theme
2. theme 只负责显示，不负责业务逻辑
3. theme 包必须尽量简单、清晰、可复制
4. theme 必须对第三方开发者和 AI agent 友好
