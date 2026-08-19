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
