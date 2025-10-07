@echo off
echo ========================================
echo   Building Telegram Gateway Server
echo ========================================
echo.

echo [INFO] Cleaning old builds...
if exist bin\telegram-gateway.exe del /q bin\telegram-gateway.exe
if not exist bin mkdir bin

echo [INFO] Downloading dependencies...
go mod tidy

if errorlevel 1 (
    echo [ERROR] Failed to download dependencies!
    pause
    exit /b 1
)

echo.
echo [INFO] Building application...
go build -o bin\telegram-gateway.exe main.go

if errorlevel 1 (
    echo [ERROR] Build failed!
    pause
    exit /b 1
)

echo.
echo [SUCCESS] Build completed successfully!
echo Binary created: bin\telegram-gateway.exe
echo.
echo To run the server, use: start.bat or bin\telegram-gateway.exe
echo.

pause
