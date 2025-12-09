@echo off
REM Clear Windows Icon and Thumbnail Cache
REM Run this script AS ADMINISTRATOR on Windows if the icon doesn't update after rebuilding

echo ============================================
echo Clearing Windows Icon and Thumbnail Cache
echo ============================================
echo.
echo IMPORTANT: This script should be run AS ADMINISTRATOR
echo.
pause

REM Stop Windows Explorer
echo [1/5] Stopping Windows Explorer...
taskkill /f /im explorer.exe

REM Delete icon cache files
echo [2/5] Deleting icon cache files...
cd /d %userprofile%\AppData\Local\Microsoft\Windows\Explorer
attrib -h IconCache.db 2>nul
del IconCache.db /a /f /q 2>nul
del iconcache_*.db /a /f /q 2>nul
for /d %%x in (*) do @rd /s /q "%%x" 2>nul

REM Delete thumbnail cache (this is often the culprit)
echo [3/5] Deleting thumbnail cache...
del thumbcache_*.db /a /f /q 2>nul

REM Clear additional caches
echo [4/5] Deleting additional caches...
cd /d %LocalAppData%
del IconCache.db /a /f /q 2>nul
cd /d %LocalAppData%\Microsoft\Windows\Caches
del *.db /a /f /q 2>nul

REM Restart Windows Explorer
echo [5/5] Restarting Windows Explorer...
start explorer.exe

echo.
echo ============================================
echo Cache clearing complete!
echo ============================================
echo.
echo Next steps:
echo   1. Delete the old quietscan-windows.exe file
echo   2. Empty the Recycle Bin
echo   3. Copy the NEW exe to a DIFFERENT folder
echo   4. Optionally rename it (e.g., quietscan-v2.exe)
echo   5. Check the icon
echo.
echo If icon still doesn't update, restart Windows.
echo.
pause
