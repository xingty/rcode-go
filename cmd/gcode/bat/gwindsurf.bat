@echo off

set "BIN_NAME=gwindsurf"
set "CODE_HOME=%~dp0.."
set "CODE_BIN=%CODE_HOME%\%BIN_NAME%.exe"

"%CODE_BIN%" %BIN_NAME% %*