# onlyRedis

<p align="center">
  <b>轻量 · 低内存 · 极速</b> — 一款原生级的 Redis 桌面管理工具<br/>
  <b>Lightweight · Low Memory · Blazing Fast</b> — a native-speed Redis desktop manager
</p>

---

## 项目介绍 | About

**onlyRedis** 是一款基于 Wails + Go + Vue 3 构建的跨平台 Redis 桌面管理工具。启动即用、内存占用极低，键值浏览与数据操作响应流畅不卡顿，为你提供原生般快速高效的 Redis 管理体验。

**onlyRedis** is a cross-platform Redis desktop manager built with Wails, Go, and Vue 3. It launches instantly with minimal memory footprint, delivering fast and fluid key browsing and data operations — a native-speed, no-bloat Redis management experience.

### 特性 | Features

- **⚡ 极速启动 | Instant Launch** — Go 编译的单一原生可执行文件，秒开无等待。Single native binary, starts in under a second.
- **🪶 超低内存 | Minimal Footprint** — 无 Electron 依赖，后台内存占用仅几十 MB。No Electron overhead, idle memory usage as low as tens of MB.
- **🚀 流畅交互 | Fluid UX** — SCAN 分页 + Pipeline 批量查询，百万级 Key 依然丝滑。Pagination via SCAN + Pipeline batching, smooth even with millions of keys.
- **🌳 层级 Key 浏览 | Hierarchical Key Browser** — 以 `:` 分隔符构建目录树，支持懒加载，大型实例从容应对。Namespace tree via `:` delimiters with lazy loading for large instances.
- **📊 实时监控面板 | Live Monitoring** — QPS、命中率、内存、CPU、网络 IO / 连接数一应俱全。QPS, hit rate, memory, CPU, network IO, and client count at a glance.
- **🔐 安全连接 | Secure Connections** — 支持 SSH 隧道直连内网 Redis，连接配置 AES-256-GCM 加密存储。SSH tunnel support for internal-network Redis, credentials encrypted via AES-256-GCM.
- **🗃️ 全类型支持 | All Data Types** — String / Hash / List / Set / ZSet 五种数据类型完整 CRUD 操作。Full CRUD for all five Redis data types.
- **🔄 多数据库切换 | Multi-DB Switching** — 支持 Redis 0-15 号数据库自由切换。Seamless switching across DB 0-15.
- **🎨 暗色主题 | Dark Theme UI** — Naive UI + Tailwind CSS 构建的深色界面，长时间使用更护眼。Dark theme with Naive UI + Tailwind CSS, easy on the eyes.
- **🖥️ 跨平台 | Cross-Platform** — 同时支持 macOS (Apple Silicon / Intel) 与 Windows (x64)。macOS (ARM + Intel) and Windows (x64) support.

## 技术栈 | Tech Stack

| 层 Layer   | 技术 Technology                            |
| ---------- | ------------------------------------------ |
| 后端 Backend  | Go 1.23+, [Wails v2](https://wails.io/), [go-redis v9](https://github.com/redis/go-redis) |
| 前端 Frontend | Vue 3, TypeScript, [Naive UI](https://www.naiveui.com/), Tailwind CSS, ECharts |
| 安全 Security | AES-256-GCM 连接凭据加密, SSH Tunnel |
| 构建 Build   | Wails CLI 多平台构建, Vite |

## 快速开始 | Quick Start

### 环境要求 | Prerequisites

- Go 1.23+
- Node.js 18+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation): `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### 开发模式 | Development

```bash
# 安装前端依赖
cd frontend && npm install && cd ..

# 启动开发服务器（热重载）
wails dev
```

### 构建 | Build

```bash
# 构建当前平台的可执行文件
wails build

# 多平台构建（macOS ARM/Intel + Windows x64）
./build.sh

# 清理旧制品后重新构建
./build.sh --clean
```

构建产物位于 `build/bin/` 目录：
- `onlyRedis-arm64.app` — macOS Apple Silicon
- `onlyRedis-amd64.app` — macOS Intel
- `onlyRedis-amd64.exe` — Windows x64

## 项目结构 | Project Structure

```
onlyRedis/
├── main.go               # 应用入口，Wails 配置
├── app.go                # 前后端方法绑定
├── services/
│   ├── redis_service.go  # Redis 核心操作（连接、SCAN、CRUD）
│   ├── monitor_service.go # 实时监控数据采集
│   └── config_store.go   # 连接配置加密存储
├── frontend/             # Vue 3 前端项目
│   └── src/
├── build.sh              # 多平台打包脚本
└── wails.json            # Wails 项目配置
```

## 截图 | Screenshots

> 即将推出 | Coming soon

## License

MIT
