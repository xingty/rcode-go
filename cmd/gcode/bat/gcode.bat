@echo off

set "BIN_NAME=gcode"
set "CODE_HOME=%~dp0.."
set "CODE_BIN=%CODE_HOME%\gcode.exe"

"%CODE_BIN%" %BIN_NAME% %*