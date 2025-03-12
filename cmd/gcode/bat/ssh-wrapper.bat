@echo off
setlocal EnableDelayedExpansion

set "use_gssh=false"
set "args="

:parse_args
if "%~1"=="" goto execute
if "%~1"=="--gssh" (
    set "use_gssh=true"
) else (
    set "args=!args! %1"
)
shift
goto parse_args

:execute
if "%use_gssh%"=="true" (
    gssh%args%
) else (
    ssh%args%
)

endlocal