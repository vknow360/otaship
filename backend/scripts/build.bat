@echo off
setlocal enabledelayedexpansion

set TARGET=%1
set VERSION=%2
if "%VERSION%"=="" set VERSION=dev

for /f "delims=" %%i in ('powershell -Command "Get-Date -Format yyyy-MM-dd"') do set BUILD_DATE=%%i

set LDFLAGS=-s -w -X main.Version=%VERSION% -X main.BuildDate=%BUILD_DATE%

if not exist dist mkdir dist

if /i "%TARGET%"=="windows" (
    echo Building for Windows...
    set GOOS=windows
    set GOARCH=amd64
    go build -ldflags "%LDFLAGS%" -o dist/otaship-server.exe cmd/server/main.go
    goto check_result
)

if /i "%TARGET%"=="mac" (
    echo Building for macOS - Apple Silicon...
    set GOOS=darwin
    set GOARCH=arm64
    go build -ldflags "%LDFLAGS%" -o dist/otaship-server_mac cmd/server/main.go
    goto check_result
)

echo Building for Linux...
set GOOS=linux
set GOARCH=amd64
go build -ldflags "%LDFLAGS%" -o dist/otaship-server cmd/server/main.go

:check_result
if %ERRORLEVEL% equ 0 (
    echo Build successful [Version: %VERSION%] [Date: %BUILD_DATE%]
    exit /b 0
) else (
    echo Build failed
    exit /b 1
)

endlocal