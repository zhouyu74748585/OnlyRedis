---
name: redis-desktop-manager
overview: 使用 Go + Wails v3 + Vue3 构建一个全平台 Redis 桌面管理工具，支持连接管理（含SSH隧道）、Key 树形浏览/编辑、多数据类型可视化、服务器实时监控。
design:
  architecture:
    framework: vue
  styleKeywords:
    - Dark Theme
    - Professional Tool
    - IDE Layout
    - High Contrast
    - Tech Blue Accent
    - Minimalist
  fontSystem:
    fontFamily: Inter, PingFang SC, Microsoft YaHei, sans-serif
    heading:
      size: 18px
      weight: 600
    subheading:
      size: 14px
      weight: 500
    body:
      size: 13px
      weight: 400
  colorSystem:
    primary:
      - "#4A90D9"
      - "#3A7BD5"
      - "#2B6CB0"
    background:
      - "#1E1E2E"
      - "#252536"
      - "#2D2D3F"
    text:
      - "#E0E0E8"
      - "#A0A0B8"
      - "#707088"
    functional:
      - "#4CAF50"
      - "#F44336"
      - "#FF9800"
      - "#2196F3"
todos:
  - id: init-project
    content: 初始化项目结构：Go module + Wails v3 脚手架 + Vue3/Vite/NaiveUI 前端工程搭建
    status: completed
  - id: connection-management
    content: 实现连接管理模块：Go 端 Redis 连接池 + SSH 隧道 + 配置加密存储 + 前端连接面板 UI
    status: completed
    dependencies:
      - init-project
  - id: key-tree-browse
    content: 实现 Key 树形浏览与 CRUD：Go 端 SCAN 分页扫描 + 前端虚拟滚动树 + 搜索过滤 + 增删改操作
    status: completed
    dependencies:
      - connection-management
  - id: data-type-viewers
    content: 实现五大数据类型可视化编辑器：String/Hash/List/Set/ZSet 各类型读写操作与前端组件
    status: completed
    dependencies:
      - key-tree-browse
  - id: server-monitor
    content: 实现服务器监控模块：Go 端 INFO 采集 + 环形缓冲区 + 前端仪表盘卡片与 QPS 折线图
    status: completed
    dependencies:
      - connection-management
  - id: integration-build
    content: 集成联调、暗色主题统一、跨平台构建配置验证，使用 [skill:frontend-verify] 验证前端构建
    status: completed
    dependencies:
      - data-type-viewers
      - server-monitor
  - id: documentation
    content: 使用 [skill:session-summary-docs] 输出项目总结文档，记录架构决策、构建配置与使用说明
    status: completed
    dependencies:
      - integration-build
---

## 产品概述

onlyRedis 是一款专注 Redis 管理的跨平台桌面应用。采用 Go + Wails v3 架构，以极快的启动速度、极低的内存占用和极小的制品体积为核心竞争力，为开发者提供高效、轻量的 Redis 日常管理工具。

## 核心功能

### 连接管理

- 支持多 Redis 实例连接配置（Host/Port/Password/DB），连接列表持久化存储
- 支持 SSH 隧道连接（通过跳板机连接内网 Redis）
- 连接状态实时显示（已连接/断开/重连中），支持一键重连
- 连接信息加密存储（密码/密钥使用 AES 加密落盘）

### Key 浏览与管理

- 树形结构展示 Redis Key，支持按前缀层级折叠展开
- 虚拟滚动优化，支撑百万级 Key 量场景无卡顿
- 支持 Key 的增删改查、设置 TTL、重命名、复制
- 实时搜索过滤 Key，支持模糊匹配

### 数据类型可视化

- String：文本编辑器展示，支持 JSON/Base64 格式化预览
- Hash：表格展示 Field-Value，支持分页加载、新增/编辑/删除
- List：列表展示元素，支持左右插入、按索引操作
- Set：集合展示成员，支持添加/移除、集合运算可视化
- ZSet：有序集合展示，支持按 Score 排序、批量操作

### 服务器监控

- 实时仪表盘展示内存使用、连接数、命中率、QPS
- CPU 使用率、网络吞吐量、Key 数量趋势
- INFO 全量信息展示，支持自动刷新（可配置间隔）

