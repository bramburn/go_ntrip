@echo off
echo Building GNSS application...

REM Set variables
set OUTPUT_DIR=..\build
set MAIN_PKG=..\cmd\gnss
set APP_NAME=gnss_receiver

REM Create output directory if it doesn't exist
if not exist %OUTPUT_DIR% mkdir %OUTPUT_DIR%

REM Build for Windows
echo Building for Windows...
go build -o %OUTPUT_DIR%\%APP_NAME%.exe %MAIN_PKG%

if %ERRORLEVEL% neq 0 (
    echo Build failed!
    exit /b %ERRORLEVEL%
)

echo Build completed successfully!
echo Executable location: %OUTPUT_DIR%\%APP_NAME%.exe
