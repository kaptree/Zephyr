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
