@echo off
REM Clean build artifacts

echo Cleaning build directories...

if exist build rmdir /s /q build
if exist dist rmdir /s /q dist

echo Done!
pause
