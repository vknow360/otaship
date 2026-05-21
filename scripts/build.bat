@echo off
setlocal

set TARGET=%1
set VERSION=%2

if "%TARGET%"=="" (
    echo Usage: build.bat ^<target^> [version]
    echo Targets: windows, mac, linux
    exit /b 1
)

if "%VERSION%"=="" set VERSION=dev

echo =========================================
echo Building Backend (Target: %TARGET%, Version: %VERSION%)
echo =========================================
cd backend
call scripts\build.bat %TARGET% %VERSION%
if %ERRORLEVEL% neq 0 (
    echo Backend build failed!
    cd ..
    exit /b %ERRORLEVEL%
)
cd ..

echo.
echo =========================================
echo Building CLI (Version: %VERSION%)
echo =========================================
cd cli
call scripts\build.bat %TARGET% %VERSION%
if %ERRORLEVEL% neq 0 (
    echo CLI build failed!
    cd ..
    exit /b %ERRORLEVEL%
)
cd ..

echo.
echo All builds completed successfully!
