@echo off
REM Clear Windows Icon Cache
REM Run this script on Windows if the icon doesn't update after rebuilding

echo Clearing Windows icon cache...
echo.

REM Stop Windows Explorer
echo Stopping Windows Explorer...
taskkill /f /im explorer.exe

REM Delete icon cache files
echo Deleting icon cache files...
cd /d %userprofile%\AppData\Local\Microsoft\Windows\Explorer
attrib -h IconCache.db
del IconCache.db /a /f /q
for /d %%x in (*) do @rd /s /q "%%x"

REM Delete additional icon cache (Windows 10/11)
del iconcache_*.db /a /f /q 2>nul

REM Restart Windows Explorer
echo Restarting Windows Explorer...
start explorer.exe

echo.
echo Icon cache cleared!
echo Please check if the icon has updated.
echo If not, try logging out and back in to Windows.
pause
