#!/usr/bin/env bash
#
# 轻燕工作台 - 服务管理脚本
# 用于检测、启动、停止、重启前后端服务
#
# 用法:
#   ./manage.sh             交互模式
#   ./manage.sh status      仅查看状态
#   ./manage.sh start all   启动全部服务
#   ./manage.sh stop all    停止全部服务

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
FRONTEND_DIR="$SCRIPT_DIR/Web-Front"
BACKEND_DIR="$SCRIPT_DIR/Server-code"
DIST_FRONTEND_DIR="$SCRIPT_DIR/dist/frontend"
DIST_BACKEND_DIR="$SCRIPT_DIR/dist/backend"
BACKEND_PORT=8090
FRONTEND_PORT=3000

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

icon_ok="${GREEN}✓${NC}"
icon_no="${RED}✗${NC}"
icon_arrow="${BLUE}→${NC}"

# ============================================================
# 工具函数
# ============================================================

log_info()  { echo -e "${CYAN}[INFO]${NC} $*"; }
log_ok()    { echo -e "${GREEN}[OK]${NC} $*"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*"; }

# -------------------------------------------------------
# 检测指定端口是否有进程监听
# -------------------------------------------------------
port_pid() {
  local port=$1
  local pid
  if command -v lsof &>/dev/null; then
    pid=$(lsof -ti "tcp:$port" -sTCP:LISTEN 2>/dev/null || true)
  elif command -v ss &>/dev/null; then
    pid=$(ss -tlnp "sport = :$port" 2>/dev/null | awk 'NR>1 {print $NF}' | grep -oP '(?<=pid=)\d+' | head -1 || true)
  elif command -v netstat &>/dev/null; then
    pid=$(netstat -tlnp 2>/dev/null | awk -v p=":$port " '$4~p {print $NF}' | sed 's|/.*||' | head -1 || true)
  fi
  echo "${pid:-}"
}

# -------------------------------------------------------
# 轮询等待端口就绪
# -------------------------------------------------------
wait_for_port() {
  local port=$1 label=$2 max_wait=${3:-30}
  local waited=0
  while [[ $waited -lt $max_wait ]]; do
    local pid
    pid=$(port_pid "$port")
    if [[ -n "$pid" ]]; then
      log_ok "${label}服务就绪 (PID $pid → 端口 $port, 耗时 ${waited}s)"
      return 0
    fi
    sleep 1
    waited=$((waited + 1))
    if [[ $((waited % 5)) -eq 0 ]]; then
      log_info "等待${label}服务启动... (${waited}s/${max_wait}s)"
    fi
  done
  log_error "${label}服务启动超时 (${max_wait}s)，请手动检查"
  return 1
}

# -------------------------------------------------------
# 获取进程命令行信息
# -------------------------------------------------------
proc_info() {
  local pid=$1
  if [[ -z "$pid" ]]; then
    echo "未知"
    return
  fi
  if ps -p "$pid" &>/dev/null; then
    local cmd
    cmd=$(ps -p "$pid" -o command= 2>/dev/null | head -1 || echo "未知")
    echo "$cmd" | cut -c1-100
  else
    echo "未知"
  fi
}

# ============================================================
# 状态检测
# ============================================================

check_backend() {
  local pid
  pid=$(port_pid "$BACKEND_PORT")
  if [[ -n "$pid" ]]; then
    echo "running" "$pid"
  else
    echo "stopped"
  fi
}

check_frontend() {
  local pid
  pid=$(port_pid "$FRONTEND_PORT")
  if [[ -n "$pid" ]]; then
    echo "running" "$pid"
  else
    echo "stopped"
  fi
}

# -------------------------------------------------------
# 打印服务状态表
# -------------------------------------------------------
print_status() {
  echo ""
  echo -e "  ${CYAN}┌─────────────────────────────────────────────────────────────┐${NC}"
  echo -e "  ${CYAN}│${NC}             轻燕工作台 · 服务运行状态                         ${CYAN}│${NC}"
  echo -e "  ${CYAN}├──────────┬───────┬───────┬──────────────────────────────────┤${NC}"
  echo -e "  ${CYAN}│${NC} 服务     ${CYAN}│${NC} 端口  ${CYAN}│${NC} 状态  ${CYAN}│${NC} 进程详情                           ${CYAN}│${NC}"
  echo -e "  ${CYAN}├──────────┼───────┼───────┼──────────────────────────────────┤${NC}"

  local be_result be_pid fe_result fe_pid
  read -r be_result be_pid <<< "$(check_backend)"
  read -r fe_result fe_pid <<< "$(check_frontend)"

  local be_status fe_status be_info fe_info
  if [[ "$be_result" == "running" ]]; then
    be_status="${GREEN}● 运行中${NC}"
    be_info="[PID $be_pid] $(proc_info "$be_pid")"
  else
    be_status="${RED}○ 未启动${NC}"
    be_info="${YELLOW}等待启动${NC}"
  fi

  if [[ "$fe_result" == "running" ]]; then
    fe_status="${GREEN}● 运行中${NC}"
    fe_info="[PID $fe_pid] $(proc_info "$fe_pid")"
  else
    fe_status="${RED}○ 未启动${NC}"
    fe_info="${YELLOW}等待启动${NC}"
  fi

  printf "  ${CYAN}│${NC} %-8s ${CYAN}│${NC} %-5s ${CYAN}│${NC} %b ${CYAN}│${NC} %-32s ${CYAN}│${NC}\n" \
    "后端" "$BACKEND_PORT" "$be_status" "$(echo "$be_info" | cut -c1-44)"
  printf "  ${CYAN}│${NC} %-8s ${CYAN}│${NC} %-5s ${CYAN}│${NC} %b ${CYAN}│${NC} %-32s ${CYAN}│${NC}\n" \
    "前端" "$FRONTEND_PORT" "$fe_status" "$(echo "$fe_info" | cut -c1-44)"
  echo -e "  ${CYAN}└──────────┴───────┴───────┴──────────────────────────────────┘${NC}"
  echo ""

  if [[ "$be_result" == "running" ]] && [[ "$fe_result" == "running" ]]; then
    echo -e "  ${GREEN}▸ 所有服务正常运行，访问 http://localhost:$FRONTEND_PORT${NC}"
    echo ""
  elif [[ "$be_result" == "stopped" ]] && [[ "$fe_result" == "stopped" ]]; then
    echo -e "  ${YELLOW}▸ 所有服务均未启动${NC}"
    echo ""
  else
    echo -e "  ${YELLOW}▸ 部分服务未启动${NC}"
    echo ""
  fi
}

# ============================================================
# 操作函数
# ============================================================

start_backend() {
  log_info "正在启动后端服务..."
  local pid
  pid=$(port_pid "$BACKEND_PORT")
  if [[ -n "$pid" ]]; then
    log_warn "后端服务已在运行 (PID $pid)，跳过启动"
    return 0
  fi

  cd "$BACKEND_DIR"
  if [[ -f "labelpro-server" ]]; then
    nohup ./labelpro-server &>/dev/null &
    local bg_pid=$!
    wait_for_port "$BACKEND_PORT" "后端" 30
  else
    nohup go run main.go &>/dev/null &
    local bg_pid=$!
    wait_for_port "$BACKEND_PORT" "后端" 60
  fi
}

start_frontend() {
  log_info "正在启动前端服务..."
  local pid
  pid=$(port_pid "$FRONTEND_PORT")
  if [[ -n "$pid" ]]; then
    log_warn "前端服务已在运行 (PID $pid)，跳过启动"
    return 0
  fi

  if [[ -f "$DIST_FRONTEND_DIR/index.html" ]]; then
    cd "$DIST_FRONTEND_DIR"
    nohup python3 -m http.server "$FRONTEND_PORT" --bind 0.0.0.0 &>/dev/null &
    wait_for_port "$FRONTEND_PORT" "前端(静态)" 5 || {
      log_warn "静态文件服务启动失败，回退到源码开发模式..."
      start_frontend_dev
    }
  else
    start_frontend_dev
  fi
}

start_frontend_dev() {
  cd "$FRONTEND_DIR"
  nohup npm run dev -- --port "$FRONTEND_PORT" &>/dev/null &
  wait_for_port "$FRONTEND_PORT" "前端(开发)" 30
}

stop_by_port() {
  local port=$1 label=$2
  local pid
  pid=$(port_pid "$port")
  if [[ -z "$pid" ]]; then
    log_warn "${label}服务未运行，无需停止"
    return 0
  fi

  log_info "正在停止${label}服务 (PID $pid)..."
  kill "$pid" 2>/dev/null || true

  local waited=0
  while [[ $waited -lt 10 ]]; do
    if ! kill -0 "$pid" 2>/dev/null; then
      log_ok "${label}服务已停止"
      return 0
    fi
    sleep 1
    waited=$((waited + 1))
  done

  log_warn "正常关闭超时，强制终止..."
  kill -9 "$pid" 2>/dev/null || true
  sleep 1
  log_ok "${label}服务已强制停止"
}

stop_backend()  { stop_by_port "$BACKEND_PORT" "后端"; }
stop_frontend() { stop_by_port "$FRONTEND_PORT" "前端"; }

restart_backend() {
  stop_backend
  sleep 1
  start_backend
}

restart_frontend() {
  stop_frontend
  sleep 1
  start_frontend
}

# ============================================================
# 批量操作
# ============================================================

start_all() {
  echo ""
  log_info "=== 启动全部服务 ==="
  start_backend
  start_frontend
  echo ""
  print_status
}

stop_all() {
  echo ""
  log_info "=== 停止全部服务 ==="
  stop_frontend
  stop_backend
  echo ""
  log_ok "全部服务已停止"
}

restart_all() {
  echo ""
  log_info "=== 重启全部服务 ==="
  stop_all
  sleep 2
  start_all
}

# ============================================================
# 交互菜单
# ============================================================

show_menu() {
  print_status

  local be_result fe_result
  read -r be_result _ <<< "$(check_backend)"
  read -r fe_result _ <<< "$(check_frontend)"

  local be_running=false fe_running=false
  [[ "$be_result" == "running" ]] && be_running=true
  [[ "$fe_result" == "running" ]] && fe_running=true

  local all_running=false all_stopped=false
  $be_running && $fe_running && all_running=true
  ! $be_running && ! $fe_running && all_stopped=true

  echo -e "  ${CYAN}┌────────────────────── 操作菜单 ──────────────────────┐${NC}"

  local opt=1

  # ---- 有服务未启动 → 提供启动选项 ----
  if $all_stopped; then
    echo -e "  ${CYAN}│${NC}  ${GREEN}[$opt]${NC} 启动全部服务                                     ${CYAN}│${NC}"
    opt=$((opt + 1))
  elif $all_running; then
    :
  else
    if ! $be_running; then
      echo -e "  ${CYAN}│${NC}  ${GREEN}[$opt]${NC} 启动后端服务                                     ${CYAN}│${NC}"
      opt=$((opt + 1))
    fi
    if ! $fe_running; then
      echo -e "  ${CYAN}│${NC}  ${GREEN}[$opt]${NC} 启动前端服务                                     ${CYAN}│${NC}"
      opt=$((opt + 1))
    fi
  fi

  # ---- 有服务在运行 → 提供停止/重启选项 ----
  if $all_running; then
    echo -e "  ${CYAN}│${NC}  ${YELLOW}[$opt]${NC} 停止全部服务                                     ${CYAN}│${NC}"
    opt=$((opt + 1))
    echo -e "  ${CYAN}│${NC}  ${YELLOW}[$opt]${NC} 重启全部服务                                     ${CYAN}│${NC}"
    opt=$((opt + 1))
    echo -e "  ${CYAN}│${NC}  ${YELLOW}[$opt]${NC} 仅重启后端服务                                   ${CYAN}│${NC}"
    opt=$((opt + 1))
    echo -e "  ${CYAN}│${NC}  ${YELLOW}[$opt]${NC} 仅重启前端服务                                   ${CYAN}│${NC}"
    opt=$((opt + 1))
  elif ! $all_stopped; then
    if $be_running; then
      echo -e "  ${CYAN}│${NC}  ${YELLOW}[$opt]${NC} 停止后端服务                                     ${CYAN}│${NC}"
      opt=$((opt + 1))
      echo -e "  ${CYAN}│${NC}  ${YELLOW}[$opt]${NC} 重启后端服务                                     ${CYAN}│${NC}"
      opt=$((opt + 1))
    fi
    if $fe_running; then
      echo -e "  ${CYAN}│${NC}  ${YELLOW}[$opt]${NC} 停止前端服务                                     ${CYAN}│${NC}"
      opt=$((opt + 1))
      echo -e "  ${CYAN}│${NC}  ${YELLOW}[$opt]${NC} 重启前端服务                                     ${CYAN}│${NC}"
      opt=$((opt + 1))
    fi
    echo -e "  ${CYAN}│${NC}  ${YELLOW}[$opt]${NC} 停止全部服务                                     ${CYAN}│${NC}"
    opt=$((opt + 1))
  fi

  local q_opt=$opt
  echo -e "  ${CYAN}│${NC}  ${RED}[q]${NC} 退出                                             ${CYAN}│${NC}"
  echo -e "  ${CYAN}└──────────────────────────────────────────────────────┘${NC}"
  echo ""
}

# ============================================================
# 交互循环
# ============================================================

interactive() {
  while true; do
    show_menu

    local be_result fe_result
    read -r be_result _ <<< "$(check_backend)"
    read -r fe_result _ <<< "$(check_frontend)"

    local be_running=false fe_running=false
    [[ "$be_result" == "running" ]] && be_running=true
    [[ "$fe_result" == "running" ]] && fe_running=true

    local all_running=false all_stopped=false
    $be_running && $fe_running && all_running=true
    ! $be_running && ! $fe_running && all_stopped=true

    # 构建选项到函数的映射
    declare -A actions=()
    local opt=1

    if $all_stopped; then
      actions[$opt]="start_all" ; opt=$((opt + 1))
    elif ! $all_running; then
      if ! $be_running; then
        actions[$opt]="start_backend" ; opt=$((opt + 1))
      fi
      if ! $fe_running; then
        actions[$opt]="start_frontend" ; opt=$((opt + 1))
      fi
    fi

    if $all_running; then
      actions[$opt]="stop_all"      ; opt=$((opt + 1))
      actions[$opt]="restart_all"   ; opt=$((opt + 1))
      actions[$opt]="restart_backend"; opt=$((opt + 1))
      actions[$opt]="restart_frontend"; opt=$((opt + 1))
    elif ! $all_stopped; then
      if $be_running; then
        actions[$opt]="stop_backend"   ; opt=$((opt + 1))
        actions[$opt]="restart_backend"; opt=$((opt + 1))
      fi
      if $fe_running; then
        actions[$opt]="stop_frontend"   ; opt=$((opt + 1))
        actions[$opt]="restart_frontend"; opt=$((opt + 1))
      fi
      actions[$opt]="stop_all" ; opt=$((opt + 1))
    fi

    echo -ne "  ${BLUE}请选择操作 [1-$((opt-1)) / q]:${NC} "
    read -r choice

    if [[ "$choice" == "q" ]] || [[ "$choice" == "Q" ]]; then
      echo ""
      log_info "已退出管理脚本"
      break
    fi

    if [[ -z "${actions[$choice]:-}" ]]; then
      log_error "无效选择，请重新输入"
      sleep 1
      clear 2>/dev/null || true
      continue
    fi

    "${actions[$choice]}"
    echo ""
    echo -ne "  ${BLUE}按 Enter 继续...${NC}"
    read -r
    clear 2>/dev/null || true
  done
}

# ============================================================
# CLI 子命令模式
# ============================================================

cli_mode() {
  local cmd="${1:-}"
  local target="${2:-all}"

  case "$cmd" in
    status)
      print_status
      ;;
    start)
      case "$target" in
        all)      start_all ;;
        backend)  start_backend ;;
        frontend) start_frontend ;;
        *)        log_error "用法: $0 start [all|backend|frontend]" ; exit 1 ;;
      esac
      ;;
    stop)
      case "$target" in
        all)      stop_all ;;
        backend)  stop_backend ;;
        frontend) stop_frontend ;;
        *)        log_error "用法: $0 stop [all|backend|frontend]" ; exit 1 ;;
      esac
      ;;
    restart)
      case "$target" in
        all)      restart_all ;;
        backend)  restart_backend ;;
        frontend) restart_frontend ;;
        *)        log_error "用法: $0 restart [all|backend|frontend]" ; exit 1 ;;
      esac
      ;;
    *)
      echo "轻燕工作台 · 服务管理脚本"
      echo ""
      echo "用法:"
      echo "  $0                交互模式"
      echo "  $0 status         查看服务状态"
      echo "  $0 start  all     启动全部服务"
      echo "  $0 start  backend 仅启动后端"
      echo "  $0 start  frontend 仅启动前端"
      echo "  $0 stop   all     停止全部服务"
      echo "  $0 stop   backend 仅停止后端"
      echo "  $0 stop   frontend 仅停止前端"
      echo "  $0 restart all    重启全部服务"
      exit 1
      ;;
  esac
}

# ============================================================
# 入口
# ============================================================

if [[ $# -gt 0 ]]; then
  cli_mode "$@"
else
  clear 2>/dev/null || true
  interactive
fi
