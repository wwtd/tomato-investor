# 番茄投资人 (Tomato Investor)

> 用投资思维管理业余碎片化时间。把**时间当资本**，一个番茄 = 一个可配置时长的投资单位（默认 8 小时），用项目管理 + 投资逻辑（必要性、最小集、里程碑、止损、归档）驱动个人项目。

## 核心概念

- **番茄**：抽象投资单位，默认 480 分钟（8h），**可全局或按项目配置**。不是传统 25 分钟番茄钟，而是可分段的长会话单位。
- **会话生命周期**：一次会话可多次暂停，每次暂停可记录情景备注（文本/语音）。`START → [PAUSE+comment → RESUME]* → END`，结束时决定是否消耗 1 番茄预算。
- **投资预算**：每个项目设定"投资几个番茄"，追踪已投入/剩余。
- **止损归档**：预算用光但产出不及预期时，可归档停止，避免沉没成本。

## 功能（MVP）

- 项目 CRUD + 必要性/最小集/止损说明 + 投资预算（番茄数）
- 番茄会话运行器：开始/暂停(带备注)/恢复/文本备注/结束
- 语音备注（录音上传，当前仅存音频 + 手填文字，未接 STT）
- 里程碑管理：添加/完成/删除
- 投入统计：已投入/剩余番茄数、分钟数
- 番茄时长配置（默认 480 分钟）

## 技术栈

| 层 | 技术 |
|----|------|
| 后端 | Go 1.19 + chi v5 + SQLite (mattn/go-sqlite3, cgo) |
| 前端 | React 19 + Vite + TypeScript |
| 数据 | SQLite 单文件，`backend/data/tomato.db` |
| 部署 | 本地自托管，前后端分离 |

## 快速开始

前置：Go 1.19+、Node 18+、pnpm、gcc（cgo 需要）。

### 启动后端
```bash
cd backend
go run . -addr :7800
# 默认数据目录 ./data，数据库 ./data/tomato.db
# 环境变量 TOMATO_DATA_DIR 可改数据目录
```

### 启动前端（开发）
```bash
cd frontend
pnpm install
pnpm dev
# http://localhost:5173，自动代理 /api 到后端 :7800
```

### 局域网访问
前端默认已绑 `0.0.0.0`，局域网内其他设备访问：
```
http://<本机局域网IP>:5173
```
> 注意：通过局域网 IP（非 localhost/HTTPS）访问时，浏览器安全上下文限制会导致**录音不可用**。详见文档。

## API 概览

前缀 `/api`，REST/JSON。

| 方法 | 路径 | 说明 |
|------|------|------|
| GET/POST | `/projects` | 列/建项目 |
| GET/PUT/DELETE | `/projects/{id}` | 项目 CRUD |
| PATCH | `/projects/{id}/status` | 完成 / 止损归档 |
| GET | `/projects/{id}/stats` | 投入统计 |
| POST | `/projects/{id}/milestones` | 添加里程碑 |
| PATCH/DELETE | `/milestones/{id}` | 改/删里程碑 |
| POST | `/projects/{id}/sessions` | 开始会话 |
| GET | `/sessions/{id}` | 会话详情（含事件流） |
| POST | `/sessions/{id}/pause` | 暂停（可带备注） |
| POST | `/sessions/{id}/resume` | 恢复 |
| POST | `/sessions/{id}/comment` | 文本备注 |
| POST | `/sessions/{id}/voice` | 语音上传（multipart） |
| POST | `/sessions/{id}/end` | 结束 |
| GET/PUT | `/settings` | 番茄时长配置 |
| GET | `/voice/{filename}` | 语音文件 |

## 目录结构

```
tomato-dir/
  backend/     Go 后端（chi + sqlite）
    db/        schema + db 层
    model/     类型
    api/       REST handlers
  frontend/    React + Vite + TS
    src/
      api.ts              API 客户端
      components/         列表/详情/会话/里程碑组件
  docs/
    DESIGN.md   设计规格 + 数据模型
    PROGRESS.md 开发进展记录
    RUN.md      详细运行说明
```

## 数据模型要点

- `projects`：标题、必要性、最小集、止损说明、预算番茄数、番茄时长覆盖、状态(active/completed/archived_stoploss)
- `sessions`：一次番茄会话，状态(running/paused/ended)，`consumed_tomato` 标记是否消耗预算
- `session_events`：会话内事件流(pause/resume/comment/voice_note)，用于重建运行时长（暂停不计时）
- `voice_notes`：语音备注文件 + 手填文字
- 单用户，所有表预留 `owner_id` 字段，未来可扩多用户

## 后续规划（非 MVP）

- STT 语音转写
- 任务链并行视图
- 移动端 APP
- 多端同步
- 多用户

## License

内部项目，未指定 License。
