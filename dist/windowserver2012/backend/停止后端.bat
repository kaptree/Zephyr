@echo off
rem ============================================
rem  LabelPro - Stop backend service
rem ============================================
taskkill /f /im labelpro-server.exe >nul 2>&1
echo LabelPro backend stopped.
timeout /t 2 >nul
