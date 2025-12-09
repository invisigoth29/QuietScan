@echo off
REM Verify you have the correct Windows executable with updated icon

echo ============================================
echo QuietScan Windows EXE Verification
echo ============================================
echo.

echo Checking file: quietscan-windows.exe
echo Expected SHA256: 5490c6e043c4568cdeb440d603206ad8785e7086fb2a50416c7422dba021934a
echo.

REM Get file hash
certutil -hashfile quietscan-windows.exe SHA256

echo.
echo If the SHA256 hash matches the expected value above,
echo you have the correct file with the new icon embedded.
echo.
echo If you see the old icon, Windows is caching it.
echo Try these steps:
echo   1. Delete this exe file
echo   2. Empty Recycle Bin
echo   3. Copy the file again to a DIFFERENT folder
echo   4. Rename it to something else (e.g., quietscan-new.exe)
echo.
pause
