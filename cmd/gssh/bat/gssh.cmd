@echo off
setlocal

rem Prefer PowerShell wrapper (lets ssh.exe be the long-running process).
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0gssh.ps1" %*
if %ERRORLEVEL%==0 exit /b 0

rem Fallback: if PowerShell is blocked (e.g. by Group Policy), run core directly.
if not exist "%~dp0gssh-core.exe" exit /b %ERRORLEVEL%
"%~dp0gssh-core.exe" %*
exit /b %ERRORLEVEL%
