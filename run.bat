@echo off
echo ========================================
echo   Telegram Gateway Server (Binary)
echo ========================================
echo.

REM Перевірка наявності .env файлу
if not exist .env (
    echo [ERROR] .env file not found!
    echo Please copy .env.example to .env and configure it.
    pause
    exit /b 1
)

REM Перевірка наявності скомпільованого бінарника
if not exist bin\telegram-gateway.exe (
    echo [ERROR] Binary not found!
    echo Please run build.bat first to compile the application.
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

echo [INFO] Configuration loaded
echo [INFO] Starting compiled server...
echo.

REM Запуск скомпільованого бінарника
bin\telegram-gateway.exe

if errorlevel 1 (
    echo.
    echo [ERROR] Server crashed or failed to start!
    pause
    exit /b 1
)

pause