## 技术栈选型

| 层级 | 技术 | 版本 | 选型理由 |
| --- | --- | --- | --- |
| 后端语言 | Go | 1.22+ | 编译为原生二进制，启动极快，内存模型高效 |
| GUI 框架 | Wails | v3 | Go+Web 前端混合，系统 WebView 渲染，无 Chromium 依赖，产物极小 |
| 前端框架 | Vue3 + TypeScript | 3.4+ | Composition API 简洁高效，TypeScript 保障类型安全 |
| 构建工具 | Vite | 5+ | 极速 HMR，Tree-shaking 优化产物 |
| UI 组件库 | Naive UI | 2.38+ | Tree-shakable，体积可控，虚拟滚动/表格等高级组件齐全 |
| Redis 客户端 | go-redis | v9 | 生态最成熟，支持集群/哨兵/管道/订阅全特性 |
| SSH 隧道 | golang.org/x/crypto | latest | Go 官方加密库，无需外部依赖 |
| 状态管理 | Pinia | 2+ | Vue3 官方推荐，类型推导完整 |


## 实现方案

### 整体策略

采用 **分层架构 + 桥接模式**：Go 层负责所有 I/O 密集型操作（Redis 通信、SSH 隧道、文件加密存储），Vue3 层负责 UI 渲染和用户交互，两者通过 Wails 的 Go-JS 双向绑定通信。Go 层暴露 struct 方法，Wails 编译时自动生成前端可调用的 Promise 接口。

### 关键性能设计

**启动速度优化（< 500ms）**

- Go 二进制启动即初始化 Wails 运行时，无 JIT 预热
- 前端资源内嵌进二进制（embed.FS），单文件分发，无 HTTP 加载开销
- 延迟初始化：SSH 连接池、Redis 客户端按需创建，启动时不建立任何外部连接
- 连接配置 JSON 文件（通常 < 1KB），启动时毫秒级解析

**内存优化（空载 < 30MB）**

- Key 树采用懒加载+虚拟滚动，仅渲染可视区域 DOM 节点
- 海量 Key 场景下 Go 端分页扫描（SCAN 命令），前端分批接收，不一次性加载全部数据
- 监控数据采用环形缓冲区，保留最近 300 个采样点，超出自动淘汰
- Naive UI 按需引入，配合 Vite Tree-shaking 剔除未使用组件代码

**制品体积优化（macOS < 15MB）**

- Go 编译使用 `-ldflags="-s -w"` 去除调试符号和 DWARF 信息
- UPX 压缩二进制（可选），macOS 下约压缩至原体积 30%-40%
- 前端资源经 Vite 打包后 gzip 内嵌，不含 node_modules
- 系统 WebView 复用操作系统内置渲染引擎，零额外运行时依赖

### 数据架构

```
┌─────────────────────────────────────────────────┐
│                   Wails Runtime                    │
│                                                    │
│  ┌──────────────┐          ┌──────────────────┐  │
│  │   Vue3 App   │◄─IPC──►│   Go Backend     │  │
│  │  (WebView)   │          │                  │  │
│  │              │          │ ┌──────────────┐ │  │
│  │  Components  │          │ │ RedisService │ │  │
│  │  Stores      │          │ │ ├─Connection │ │  │
│  │  Router      │          │ │ ├─KeyOps     │ │  │
│  │              │          │ │ ├─TypeOps    │ │  │
│  │  Naive UI    │          │ │ └─Monitor    │ │  │
│  └──────────────┘          │ └──────┬───────┘ │  │
│                             │ ┌──────┴───────┐ │  │
│                             │ │ SSHService   │ │  │
│                             │ │ ConfigStore  │ │  │
│                             │ └──────────────┘ │  │
│                             └──────────────────┘  │
│                                         │          │
│                                 go-redis/ssh        │
│                                         │          │
│                                  Redis Server      │
└─────────────────────────────────────────────────┘
```

### 通信协议

Go 层暴露方法签名遵循统一约定：

