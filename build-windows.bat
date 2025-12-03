@echo off
echo Building QuietScan for Windows...

REM Check if go-winres is installed
where go-winres >nul 2>&1
if errorlevel 1 (
    echo go-winres not found. Installing...
    go install github.com/tc-hib/go-winres@latest
    if errorlevel 1 (
        echo Failed to install go-winres. Building without icon...
        goto :build
    )
)

REM Generate Windows resource file if icon exists
if exist "assets\icon.ico" (
    echo Generating Windows resources from icon...
    go-winres simply --icon assets\icon.ico --manifest gui --out cmd\quietscan\resource.syso
    if errorlevel 1 (
        echo Warning: Failed to generate resources. Building without icon...
        goto :build
    )
    echo Resources generated successfully.
    echo Verifying cmd\quietscan\resource.syso exists...
    if not exist "cmd\quietscan\resource.syso" (
        echo ERROR: cmd\quietscan\resource.syso was not created!
        goto :build
    )
) else (
    echo Warning: assets\icon.ico not found. Building without icon.
    echo To add an icon: Convert assets\icon.png to assets\icon.ico
)

:build
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=1

echo Building executable...
echo Note: resource.syso will be automatically included by Go compiler
go build -mod=mod -ldflags "-H=windowsgui" -o quietscan.exe ./cmd/quietscan

REM Don't delete resource.syso - keep it for future builds
REM if exist "cmd\quietscan\resource.syso" (
REM     del cmd\quietscan\resource.syso
REM     echo Cleaned up resource file.
REM )

if errorlevel 1 (
    echo Build failed!
    exit /b 1
)

echo.
echo Build successful: quietscan.exe
echo This is a fully self-contained executable - no external files needed!
if exist "assets\icon.ico" (
    echo Icon embedded successfully!
) else (
    echo Note: No icon embedded. Convert assets\icon.png to assets\icon.ico to add an icon.
)
pause
