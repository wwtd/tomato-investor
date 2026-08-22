# 进展记录

> 续作入口。上下文重开时先读本文件 + DESIGN.md，然后从"下一步"继续。

## 决策日志
- 2026-08-21: 锁定 MVP 范围。Go+React，单用户预留 owner_id，番茄=可配置时长(默认8h/480min)，会话=一次会话多次暂停(START→[PAUSE+comment→RESUME]*→END)，语音备注仅存音频+手填文本，部署=本地自托管 SQLite。
- 2026-08-21: Android 打包方案 = 本地只做前端适配 + GitHub CI 出 Debug APK（无需本机 Android SDK/JDK）。API 地址改为运行时可配置（localStorage server_url）。

## 已完成
- [x] 设计文档 docs/DESIGN.md
- [x] 后端骨架 + SQLite schema + db 层
- [x] 后端 API: 项目CRUD/状态/统计、里程碑CRUD、番茄会话生命周期(pause/resume/comment/voice/end)、设置、单用户seed
- [x] 前端: 项目列表+创建、项目详情(预算/必要性/里程碑/历史会话)、番茄会话运行器(暂停/备注/语音录制)、设置
- [x] 端到端 API smoke 验证通过
- [x] 运行文档 docs/RUN.md
- [x] Android: Capacitor 集成 (core/android/cli 8.5.0)，capacitor.config.ts (appId com.wwtd.tomatoinvestor)
- [x] Android 适配: cleartext HTTP + RECORD_AUDIO 权限，删除模板占位测试
- [x] API 服务地址可配置: api.ts getBase()/setServerUrl() 基于 localStorage; UI 加「🔗服务地址」+测试连接
- [x] GitHub CI: .github/workflows/android-build.yml，自动 assembleDebug + 上传 APK artifact

## E2E 验证记录 (2026-08-21)
全流程通过：
1. 创建项目(含必要性/最小集/止损说明/预算) ✅
2. 添加里程碑 ✅
3. 里程碑切换完成 PATCH /milestones/{mid} ✅ (需独立挂载路由)
4. 开始番茄会话 ✅
5. 暂停(带备注) ✅
6. 恢复 ✅
7. 加文本备注 ✅
8. 事件流查看(pause/resume/comment) ✅
9. 结束会话(消耗番茄) ✅ consumed_tomato=1
10. 统计: consumed_tomatoes=1, remaining=2 ✅
11. 止损归档 ✅
12. 语音上传(运行中会话) ✅ file+event_id 返回, 事件流含 voice_note

## 关键实现点
- 番茄会话时长计算: 基于事件流重建 running 段总和，暂停不计时
- 路由: /api/projects/{id}/sessions 用于 list/start; /api/sessions/{sid}/* 用于单会话操作（独立挂载）
- 同理 /api/milestones/{mid} 独立挂载
- 语音: multipart 上传，存 data/voice/，不接 STT

## 下一步（非 MVP，按需迭代）
- [x] 移动端 APP（Capacitor + CI 出 debug APK，首次构建需 push 后查看 Actions 结果）
- [ ] 前端浏览器实测 UI（需 pnpm dev + 浏览器交互）
- [ ] STT 语音转写
- [ ] 任务链并行视图
- [ ] 多端同步优化

## 环境备忘
- Go 1.19.8 (注意：较老，避免用泛型之外的 1.20+ 特性；chi v5 兼容)
- Node 22, pnpm 可用
- gcc 12.2 可用 → 可用 cgo sqlite3 (mattn/go-sqlite3)
- 工作目录: /home/tengfei/tomato-dir
- 注意: 机器有 http_proxy 环境变量，本地 curl 测试需 `export no_proxy='*'`
