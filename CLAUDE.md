# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 1. 项目概述

### 1.1 项目简介
Image Viewer 是一个基于 Web 的高性能图片查看与筛片管理应用。系统采用**单文件分发架构**（后端 Go 编译为单个二进制文件并硬编码嵌入前端静态资源），既支持单机双击即用，也支持局域网内跨设备网页操作。

### 1.2 核心功能要求
*   **双轨匹配引擎:** 自动遍历指定目录，将同名 RAW 文件（如 `.CR3`, `.ARW`, `.NEF`）与 JPG 文件聚合为单一“逻辑资产”（Asset）。对无法相互匹配的“孤儿文件”进行自动标注。
*   **高效元数据筛选:** 支持基于 EXIF 信息（拍摄时间、相机型号、焦段、光圈、ISO 等）的极速条件过滤。
*   **分级评分系统:** 提供 0-5 星级评分及颜色标签（Color Label）机制，支持快捷键高频筛片。
*   **AI 扩展接口:** 预留本地 AI 模块切入点（如画质盲评、闭眼检测、语义检索），数据层提供 AI 推理状态标记。

### 1.3 非功能要求
*   **性能 (核心指标):** 
    *   万级图库大瀑布流滚动流畅，图片加载响应时间小于 2s。
    *   支持**双层缓存管道**（200px 视图缩略图 + 2048px 全屏高清图），RAW 格式优先提取内嵌预览图，严禁高频解码几十兆的原始物理文件。
*   **兼容性:** 完美支持现代现代浏览器 (Chrome, Firefox, Safari, Edge)。
*   **可扩展性:** 前后端代码高度解耦，数据流单向传递，留出完整的 AI 插件接口。

## Tech Stack

- **Backend**: Go + Gin framework, SQLite for metadata
- **Frontend**: Vue 3 Composition API (Script Setup), Pinia state management
- **Shared types**: Go structs in `shared/types/` — these are the source of truth; frontend mirrors them as TypeScript types

##  开发与目录规范

*   **前端规范:** 遵循 Vue 3 Composition API 风格，使用 Script Setup 语法，UI 组件要求具备高性能虚拟滚动（Virtual Scroll）能力。
*   **后端规范:** 遵循 Go 标准项目布局（Standard Go Project Layout），引入有界并发工作池（Worker Pool）处理 I/O。
*   **类型共享规范:** 所有前后端通用的核心结构体与数据类型定义，统一维护在项目根目录的 `shared/types/` 目录下（Go 端作为独立包引入，前端可通过工具或手动对齐为 TS 类型），确保前后端契约绝对一致。

### 项目目录结构

```text
image-viewer/
├── shared/
│   └── types/              # 💡 共享类型定义目录
│       └── asset.go        # 核心资产、物理文件、EXIF等结构体定义
├── cmd/
│   └── viewer/
│       └── main.go         # 后端唯一入口：初始化SQLite、初始化服务、启动HTTP监听
├── internal/               # 后端私有业务逻辑
│   ├── config/
│   │   └── config.go       # 全局配置（缓存路径、支持后缀、并发度控制）
│   ├── repository/         # 基础设施层：数据持久化（SQLite CRUD，屏蔽SQL细节）
│   │   ├── db.go           # SQLite 连接初始化与表结构自动迁移
│   │   └── asset_repo.go   # Assets 仓储接口与实现
│   ├── service/            # 核心业务层：纯粹业务逻辑
│   │   ├── scanner.go      # 多线程并发扫描、双轨匹配算法
│   │   ├── asset.go        # 评分、标签切换、多条件过滤
│   │   └── thumb.go        # 缩略图生成与本地 .cache/ 二级缓存管理
│   └── transport/          # 传输层
│       └── http/
│           ├── router.go   # 路由注册
│           └── handler.go  # HTTP 处理器（面向前端的 JSON 接口）
├── web/                    # 前端工程（独立开发，编译后通过 go:embed 嵌入）
│   ├── dist/               # 前端静态编译产物
│   └── src/                # Vue 3 Composition API 源码
├── go.mod
└── go.sum

## Key Architectural Decisions

1. **Single-file distribution**: Go binary embeds frontend via `go:embed`. No separate deployment, no Node runtime needed at runtime.

2. **Dual-track matching engine**: RAW files (.CR3, .ARW, .NEF) and JPG files with the same base name are aggregated into a single "Asset" logical unit. Orphan files (no match) are flagged. Composite key for matching: `DirPath + "_" + Lowercase(AssetName)`.

3. **Dual-layer cache pipeline** (critical for performance):
   - Layer 1: 200px WebP thumbnails for grid/waterfall view
   - Layer 2: 2048px WebP for full-screen preview
   - RAW files: **always extract embedded preview JPEG** — never decode the full RAW file for thumbnails
   - Cache stored in `storage/cache/`, served from filesystem when available

4. **Scan workflow**: Frontend submits directory path → Backend returns `202 Accepted` immediately → async goroutine walk + concurrent worker pool via bounded channel → batch SQLite insert via transaction → progress reported via WebSocket or polling.

5. **Unidirectional data flow**: Frontend → HTTP → Handlers → Services → Repositories → SQLite. No circular dependencies.

6. **AI extension interface**: Data layer reserves `ai_status` fields on assets for future AI integration (quality assessment, closed-eye detection, semantic search).

## Core Features

- 0-5 star rating + color labels with keyboard shortcuts
- EXIF-based filtering (capture time, camera model, focal length, aperture, ISO)
- Virtual scrolling for 10k+ image libraries
- RAW format support: ARW (Sony), with CR2, CR3, NEF planned

## Code Conventions
- 使用 snake_case 命名变量
- 所有 API 需要写单元测试
- PR 合并前必须通过 CI/UT