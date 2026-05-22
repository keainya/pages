# Pages — 静态页面管理系统

一个基于 Go + Gin 的轻量级静态页面管理系统，支持动态管理 URL 路径到 HTML 内容的映射，通过统一的账户系统进行认证。

---

## 目录

- [1. 功能概述](#1-功能概述)
- [2. 服务配置](#2-服务配置)
- [3. 认证集成](#3-认证集成)
- [4. 管理页面](#4-管理页面)
- [5. 页面数据模型](#5-页面数据模型)
- [6. URL 路由解析规则](#6-url-路由解析规则)
- [7. API 设计](#7-api-设计)
- [8. 技术栈](#8-技术栈)
- [9. 项目结构](#9-项目结构)
- [10. 运行与构建](#10-运行与构建)

---

## 1. 功能概述

本系统是一个**静态页面管理平台**，允许管理员通过 Web 管理界面动态添加、编辑、删除静态页面。每个页面由**唯一路径**和**HTML 内容**组成，访问对应路径时直接返回该 HTML 内容。

核心特性：

- ✅ 基于 URL 路径的动态页面路由
- ✅ 管理后台 (SPA) 的 CRUD 操作
- ✅ 对接统一账户系统进行认证与授权
- ✅ 智能路径容错（`xxx` ↔ `xxx.html`）

## 2. 服务配置

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| 监听端口 | `8082` | HTTP 服务端口 |
| 数据库 | SQLite (`data/pages.db`) | 存储页面数据 |
| 前端文件 | 内嵌 (`embed.FS`) | 编译时嵌入二进制 |

## 3. 认证集成

本系统对接 [账户系统](../docs/账户系统API.md)，所有管理操作需登录后方可使用。

### 3.1 账户系统信息

| 项目 | 值 |
|------|-----|
| Base URL | `https://account.takemeto.icu` |

### 3.2 认证方式

使用 **Session Cookie** 方式：

1. 用户在本系统的登录页提交用户名/密码
2. 后端向账户系统 `POST /api/auth/login` 发起请求
3. 账户系统返回用户信息并设置 Session Cookie
4. 前端携带 Cookie 调用本系统 API，后端通过 `GET /api/auth/me` 验证身份
5. 登出时调用 `POST /api/auth/logout` 清除 Session

### 3.3 认证相关的 API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/auth/register` | 注册（首个用户自动为管理员） |
| POST | `/api/auth/login` | 登录并获取 Session |
| POST | `/api/auth/logout` | 登出 |
| GET | `/api/auth/me` | 获取当前用户信息 |

### 3.4 权限要求

- **普通用户** (`user`)：可浏览已发布的静态页面
- **管理员** (`admin`)：可访问管理后台，管理所有静态页面

## 4. 管理页面

### 4.1 入口

```
http://localhost:8082/#/ctrl
```

### 4.2 功能

管理页面提供对静态页面的完整 CRUD 操作：

| 功能 | 说明 |
|------|------|
| **页面列表** | 展示所有已创建的页面路径（支持分页） |
| **新建页面** | 指定 URL 路径和 HTML 内容 |
| **编辑页面** | 修改已存在页面的 HTML 内容 |
| **删除页面** | 移除页面映射 |
| **预览** | 在新标签页打开对应路径预览效果 |

### 4.3 页面路由

前端基于 Hash SPA 的路由结构：

| 路由　　　　 | 页面　　　　 | 需认证　　　 |
| --------------| --------------| :------------:|
| `#/login`　　| 登录页　　　 | 否　　　　　 |
| `#/register` | 注册页　　　 | 否　　　　　 |
| `#/`　　　　 | 公开页面浏览 | 否　　　　　 |
| `#/ctrl`　　 | 管理后台　　 | 是（管理员） |

## 5. 页面数据模型

### 5.1 存储结构

```
url_path  →  html_content
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 主键，自增 |
| `url_path` | string | 页面路径（唯一索引），如 `about`、`docs/getting-started` |
| `html_content` | text | 页面对应的完整 HTML 内容 |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

### 5.2 路径规范

- `url_path` **不含前导** `/`
- `url_path` **不含** `.html` 后缀（系统自动适配）
- 示例：`index`、`about`、`docs/intro`、`projects/my-app`

## 6. URL 路由解析规则

当用户访问 URL 时，系统按以下优先级查找页面内容：

```
用户访问: /about
  ├─ 查找 url_path = "about"        → 有 → 返回 about 的 HTML
  └─ 无
     └─ 查找 url_path = "about.html"  → 有 → 返回 about.html 的 HTML
        └─ 无 → 返回 404

用户访问: /about.html
  ├─ 查找 url_path = "about.html"  → 有 → 返回
  └─ 无 → 返回 404

用户访问: /
  ├─ 查找 url_path = "index"        → 有 → 返回 index 的 HTML
  └─ 无
     └─ 查找 url_path = "index.html"  → 有 → 返回 index.html 的 HTML
        └─ 无 → 返回 404
```

### 6.1 规则总结

| 请求路径 | 优先级1 | 优先级2 | 说明 |
|----------|---------|---------|------|
| `/` | `index` | `index.html` | 首页兜底 |
| `/xxx` | `xxx` | `xxx.html` | 无后缀时补充 `.html` 查找 |
| `/xxx.html` | `xxx.html` | — | 精确匹配，不做转换 |

### 6.2 路由优先级

整体请求处理优先级：

1. **API 路由**（`/api/*`、`/status` 等）
2. **管理页面 SPA**（前端静态资源，`/index.html`）
3. **动态页面**（按上述规则匹配 `pages` 表，返回 HTML 内容）
4. **内嵌静态文件**（`web/` 目录下的 CSS、JS 等）
5. **404 Not Found**

## 7. API 设计

### 7.1 统一响应格式

```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

- `code = 0`：成功
- `code != 0`：错误，详见 `msg`

### 7.2 页面管理 API（需管理员 Session）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/pages` | 获取页面列表（分页：`?page=1&page_size=20`） |
| GET | `/api/pages/:id` | 获取单个页面详情 |
| POST | `/api/pages` | 新建页面 |
| PUT | `/api/pages/:id` | 更新页面 |
| DELETE | `/api/pages/:id` | 删除页面 |

请求/响应示例：

**新建页面** `POST /api/pages`
```json
{
  "url_path": "about",
  "html_content": "<h1>关于我们</h1><p>这是一个关于页面。</p>"
}
```

**获取页面列表** `GET /api/pages?page=1&page_size=20`
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "items": [
      {
        "id": 1,
        "url_path": "about",
        "html_content": "<h1>关于我们</h1>",
        "created_at": "2026-05-23T00:00:00Z",
        "updated_at": "2026-05-23T00:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 20
  }
}
```

### 7.3 认证代理 API

本系统作为账户系统的代理层，不对认证逻辑做自行实现，仅转发请求：

| 本系统路由 | 转发至账户系统 |
|------------|---------------|
| `POST /api/auth/register` | `POST {account}/api/auth/register` |
| `POST /api/auth/login` | `POST {account}/api/auth/login` |
| `POST /api/auth/logout` | `POST {account}/api/auth/logout` |
| `GET /api/auth/me` | `GET {account}/api/auth/me` |

## 8. 技术栈

| 层级 | 技术 | 说明 |
|------|------|------|
| **后端框架** | Gin 1.x | Go HTTP Web 框架 |
| **ORM** | GORM + SQLite | 对象关系映射 + 嵌入式数据库 |
| **Session** | gin-contrib/sessions | Cookie 存储的会话管理 |
| **CORS** | gin-contrib/cors | 跨域支持 |
| **前端** | 原生 HTML/CSS/JS | 零依赖 SPA（Hash 路由） |
| **文件嵌入** | Go embed | 编译时嵌入前端静态资源 |

## 9. 项目结构

```
pages/
├── main.go                 # 程序入口
├── embed.go                # 静态文件嵌入
├── go.mod / go.sum         # Go 模块依赖
├── docs/
│   └── 账户系统API.md       # 账户系统 API 文档
├── object/
│   ├── database.go         # 数据库初始化
│   └── page.go             # 页面数据模型 (新增)
├── router/
│   └── default.go          # 路由注册
├── service/
│   ├── response.go         # 统一响应结构体
│   ├── default.go          # 通用处理器
│   ├── auth.go             # 认证代理处理器 (新增)
│   └── page.go             # 页面 CRUD 处理器 (新增)
├── utils/
│   └── fs.go               # 工具函数
└── web/
    ├── index.html           # 前端入口
    ├── favicon.ico
    ├── css/
    │   └── style.css        # 样式
    ├── js/
    │   └── app.js           # 前端逻辑
    └── style.md
```

> 标注 "(新增)" 的文件为本次需求所需的增量开发项。

## 10. 运行与构建

### 10.1 开发环境

```bash
# 安装依赖
go mod tidy

# 运行（默认端口 8082）
go run . -port 8082
```

### 10.2 构建

```bash
# 编译二进制（前端文件已嵌入）
go build -o pages .

# 运行
./pages -port 8082
```

### 10.3 访问

- 公开页面：`http://localhost:8082/`
- 管理后台：`http://localhost:8082/#/ctrl`
- 状态检查：`http://localhost:8082/status`
