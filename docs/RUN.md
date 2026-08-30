# 运行说明

## 前置
- Go 1.19+（用了 cgo sqlite3，需 gcc）
- Node 18+/pnpm

## 一键启动 / 停止（推荐）
```bash
./start.sh   # 构建后端（有改动时）→ 后台启动后端(:7800) + 前端(:5173)，就绪后打印访问地址
./stop.sh    # 一键停止前后端（清理各自进程组）
```
- 日志与 PID：`logs/backend.log`、`logs/frontend.log`、`logs/*.pid`（已 gitignore）
- 已启动时再跑 `./start.sh` 会提示先 `./stop.sh`
- 首次运行自动 `pnpm install`；后端 go 源码有更新时自动重新构建
- 端口被其他进程占用时 `./stop.sh` 会提示

## 启动后端
```bash
cd backend
go run . -addr :7800
# 默认数据目录 ./data，数据库 ./data/tomato.db
# 可用环境变量 TOMATO_DATA_DIR 改目录
# 参数: -addr 监听地址, -db sqlite路径
```

## 启动前端（开发）
```bash
cd frontend
pnpm install   # 首次
pnpm dev       # http://localhost:5173, 自动代理 /api 到后端
```

## 构建
```bash
# 后端
cd backend && go build -o tomato-server .

# 前端
cd frontend && pnpm build    # 产物 dist/
```

## 端到端 smoke
见 docs/PROGRESS.md 中的验证记录。

## API 概览
| 方法 | 路径 | 说明 |
|------|------|------|
| GET/POST | /api/projects | 列/建项目 |
| GET/PUT/DELETE | /api/projects/{id} | 项目 CRUD |
| PATCH | /api/projects/{id}/status | 完成/止损归档 |
| GET | /api/projects/{id}/stats | 投入统计 |
| POST | /api/projects/{id}/milestones | 加里程碑 |
| PATCH/DELETE | /api/milestones/{mid} | 改/删里程碑 |
| POST | /api/projects/{id}/sessions | 开始会话 |
| GET | /api/sessions/{id} | 会话详情(含事件流) |
| POST | /api/sessions/{id}/pause | 暂停(可带备注) |
| POST | /api/sessions/{id}/resume | 恢复 |
| POST | /api/sessions/{id}/comment | 加文本备注 |
| POST | /api/sessions/{id}/voice | 上传语音(multipart) |
| POST | /api/sessions/{id}/end | 结束 |
| GET/PUT | /api/settings | 番茄时长配置 |
| GET | /api/voice/{filename} | 语音文件 |
