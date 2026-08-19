#!/usr/bin/env bash
#
# 轻燕工作台 - Windows Server 2012 部署包构建脚本
#
# 功能：
#   1. 交叉编译后端为 Windows amd64 可执行文件 (labelpro-server.exe)
#   2. 构建前端静态资源 (Vite build)
#   3. 生成 Nginx 配置、Windows 启动/停止脚本、部署说明
#   4. 全部打包到 dist/windowserver2012/
#
# 用法：
#   ./build-windows2012.sh
#
# 前置要求：
#   - Go 1.21+（支持 windows/amd64 交叉编译）
#   - Node.js + npm
#
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
FRONT_DIR="$ROOT/Web-Front"
BACKEND_DIR="$ROOT/Server-code"
OUT_DIR="$ROOT/dist/windowserver2012"

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; RED='\033[0;31m'; NC='\033[0m'
log()  { echo -e "${CYAN}[INFO]${NC} $*"; }
ok()   { echo -e "${GREEN}[OK]${NC} $*"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $*"; }
die()  { echo -e "${RED}[ERROR]${NC} $*"; exit 1; }

# ============================================================
# 0. 前置检查
# ============================================================
command -v go >/dev/null 2>&1 || die "未找到 go，请先安装 Go 并加入 PATH"
command -v npm >/dev/null 2>&1 || die "未找到 npm，请先安装 Node.js"
[[ -d "$BACKEND_DIR" ]] || die "后端目录不存在: $BACKEND_DIR"
[[ -d "$FRONT_DIR" ]] || die "前端目录不存在: $FRONT_DIR"

# ============================================================
# 1. 清空并创建输出目录
# ============================================================
log "清空输出目录 $OUT_DIR"
rm -rf "$OUT_DIR"
mkdir -p "$OUT_DIR/backend/keys" \
         "$OUT_DIR/backend/logs" \
         "$OUT_DIR/backend/uploads" \
         "$OUT_DIR/frontend" \
         "$OUT_DIR/nginx"

# ============================================================
# 2. 交叉编译后端 (Windows amd64)
# ============================================================
log "交叉编译后端 -> Windows amd64 (labelpro-server.exe)"
cd "$BACKEND_DIR"
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-s -w" \
    -o "$OUT_DIR/backend/labelpro-server.exe" .
[[ -f "$OUT_DIR/backend/labelpro-server.exe" ]] || die "后端编译失败"
ok "后端编译完成: backend/labelpro-server.exe"

# ============================================================
# 3. 拷贝后端配置与密钥
# ============================================================
log "拷贝后端配置与密钥"
[[ -f "config.json" ]] || die "缺少 config.json"
cp config.json "$OUT_DIR/backend/"
if [[ -f "keys/private.pem" && -f "keys/public.pem" ]]; then
    cp keys/private.pem keys/public.pem "$OUT_DIR/backend/keys/"
else
    warn "keys 目录缺少 private.pem / public.pem，后端启动时会自动生成临时密钥（重启后 token 会失效）"
fi
ok "配置与密钥已就位"

# ============================================================
# 4. 构建前端
# ============================================================
log "构建前端 (npm run build)"
cd "$FRONT_DIR"
npm run build || die "前端构建失败"
[[ -d "dist" && -f "dist/index.html" ]] || die "前端构建产物缺失"
ok "前端构建完成"

# ============================================================
# 5. 拷贝前端静态文件
# ============================================================
log "拷贝前端静态文件 -> frontend/"
cp -R dist/. "$OUT_DIR/frontend/"
rm -f "$OUT_DIR/frontend/.gitkeep"
ok "前端静态文件已就位"

# ============================================================
# 6. 生成 Nginx 配置
# ============================================================
log "生成 Nginx 配置"
cat > "$OUT_DIR/nginx/nginx.conf" <<'NGINXEOF'
# ============================================================
# 轻燕工作台 - Windows Server 2012 Nginx 配置示例
# ============================================================
# 使用说明：
#   1) 部署方式 A（推荐）：将 dist/windowserver2012/frontend/ 下的所有文件
#      复制到 Nginx 安装目录的 html/ 文件夹内（覆盖），本配置使用 root html;
#   2) 部署方式 B：保留本文件于 nginx/conf/ 下，并把下方 root 改为实际绝对路径，
#      如 root  D:/labelpro/frontend;
#   3) 后端进程 labelpro-server.exe 需保持运行，默认监听 127.0.0.1:8090
#   4) 如 80 端口被占用，请修改 listen 端口
# ============================================================

worker_processes  1;

events {
    worker_connections  1024;
}

http {
    include       mime.types;
    default_type  application/octet-stream;

    sendfile        on;
    keepalive_timeout  65;

    # 前端静态资源压缩
    gzip  on;
    gzip_min_length 1k;
    gzip_types text/plain text/css application/javascript application/json image/svg+xml;

    server {
        listen       80;
        server_name  localhost;

        # 前端页面（Vue history 路由回退到 index.html）
        location / {
            root      html;
            index     index.html;
            try_files $uri $uri/ /index.html;
        }

        # 后端 REST API
        location /api/ {
            proxy_pass http://127.0.0.1:8090;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_connect_timeout 15s;
            proxy_read_timeout   300s;
            client_max_body_size 20m;
        }

        # 上传文件目录（后端静态服务）
        location /uploads/ {
            proxy_pass http://127.0.0.1:8090;
            proxy_set_header Host $host;
        }

        # WebSocket（通知 / 聊天实时推送）
        location /ws/ {
            proxy_pass http://127.0.0.1:8090;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";
            proxy_set_header Host $host;
            proxy_read_timeout 3600s;
            proxy_send_timeout 3600s;
        }
    }
}
NGINXEOF
ok "Nginx 配置已生成: nginx/nginx.conf"

# ============================================================
# 7. 生成 Windows 启动 / 停止脚本
# ============================================================

# ---- 后端启动脚本 ----
cat > "$OUT_DIR/backend/启动后端.bat" <<'BATEOF'
@echo off
rem ============================================
rem  LabelPro - Start backend service
rem  Default listen port: 8090
rem ============================================
cd /d "%~dp0"
if exist labelpro-server.exe (
    echo Starting LabelPro backend on port 8090 ...
    start "" /D "%~dp0" labelpro-server.exe
    echo Backend started. You can close this window.
    timeout /t 3 >nul
) else (
    echo [ERROR] labelpro-server.exe not found in %~dp0
    pause
)
BATEOF

# ---- 后端停止脚本 ----
cat > "$OUT_DIR/backend/停止后端.bat" <<'BATEOF'
@echo off
rem ============================================
rem  LabelPro - Stop backend service
rem ============================================
taskkill /f /im labelpro-server.exe >nul 2>&1
echo LabelPro backend stopped.
timeout /t 2 >nul
BATEOF

# ---- 一键启动（后端 + Nginx）----
cat > "$OUT_DIR/一键启动.bat" <<'BATEOF'
@echo off
rem ============================================
rem  LabelPro - Start All (Backend + Nginx)
rem  Install Nginx and deploy frontend first
rem ============================================
cd /d "%~dp0"

echo [1/2] Starting backend ...
cd backend
start "" /D "%~dp0backend" labelpro-server.exe
cd ..

echo [2/2] Starting Nginx ...
set NGINX_EXE=
if exist "C:\nginx\nginx.exe" set NGINX_EXE=C:\nginx\nginx.exe
if exist "D:\nginx\nginx.exe" set NGINX_EXE=D:\nginx\nginx.exe
if defined NGINX_EXE (
    cd /d "%NGINX_EXE%\.."
    start "Nginx" nginx.exe
    echo Nginx started.
) else (
    echo [WARN] Nginx not found in C:\nginx or D:\nginx.
    echo        Please install Nginx and start it manually.
)
echo.
echo Done. Open http://localhost in your browser.
timeout /t 5 >nul
BATEOF

# ---- 一键停止（后端 + Nginx）----
cat > "$OUT_DIR/一键停止.bat" <<'BATEOF'
@echo off
rem ============================================
rem  LabelPro - Stop All (Backend + Nginx)
rem ============================================
taskkill /f /im labelpro-server.exe >nul 2>&1
echo LabelPro backend stopped.
set NGINX_EXE=
if exist "C:\nginx\nginx.exe" set NGINX_EXE=C:\nginx\nginx.exe
if exist "D:\nginx\nginx.exe" set NGINX_EXE=D:\nginx\nginx.exe
if defined NGINX_EXE (
    "%NGINX_EXE%" -s stop >nul 2>&1
    echo Nginx stopped.
) else (
    echo [WARN] Nginx not found, skipped.
)
timeout /t 2 >nul
BATEOF

# 批处理文件统一转换为 Windows CRLF 换行，避免在 cmd 下解析异常
find "$OUT_DIR" -name "*.bat" -exec perl -pi -e 's/\r?\n$/\r\n/' {} +

ok "Windows 启动/停止脚本已生成"

# ============================================================
# 8. 生成部署说明
# ============================================================
log "生成部署说明"
cat > "$OUT_DIR/部署说明.txt" <<'DOCEOF'
============================================================
  轻燕工作台  Windows Server 2012 部署说明
============================================================

一、部署包目录结构
  windowserver2012/
  ├─ backend/                后端程序目录
  │   ├─ labelpro-server.exe 后端可执行文件（Windows amd64）
  │   ├─ config.json         后端配置文件（数据库/Redis/端口等）
  │   ├─ keys/               JWT 签名密钥（private.pem / public.pem）
  │   ├─ logs/               日志目录（自动生成）
  │   ├─ uploads/            上传文件目录（自动生成）
  │   ├─ 启动后端.bat        启动后端
  │   └─ 停止后端.bat        停止后端
  ├─ frontend/               前端静态资源（需放入 Nginx 的 html/ 目录）
  ├─ nginx/
  │   └─ nginx.conf          Nginx 配置示例（代理 /api、/ws、/uploads 到后端）
  ├─ 一键启动.bat            同时启动后端 + Nginx
  ├─ 一键停止.bat            同时停止后端 + Nginx
  └─ 部署说明.txt            本说明

二、环境依赖
  1. PostgreSQL 数据库（必须）
     - 安装 PostgreSQL 并在 config.json 中配置 database 连接信息
     - 首次启动后端会自动建表（AutoMigrate），无需手动导入 SQL
  2. Redis（可选）
     - 未安装时后端会警告并继续运行（仅缓存功能不可用）
     - 如已安装，请在 config.json 中配置 redis 连接信息
  3. Nginx（必须，用于托管前端页面与反向代理）
     - Windows 版 Nginx，下载后解压到 C:\nginx 或 D:\nginx

三、部署步骤
  1. 部署后端
     - 将 backend/ 整个目录复制到服务器，如 D:\labelpro\backend
     - 编辑 config.json，重点修改：
         * database.host / port / user / password / dbname
         * redis.host / port / password（如启用）
         * security.cors_allowed_origins 增加你的访问地址
           （如 http://服务器IP 或 http://域名）
     - 双击「启动后端.bat」启动，后端默认监听 8090 端口
  2. 部署前端（Nginx）
     - 将 frontend/ 下的所有文件复制到 Nginx 安装目录的 html/ 文件夹内
       （注意是复制 html 里面的内容，不要整层嵌套）
     - 将 nginx/nginx.conf 复制到 Nginx 安装目录的 conf/ 下（覆盖），
       或参照该文件内容手动修改现有配置
     - 双击 Nginx 安装目录下的 nginx.exe 启动 Nginx
  3. 访问系统
     - 浏览器打开 http://localhost（或服务器 IP）
     - 默认账号：admin / Admin@123（首次登录后请尽快修改）

四、常见问题
  1. 前端页面打不开 / 显示 404
     - 检查 Nginx 是否启动、80 端口是否被占用
     - 确认 html/ 目录内确实存在 index.html
  2. 页面能打开但登录提示网络异常
     - 检查后端进程是否运行（端口 8090）
     - 检查 Nginx 是否配置了 /api 与 /ws 代理
  3. WebSocket 实时通知不生效
     - 确认 nginx.conf 中 /ws/ 代理包含 Upgrade / Connection 头
  4. 登录提示 CORS / 跨域
     - 在 config.json 的 security.cors_allowed_origins 中加入实际访问地址
  5. 端口被占用
     - 修改 config.json 中 server.port 与 nginx.conf 中的 listen 端口、
       以及 /api、/ws、/uploads 的 proxy_pass 目标端口，保持一致
  6. 上传附件失败
     - 检查 config.json storage.local_path 目录是否存在可写权限

============================================================
  轻燕工作台 服务器部署包 - 由构建脚本自动生成
============================================================
DOCEOF

# 部署说明补 UTF-8 BOM，保证 Windows Server 2012 记事本打开中文不乱码
_bom_tmp="$(mktemp)"
printf '\xef\xbb\xbf' > "$_bom_tmp"
cat "$OUT_DIR/部署说明.txt" >> "$_bom_tmp"
mv "$_bom_tmp" "$OUT_DIR/部署说明.txt"

ok "部署说明已生成"

# ============================================================
# 9. 汇总输出
# ============================================================
echo ""
echo -e "${GREEN}========================================================${NC}"
echo -e "${GREEN}  打包完成: $OUT_DIR${NC}"
echo -e "${GREEN}========================================================${NC}"
du -sh "$OUT_DIR"
find "$OUT_DIR" -type f | sed "s|$OUT_DIR/|  |" | sort
echo ""
warn "提示：部署前请修改 backend/config.json 中的数据库与 CORS 配置"
