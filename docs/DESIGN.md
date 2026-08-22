# 番茄投资人 — 设计规格

> 状态：MVP 设计定稿。本文件是项目持久记录，上下文重开后从这里延续。

## 1. 灵感来源

面向业余碎片化时间管理。核心隐喻：把时间当投资资本，用项目管理 + 投资思维（必要性评估、最小集、里程碑、止损、归档）管理项目。

- 番茄 = 抽象投资单位，默认 1 番茄 = 8 小时，**可配置**
- 不是传统 25 分钟番茄钟，而是一个可分段的长会话单位

## 2. MVP 锁定决策（已与用户确认）

| 维度 | 决策 |
|------|------|
| 用户模型 | 单用户，数据模型预留 `owner_id` 字段，未来可扩多用户 |
| 技术栈 | Go (标准库 + chi 路由 + sqlite3) ；React + Vite + TypeScript |
| 番茄定义 | 番茄 = 可配置时长（默认 480 分钟 = 8h），会话有完整生命周期 |
| 会话模型 | 一次会话多次暂停：START → [PAUSE+comment → RESUME]* → END，END 时消耗 1 番茄预算 |
| 语音备注 | 仅存音频文件 + 手填文本，不接 STT（后续可加） |
| 部署 | 本地自托管，SQLite 文件存储 |
| 数据库 | SQLite，cgo 驱动 mattn/go-sqlite3 |

## 3. 核心领域模型

```
User
  id, owner_id (预览), tomato_minutes (番茄配置时长), created_at

Project
  id, owner_id
  title, necessity (必要性说明), mvp_scope (最小集), stoploss_note (止损说明)
  budget_tomatoes (投资番茄数)
  tomato_minutes (本项目番茄时长，null=用全局默认)
  status: active | completed | archived_stoploss (止损归档)
  created_at, archived_at

Milestone
  id, project_id, title, done bool, order_idx, created_at

Session (番茄会话)
  id, project_id
  status: running | paused | ended
  started_at, ended_at
  consumed_tomato (结束时为0或1; 1 表示正式消耗一番茄)
  note (结束总结)

SessionEvent (会话内事件流)
  id, session_id
  type: pause | resume | comment | voice_note
  payload (JSON: comment 文本 / voice_note 引用)
  voice_file (文件名，type=voice_note 时)
  at (时间戳)

VoiceNote
  id, session_event_id, file_path, mime, duration_ms, text (手填), created_at
```

会话实际运行时长 = 所有 running 段总和（从事件流计算），非墙钟时间。暂停不计时。

## 4. API 设计 (REST/JSON)

前缀 `/api`

### 项目
- `GET    /projects`                列出项目（可按 status 过滤）
- `POST   /projects`                创建（含 necessity/mvp_scope/budget_tomatoes/tomato_minutes）
- `GET    /projects/{id}`           详情（含里程碑、会话数、已投入番茄、剩余）
- `PUT    /projects/{id}`           编辑
- `PATCH  /projects/{id}/status`    状态变更（complete / archive_stoploss）
- `DELETE /projects/{id}`           删除（仅无会话时）

### 里程碑
- `POST   /projects/{id}/milestones`
- `PATCH  /milestones/{id}`         (toggle done / 改 title / 排序)
- `DELETE /milestones/{id}`

### 番茄会话
- `POST   /projects/{id}/sessions`           开始会话 → 返回 session
- `GET    /sessions/{id}`                    详情（含事件流）
- `POST   /sessions/{id}/pause`              暂停（body: comment 可选）
- `POST   /sessions/{id}/resume`             恢复
- `POST   /sessions/{id}/comment`            运行中加文本备注
- `POST   /sessions/{id}/voice`              上传语音备注 (multipart: file + text)
- `POST   /sessions/{id}/end`                结束 (body: note, consume_tomato bool)
- `GET    /projects/{id}/sessions`           项目下所有会话

### 配置
- `GET    /settings`            tomato_minutes 及其他
- `PUT    /settings`
- `GET    /voice/{filename}`    静态音频文件

### 统计（MVP 简版）
- `GET    /projects/{id}/stats`  已投入番茄数、已投入分钟、剩余番茄、剩余分钟

## 5. 目录结构

```
tomato-dir/
  backend/
    go.mod
    main.go
    db/
      schema.sql
      db.go
    model/
      types.go
    api/
      projects.go
      milestones.go
      sessions.go
      voice.go
      settings.go
    docs/
      DESIGN.md  (本文件副本引用)
  frontend/
    (Vite React TS)
  docs/
    DESIGN.md
    RUN.md
    PROGRESS.md   (续作记录)
```

## 6. 后续（非 MVP）
- STT 语音转写
- 多端同步（当前已单后端，前端 web 即可多端浏览器访问）
- 任务链并行视图
- APP（移动端）
- 多用户

## 7. 续作约定
- 任何阶段性进展记录到 `docs/PROGRESS.md`
- 架构变更同步更新本文件
- 上下文重开时先读 `docs/PROGRESS.md` 和本文件
