@echo off
setlocal EnableDelayedExpansion

REM Find Go
where go >nul 2>&1
if errorlevel 1 (
    echo Go not found in PATH
    exit /b 1
)

set BUILD_TARGET=%1
set OS_TARGETS=%2

REM Defaults
if "%BUILD_TARGET%"=="" set BUILD_TARGET=-a
if "%OS_TARGETS%"=="" set OS_TARGETS=-a

set BUILD_SERVER=false
set BUILD_CLIENT=false

if "%BUILD_TARGET%"=="-a" (
    set BUILD_SERVER=true
    set BUILD_CLIENT=true
) else if "%BUILD_TARGET%"=="-s" (
    set BUILD_SERVER=true
) else if "%BUILD_TARGET%"=="-c" (
    set BUILD_CLIENT=true
) else (
    echo Unknown build target: %BUILD_TARGET%
    exit /b 1
)

if "%OS_TARGETS%"=="-a" set OS_TARGETS=-wlm

if not exist build mkdir build

call :build server ./cmd/server
call :build client ./cmd/client

echo Build complete.
exit /b 0

REM -------------------------
REM Build function
REM -------------------------
:build
set NAME=%1
set BUILD_PATH=%2

if "%NAME%"=="server" if "%BUILD_SERVER%"=="false" exit /b
if "%NAME%"=="client" if "%BUILD_CLIENT%"=="false" exit /b

echo Building %NAME%...

echo %OS_TARGETS% | find "w" >nul
if not errorlevel 1 (
    echo  - Windows
    set GOOS=windows
    set GOARCH=amd64
    go build -o build\%NAME%-windows.exe %BUILD_PATH%
)

echo %OS_TARGETS% | find "l" >nul
if not errorlevel 1 (
    echo  - Linux
    set GOOS=linux
    set GOARCH=amd64
    go build -o build\%NAME%-linux %BUILD_PATH%
)

echo %OS_TARGETS% | find "m" >nul
if not errorlevel 1 (
    echo  - macOS
    set GOOS=darwin
    set GOARCH=amd64
    go build -o build\%NAME%-macos %BUILD_PATH%
)

exit /b
