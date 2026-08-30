#!/usr/bin/env bash
# 一键启动番茄投资人（后端 :7800 + 前端 :5173，后台运行，日志在 logs/）
set -u

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_BIN="$ROOT/backend/tomato-server"
LOG_DIR="$ROOT/logs"
BE_LOG="$LOG_DIR/backend.log"
FE_LOG="$LOG_DIR/frontend.log"
PID_BE="$LOG_DIR/backend.pid"
PID_FE="$LOG_DIR/frontend.pid"

# 检查是否已在运行
running() { [ -f "$1" ] && kill -0 "$(cat "$1")" 2>/dev/null; }
if running "$PID_BE" || running "$PID_FE"; then
  echo "已在运行（backend.pid=$(cat "$PID_BE" 2>/dev/null) / frontend.pid=$(cat "$PID_FE" 2>/dev/null)）。先 ./stop.sh 再启动。"
  exit 1
fi

mkdir -p "$LOG_DIR"

# 后端：无产物或源码更新时自动构建（TOMATO_DATA_DIR 默认 backend/data）
if [ ! -x "$BACKEND_BIN" ] || [ "$(find "$ROOT/backend" -name '*.go' -newer "$BACKEND_BIN" | wc -l)" -gt 0 ]; then
  echo "构建后端 ..."
  (cd "$ROOT/backend" && go build -o "$BACKEND_BIN" .) || { echo "后端构建失败"; exit 1; }
fi

# 前端：首次运行安装依赖
if [ ! -d "$ROOT/frontend/node_modules" ]; then
  echo "安装前端依赖（首次）..."
  (cd "$ROOT/frontend" && pnpm install) || { echo "pnpm install 失败"; exit 1; }
fi

# 启动后端（setsid 独立会话；内层 bash 写入真实 pid 再 exec，避免 setsid fork 造成 pid 偏移）
(cd "$ROOT/backend" && setsid bash -c 'echo $$ >"$0"; exec "$1"' "$PID_BE" "$BACKEND_BIN" >"$BE_LOG" 2>&1 </dev/null &)
# 启动前端（开发服务器）
(cd "$ROOT/frontend" && setsid bash -c 'echo $$ >"$0"; exec pnpm dev' "$PID_FE" >"$FE_LOG" 2>&1 </dev/null &)

# 等待就绪
echo "等待服务启动 ..."
OK_BE=0; OK_FE=0
for _ in $(seq 1 30); do
  [ "$OK_BE" = 0 ] && curl -sf -o /dev/null "http://127.0.0.1:7800/api/projects" && OK_BE=1
  [ "$OK_FE" = 0 ] && curl -sf -o /dev/null "http://127.0.0.1:5173/" && OK_FE=1
  [ "$OK_BE" = 1 ] && [ "$OK_FE" = 1 ] && break
  sleep 1
done

# 校验端口归属：监听进程可能是记录 pid 的进程组内子进程（pnpm → vite），
# 只要 owner 属于我们启动的进程组即视为正常；否则说明端口被别的进程占用
owns_port() { # $1=记录pid文件 $2=owner pid；owner 与记录 pid 同进程组即视为正常
  local leader lpgid opgid
  leader="$(cat "$1" 2>/dev/null || true)"
  [ -z "$leader" ] && return 1
  [ "$2" = "$leader" ] && return 0
  lpgid="$(ps -o pgid= -p "$leader" 2>/dev/null | tr -d ' ')"
  opgid="$(ps -o pgid= -p "$2" 2>/dev/null | tr -d ' ')"
  [ -n "$lpgid" ] && [ "$lpgid" = "$opgid" ] && return 0
  return 1
}

if [ "$OK_FE" = 1 ] && [ -f "$PID_FE" ]; then
  OWNER="$(ss -tlnp 2>/dev/null | grep -E ':(5173) ' | sed -n 's/.*pid=\([0-9]*\).*/\1/p' | head -1)"
  if [ -n "$OWNER" ] && ! owns_port "$PID_FE" "$OWNER"; then
    echo "⚠️  5173 端口被其他进程占用（pid=$OWNER），本次启动的前端实际在 5174，请先停止占用进程"
    exit 1
  fi
fi
if [ "$OK_BE" = 1 ] && [ -f "$PID_BE" ]; then
  OWNER="$(ss -tlnp 2>/dev/null | grep -E ':(7800) ' | sed -n 's/.*pid=\([0-9]*\).*/\1/p' | head -1)"
  if [ -n "$OWNER" ] && ! owns_port "$PID_BE" "$OWNER"; then
    echo "⚠️  7800 端口被其他进程占用（pid=$OWNER），后端未正常监听，请先停止占用进程"
    exit 1
  fi
fi

if [ "$OK_BE" = 1 ] && [ "$OK_FE" = 1 ]; then
  LAN_IP="$(hostname -I 2>/dev/null | awk '{print $1}')"
  echo
  echo "✅ 番茄投资人已启动"
  echo "  本机:  http://localhost:5173"
  [ -n "${LAN_IP:-}" ] && echo "  局域网: http://$LAN_IP:5173"
  echo "  日志:  $LOG_DIR （停止: ./stop.sh）"
else
  echo "启动超时，请检查日志："
  [ "$OK_BE" = 0 ] && echo "  tail $BE_LOG"
  [ "$OK_FE" = 0 ] && echo "  tail $FE_LOG"
  exit 1
fi
