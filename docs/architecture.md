# 架构设计规范 (v1.2)

## 1. 系统架构

```
┌─────────────────┐             ┌─────────────────┐
│                 │             │                 │
│   Frontend      │────────────▶│   Backend       │
│   (Vue 3)       │   HTTP      │   (Go + Gin)    │
│                 │             │                 │
└─────────────────┘             └────────┬────────┘
                                         │
                         ┌───────────────┴───────────────┐
                         ▼                               ▼
              ┌─────────────────┐             ┌─────────────────┐
              │   File System   │             │  Local Database │
              │ (直接读取本地RAW/JPG)│             │    (SQLite 3)   │
              └─────────────────┘             └─────────────────┘
                         │
                         ▼
              ┌─────────────────┐
              │  Local .cache/  │
              │ (二级WebP缩略图管道)│
              └─────────────────┘

```

---

## 2. 目录结构设计

### 2.1 Shared (类型契约)

```text
shared/
└── types/             # 核心实体（Go语言编写，前端手动或通过工具转为 TS Type）
    └── asset.go       # 包含 Asset, MediaFile, ExifMeta 等核心结构

```

### 2.2 Backend (Go + Gin)

遵循 Go 语言标准布局，解耦传输层（Handlers）、业务层（Services）与存储层（Repositories）。

```text
backend/
├── cmd/
│   └── server/
│       └── main.go         # 主程序入口（拉起 SQLite、注册路由、绑定端口）
├── internal/
│   ├── config/
│   │   └── config.go       # 配置管理（扫描路径、缓存路径限制、支持的后缀名）
│   ├── handlers/           # HTTP 处理器（负责解析 Gin 上下文，返回标准 JSON）
│   ├── services/           # 核心业务逻辑（Scanner并发扫描引擎、Thumb缩略图解码器）
│   ├── repositories/       # 💡 仓储层（负责 SQLite 读写，隔离原生 SQL 语句）
│   └── middleware/         # 跨域控制 (CORS)、静态缓存控制
└── storage/                # 💡 本地运行时目录（不包含用户照片）
    ├── cache/              # 存放系统生成的 200px 和 2048px 的高频 WebP 缩略图
    └── viewer.db           # 系统的 SQLite 单文件数据库

```

### 2.3 Frontend (Vue 3 + Composition API)

```text
frontend/
├── src/
│   ├── api/                # Axios / Fetch 封装的请求
│   ├── components/         # 核心组件（瀑布流组件、虚拟滚动容器、大图预览浮层）
│   ├── composables/        # 组合式函数（如 useKeyboardShortcut 快捷键打分绑定）
│   ├── stores/             # Pinia 状态管理（当前激活的图片、过滤条件、扫描进度）
│   ├── types/              # 映射 shared/types/ 的 TypeScript 类型声明文件
│   └── views/              # 页面（Gallery 瀑布流主页、Folder 文件夹选择页）

```

---

## 3. 核心数据流设计 (Data Flow Design)

系统的数据流向遵循“单向循环、缓存优先”的原则，以下是三个最高频场景的数据流转逻辑：

### 3.1 资产并发扫描与双轨匹配数据流

当用户在前端指定一个本地照片目录（如 `/Volumes/Photos/2026_Raw/`）并点击扫描：

1. **触发扫描:** Frontend -.
Backend">
Frontend 提交绝对路径。Backend 验证路径有效性后，立即返回 `202 Accepted`，开启异步 Goroutine 扫描，并建立 Web-Socket 或轮询通道供前端监听进度。


2. **并发消费与聚拢:** Backend Local.
主线程进行 `WalkDir`，把照片路径丢进有界 Channel。多个 Worker 线程并发并发读取物理文件，通过 `shared/types.Asset` 的复合键 `DirPath + "_" + Lowercase(AssetName)` 将同名 RAW 和 JPG 聚合进同一个 Asset 实例，并标记配对状态。


3. **批处理落库:** Backend -.
SQLite">
扫描结束后，将聚合好的 `[]Asset` 和 `[]MediaFile` 通过事务（Transaction）批量写入 SQLite 数据库，并对时间（CapturedAt）和评分建立索引。


### 3.2 瀑布流高效加载数据流 (双层缓存机制)

前端首次进入画廊主页，加载数万张照片的视图：

```
[前端: 触发虚拟滚动] 
         │
         ▼ (请求当前可视区域内的几百张资产)
[GET /api/v1/assets?page=1&limit=50] ──► [后端: 极速查询 SQLite 索引]
         │
         ▼ (返回包含 Asset 核心元数据及两层缓存路径的 JSON)
[前端: 渲染图片渲染列表] ──► `<img src="/api/v1/thumbs/:id?size=grid" />`
         │
         ▼ 
[后端: 命中缓存检查] 
    ├── 1. 如果 storage/cache/ 存在对应的 200px WebP -> 毫秒级直接返回
    └── 2. 如果不存在 -> 丢入后台线程池异步提取 RAW 内嵌预览图，生成 WebP 写入缓存并返回

```

### 3.3 快捷键分级评分数据流

用户在前端全屏查看照片，按下数字键 `5`（打 5 星）：

```
[前端: 监听键盘事件] ──► 触发 `useKeyboardShortcut` ──► 内存变更当前图片状态
         │
         ▼ (发送异步轻量请求)
[POST /api/v1/assets/:id/rate  { "rating": 5 }]
         │
         ▼ 
[后端 Handlers] ──► 调用 [Services] ──► 调用 [Repositories]
         │
         ▼ (执行高效局部更新 SQL)
[UPDATE assets SET rating = 5 WHERE id = :id;] ──► 写入 SQLite (毫秒级成功)
         │
         ▼ 
[后端返回 200 OK] ──► 前端 UI 无缝保持高亮，不触发整页刷新

```
