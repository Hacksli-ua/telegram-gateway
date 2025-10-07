@echo off
REM Build script for Telegram Gateway Java ME Client
REM Requires Sun Java Wireless Toolkit (WTK) installed

REM Set WTK path - adjust this to your installation
set WTK_HOME=C:\WTK2.5.2

REM Check if WTK exists
if not exist "%WTK_HOME%\bin\preverify.exe" (
    echo ERROR: WTK not found at %WTK_HOME%
    echo Please install Sun Java Wireless Toolkit 2.5.2 or adjust WTK_HOME in this script
    pause
    exit /b 1
)

echo ========================================
echo Telegram Gateway Java ME Client Builder
echo ========================================
echo.

REM Create directories
if not exist build mkdir build
if not exist build\classes mkdir build\classes
if not exist build\verified mkdir build\verified
if not exist dist mkdir dist

echo Step 1: Compiling Java sources...
javac -source 1.3 -target 1.3 ^
      -bootclasspath "%WTK_HOME%\lib\cldcapi11.jar;%WTK_HOME%\lib\midpapi20.jar" ^
      -d build\classes ^
      src\*.java

if errorlevel 1 (
    echo ERROR: Compilation failed!
    pause
    exit /b 1
)

echo Step 2: Preverifying classes...
"%WTK_HOME%\bin\preverify.exe" ^
    -classpath "%WTK_HOME%\lib\cldcapi11.jar;%WTK_HOME%\lib\midpapi20.jar" ^
    -d build\verified ^
    build\classes

if errorlevel 1 (
    echo ERROR: Preverification failed!
    pause
    exit /b 1
)

echo Step 3: Creating JAR file...
cd build\verified
jar cvfm ..\..\dist\TelegramClient.jar ..\..\manifest.mf *.class
cd ..\..

if errorlevel 1 (
    echo ERROR: JAR creation failed!
    pause
    exit /b 1
)

echo.
echo ========================================
echo BUILD SUCCESSFUL!
echo ========================================
echo JAR file created: dist\TelegramClient.jar
echo JAD file created: dist\TelegramClient.jad
echo.
echo You can now install TelegramClient.jad on your phone
echo ========================================
pause
