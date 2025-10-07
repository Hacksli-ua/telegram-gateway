@echo off
echo ========================================
echo   Telegram Gateway Server
echo ========================================
echo.

REM Перевірка наявності .env файлу
if not exist .env (
    echo [ERROR] .env file not found!
    echo Please copy .env.example to .env and configure it.
    pause
    exit /b 1
)

echo [INFO] Loading environment variables from .env...
echo.

REM Завантаження змінних з .env файлу
for /f "usebackq tokens=1,* delims==" %%a in (".env") do (
    set "line=%%a"
    if not "!line:~0,1!"=="#" (
        if not "%%a"=="" (
            set "%%a=%%b"
        )
    )
)

echo [INFO] Configuration loaded:
echo   - API ID: %TELEGRAM_API_ID%
echo   - Server Port: %SERVER_PORT%
echo   - Server Host: %SERVER_HOST%
echo.

echo [INFO] Starting server...
echo.

REM Запуск Go сервера
go run main.go

if errorlevel 1 (
    echo.
    echo [ERROR] Server crashed or failed to start!
    pause
    exit /b 1
)

pause
