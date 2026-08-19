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
