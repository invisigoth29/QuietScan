@echo off
echo Verifying Windows icon embedding...
echo.

if not exist "quietscan-windows.exe" (
    if not exist "quietscan.exe" (
        echo ERROR: No executable found!
        echo Please build first using build-windows.bat or build-all.sh
        pause
        exit /b 1
    ) else (
        set EXE=quietscan.exe
    )
) else (
    set EXE=quietscan-windows.exe
)

echo Checking: %EXE%
echo.

REM Check if resource.syso exists
if exist "resource.syso" (
    echo [OK] resource.syso found - resources should be embedded
) else (
    echo [WARNING] resource.syso not found - icon may not be embedded
    echo Run build-windows.bat or build-all.sh to generate it
)

echo.
echo To verify icon is embedded:
echo 1. Use Resource Hacker: http://www.angusj.com/resourcehacker/
echo 2. Open %EXE% in Resource Hacker
echo 3. Look under "Icon" section - you should see icon resources
echo.
echo To clear Windows icon cache:
echo 1. Press Win+R
echo 2. Type: ie4uinit.exe -show
echo 3. Press Enter
echo.
echo Or restart Windows Explorer:
echo 1. Press Ctrl+Shift+Esc
echo 2. Find "Windows Explorer"
echo 3. Right-click - Restart
echo.
pause


