#!/usr/bin/env bash
# 一键停止番茄投资人（清理 backend + frontend 及各自进程组）
set -u

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="$ROOT/logs"
PID_BE="$LOG_DIR/backend.pid"
PID_FE="$LOG_DIR/frontend.pid"

STOPPED=0
OLD_BE=""
OLD_FE=""

stop_pg() { # $1=pid文件 $2=名称
  [ ! -f "$1" ] && return
  local pid pgid
  pid="$(cat "$1" 2>/dev/null)"
  [ -z "${pid:-}" ] && return
  # 校验 pid 归属：只杀由 start.sh 拉起的进程组
  pgid="$(ps -o pgid= -p "$pid" 2>/dev/null | tr -d ' ')"
  [ -z "${pgid:-}" ] && return
  # 用命令行再确认，避免误杀同名进程
  local marker
  marker="$(ps -o args= -p "$pid" 2>/dev/null)"
  case "$marker" in
    *tomato-server*|*pnpm*dev*|*vite*) ;;
    *) return ;;
  esac
  echo "停止 $2 (pid=$pid, pgid=$pgid)"
  kill -TERM -- "-$pgid" 2>/dev/null || kill -TERM "$pid" 2>/dev/null
  STOPPED=1
}

stop_pg "$PID_BE" "后端 backend"
stop_pg "$PID_FE" "前端 frontend"

# 先保存 pid 再删文件，供退出等待使用
OLD_BE="$(cat "$PID_BE" 2>/dev/null || true)"
OLD_FE="$(cat "$PID_FE" 2>/dev/null || true)"
rm -f "$PID_BE" "$PID_FE"

if [ "$STOPPED" = 1 ]; then
  # 等待进程组真正退出，避免端口未释放
  for _ in $(seq 1 10); do
    if { [ -z "$OLD_BE" ] || ! kill -0 "$OLD_BE" 2>/dev/null; } && \
       { [ -z "$OLD_FE" ] || ! kill -0 "$OLD_FE" 2>/dev/null; }; then
      break
    fi
    sleep 1
  done
  echo "✅ 已停止"
fi

# 兜底提醒：端口仍被占用（可能是有别的进程占用）
if curl -sf -o /dev/null "http://127.0.0.1:7800/api/projects" 2>/dev/null; then
  echo "⚠️  7800 端口仍有服务响应，可能被其他进程占用: $(ss -tlnp 2>/dev/null | grep ':7800' || true)"
fi
if curl -sf -o /dev/null "http://127.0.0.1:5173/" 2>/dev/null; then
  echo "⚠️  5173 端口仍有服务响应，可能被其他进程占用: $(ss -tlnp 2>/dev/null | grep ':5173' || true)"
fi
