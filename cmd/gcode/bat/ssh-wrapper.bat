@echo off
setlocal

powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0ssh-wrapper.ps1" %*
exit /b %ERRORLEVEL%