```
// Wails 绑定方法：返回 (data, error) 或 error
func (a *App) GetKeys(connId string, pattern string, cursor uint64, count int64) ([]KeyInfo, uint64, error)
func (a *App) GetStringValue(connId string, key string) (string, error)
func (a *App) SetStringValue(connId string, key string, value string, ttl int64) error
func (a *App) StartMonitor(connId string, interval int) error
func (a *App) StopMonitor(connId string) error
```

## 实现注意事项

### 性能

- Key 树加载：首次仅加载一级 prefix，展开节点时按需 SCAN 子节点；每次 SCAN COUNT 设为 100，前端虚拟滚动高度固定 36px/row
- 监控数据推送：Go 端每秒采集一次 INFO，前端使用 `requestAnimationFrame` 批量更新图表，避免频繁 DOM 操作
- Hash/ZSet 大 Key：前端分页加载（每页 100 条），Go 端使用 HSCAN/ZSCAN 游标分页，避免 HGETALL/ZRANGE 全量阻塞

### 安全

- 连接密码使用 AES-256-GCM 加密后存入本地 JSON 文件，密钥通过机器指纹（hostname+platform+username 组合）派生
- SSH 私钥内容同样加密存储，内存中用完即清理
- 前端禁止暴露原始密码给 DOM，Go 端返回连接列表时密码字段返回脱敏字符串 "****"

### 兼容性

- go-redis v9 兼容 Redis 6.0/7.0，对旧版本 Redis 做降级处理（如 RESP2 协议）
- WebView 兼容：macOS 使用 WKWebView，Windows 使用 WebView2（Win10+ 自带），Linux 使用 WebKitGTK
- 前端 Naive UI 支持暗色/亮色主题切换，默认暗色（数据库管理工具主流偏好）

## 设计风格

采用 **Dark Professional Tool** 风格，定位为专业开发者效率工具。以深色背景为主基调，搭配高对比度科技蓝作为强调色，营造沉浸式数据管理体验。布局采用经典的 **三栏式 IDE 布局**——左侧连接与导航栏、中间数据编辑区、右侧可折叠监控面板，符合开发者对数据库管理工具的直觉预期。

## 页面设计

### 主界面 — 三栏布局

- **左侧栏（宽 260px，可拖拽调整）**
- 顶部：连接管理区域，卡片式连接列表，每个连接显示名称+状态圆点（绿=已连接/红=断开），右键菜单支持编辑/删除/重连
- 底部：选中连接的 Key 树，使用 Naive UI Tree 组件，带虚拟滚动，输入框实时搜索过滤 Key
- 连接列表和 Key 树之间用分割线清晰分隔

- **中央区域（自适应宽度）**
- 顶部 Tabs 栏：为每个打开的 Key 创建独立 Tab，支持关闭/切换
- 内容区：根据数据类型动态切换 Viewer 组件——String 使用代码编辑器（Monaco-Editor-Lite 轻量版）、Hash 使用 Naive UI Table（分页）、List/Set/ZSet 使用虚拟列表
- 工具栏：刷新按钮、TTL 编辑、删除按钮、复制按钮，统一放在 Tab 栏右侧

- **右侧面板（宽 320px，可折叠）**
- 监控仪表盘：顶部三个指标卡片（内存使用率/连接数/命中率），中部为 QPS 折线图，底部显示 INFO 全量信息表格
- 自动刷新指示器和手动刷新按钮

### 连接配置弹窗

- 模态对话框，分区填写：基本信息（名称/Host/Port/Password/DB）、SSH 隧道（开关+Host/Port/User/Key 文件选择）、高级（连接超时/重试次数）
- 测试连接按钮，实时反馈连接结果

### 新增 Key 弹窗

- 输入 Key 名称、选择数据类型、初始 TTL，创建后自动打开编辑器

## Agent Extensions

### Skill

- **session-summary-docs**
- 用途：在每个开发阶段结束后，将实现总结、技术决策和验证结果记录到 docs 目录
- 预期成果：生成带时间戳的阶段总结文档，便于项目回溯与知识沉淀

- **frontend-verify**
- 用途：在前端代码改动完成后执行 `npm run build` 和 `type-check`，确保构建无报错
- 预期成果：每次前端修改后验证通过，确保 TypeScript 类型安全和 Vite 构建正常