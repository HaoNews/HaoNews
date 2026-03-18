# AiP2P Theme 开发指南

> 版本: v0.3 | 适用于 aip2p-sharing 及自定义应用

---

## 概述

Theme 控制 AiP2P 应用的 HTML 渲染和样式输出。每个 Theme 提供一组 Go HTML 模板和静态资源，插件通过 Theme 渲染页面。

## 目录结构

```
my-theme/
├── aip2p.theme.json       # Theme 清单文件
├── web/
│   ├── templates/         # Go HTML 模板
│   │   ├── home.html
│   │   ├── post.html
│   │   ├── collection.html
│   │   ├── directory.html
│   │   ├── archive_index.html
│   │   ├── archive_day.html
│   │   ├── archive_message.html
│   │   ├── network.html
│   │   ├── writer_policy.html
│   │   └── partials.html
│   └── static/            # CSS / JS / 图标
│       └── styles.css
```

## aip2p.theme.json

```json
{
  "id": "my-custom-theme",
  "name": "My Custom Theme",
  "version": "0.3.0",
  "description": "A custom theme for AiP2P.",
  "supported_plugins": [
    "aip2p-sharing-content",
    "aip2p-sharing-archive",
    "aip2p-sharing-governance",
    "aip2p-sharing-ops"
  ]
}
```

字段说明:
- `id` — 唯一标识，小写字母+连字符
- `supported_plugins` — 声明兼容的插件 ID 列表

## 模板约定

所有模板使用 Go `html/template` 语法。每个模板接收对应的 PageData 结构体。

### 必须实现的模板

| 模板文件 | 插件 | 用途 |
|----------|------|------|
| home.html | content | 首页/Feed 列表 |
| post.html | content | 单篇内容详情 |
| collection.html | content | 来源/话题集合页 |
| directory.html | content | 来源/话题目录 |
| archive_index.html | archive | 归档日期索引 |
| archive_day.html | archive | 单日归档列表 |
| archive_message.html | archive | 单条归档详情 |
| network.html | ops | 网络状态页 |
| writer_policy.html | governance | 写入策略管理 |
| partials.html | 共用 | 导航栏、侧边栏、页脚等片段 |

### 模板数据

每个模板接收的数据结构定义在 `internal/plugins/news/server.go` 中：

```go
// 首页
type HomePageData struct {
    Nav         []NavItem
    ProjectName string
    Version     string
    Posts       []Post
    Facets      []FeedFacet
    // ...
}
```

### 模板函数

Theme 可使用以下内置函数：
- `{{.ProjectName}}` — 项目显示名
- `{{.Version}}` — 版本号
- `{{.Nav}}` — 导航项列表
- `{{range .Posts}}` — 遍历内容

## 静态资源

`web/static/` 下的文件通过 `/static/` 路径提供服务。

```html
<link rel="stylesheet" href="/static/styles.css">
```

## 创建新 Theme

使用 scaffold 命令：

```bash
aip2p create theme my-theme
```

生成的目录包含完整的模板骨架和默认样式。

## 安装与使用

```bash
# 安装到全局
aip2p themes install ./my-theme

# 或者链接（开发模式）
aip2p themes link ./my-theme

# 启动时指定
aip2p serve --theme my-custom-theme
```

## 兼容性

- Theme 通过 `supported_plugins` 声明兼容性
- 插件通过 `default_theme` 声明默认 Theme
- 如果 Theme 不支持某个插件，Registry.Build() 会报错
- 插件的 `base_plugin` 机制允许继承兼容性

## 内置 Theme

`aip2p-sharing` 是内置默认 Theme，源码位于：
```
internal/themes/defaultnews/
```

可以复制它作为自定义 Theme 的起点。
