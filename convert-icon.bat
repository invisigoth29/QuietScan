@echo off
echo Converting icon.png to icon.ico...
echo.
echo Option 1: Use online converter (Recommended):
echo   1. Go to https://convertio.co/png-ico/
echo   2. Upload assets\icon.png
echo   3. Download as icon.ico
echo   4. Save to assets\icon.ico
echo.
echo Option 2: Use ImageMagick (if installed):
echo   magick convert assets\icon.png -define icon:auto-resize=256,128,64,48,32,16 assets\icon.ico
echo.
echo Option 3: Use PowerShell (Windows 10+):
echo   This will attempt to convert using PowerShell...
powershell -Command "$ErrorActionPreference='Stop'; try { Add-Type -AssemblyName System.Drawing; $img = [System.Drawing.Image]::FromFile('assets\icon.png'); $stream = New-Object System.IO.FileStream('assets\icon.ico', [System.IO.FileMode]::Create); $icon = New-Object System.Drawing.Icon($img.Handle); $icon.Save($stream); $stream.Close(); $img.Dispose(); Write-Host 'Conversion successful!' } catch { Write-Host 'PowerShell conversion failed. Please use online converter.' }"
echo.
pause


