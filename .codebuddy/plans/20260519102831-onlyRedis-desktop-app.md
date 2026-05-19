# 主题：onlyRedis-desktop-app

## 记录

### 2026-05-19 10:28:31

## 本次目标
构建 onlyRedis —— 一款专注 Redis 管理的跨平台桌面应用，优先级：启动速度 > 内存占用低 > 制品体积小。

## 技术选型
| 层级 | 技术 | 说明 |
|------|------|------|
| 后端 | Go 1.26 + Wails v2.12 | 编译为原生二进制，单文件分发 |
| 前端 | Vue3 + TypeScript + Vite 5 | Composition API，类型安全 |
| UI 框架 | Naive UI 2.41 + TailwindCSS 3.4 | Tree-shakable，暗色主题 |
| Redis 客户端 | go-redis v9 | 生态最成熟，全特性支持 |
| 图表 | ECharts 5 + vue-echarts | 监控数据可视化 |

## 关键实现

### Go 后端 (`services/`)
- **redis_service.go**: 完整的 Redis 操作封装 - 连接池管理、SCAN 分页、五大数据类型 CRUD、SSH 隧道支持
- **config_store.go**: 连接配置持久化，采用 AES-256-GCM 加密保护密码，密钥由机器指纹派生
- **monitor_service.go**: 周期性 INFO 采集，解析 20+ 指标，QPS 通过瞬时 ops 差值计算

### Vue 前端 (`frontend/src/`)
- **三栏 IDE 布局**: 左栏连接+Key树 / 中栏数据查看器 / 右栏监控面板
- **5 种数据类型可视化**: String(编辑器) / Hash(分页表格) / List(虚拟列表) / Set / ZSet
- **实时监控**: 指标卡片 + ECharts QPS 折线图 + INFO 全量信息
- **暗色专业主题**: 深色底色 + 科技蓝强调色 + 玻璃态面板效果

## 构建验证
- ✅ TypeScript 类型检查 (vue-tsc --noEmit) — 通过
- ⬕ Vite 生产构建 (vite build) — 通过
- ✅ Wails 完整构建 (wails build -ldflags="-s -w") — 通过
- ✅ 构建时间: ~12s (含前后端编译+打包)
- ✅ macOS 制品: 14MB (目标 < 15MB)

## 项目结构
```
onlyRedis/
├── main.go              # 应用入口，Wails 配置
├── app.go               # 前后端桥接，绑定 40+ 方法
├── services/
│   ├── redis_service.go # Redis 核心操作
│   ├── config_store.go  # 加密配置存储
│   └── monitor_service.go # 监控采集
├── frontend/
│   ├── src/
│   │   ├── components/  # 11 个 Vue 组件
│   │   ├── stores/      # Pinia 状态管理
│   │   ├── views/       # 主布局
│   │   └── router/      # Vue Router
│   └── dist/            # 构建产物（内嵌到 Go 二进制）
└── build/bin/onlyRedis.app  # macOS 应用包
```

## 运行方式
```bash
# 开发模式 (热重载)
wails dev

# 生产构建
wails build -ldflags="-s -w"

# 仅构建前端
cd frontend && npm run build
```
