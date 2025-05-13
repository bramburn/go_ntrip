@echo off
echo Building GNSS applications...

REM Set variables
set OUTPUT_DIR=..\build
set GNSS_PKG=..\cmd\gnss
set NTRIP_CLIENT_PKG=..\cmd\ntrip-client
set GNSS_APP_NAME=gnss_receiver
set NTRIP_CLIENT_APP_NAME=ntrip-client

REM Create output directory if it doesn't exist
if not exist %OUTPUT_DIR% mkdir %OUTPUT_DIR%

REM Build GNSS application for Windows
echo Building GNSS application for Windows...
go build -o %OUTPUT_DIR%\%GNSS_APP_NAME%.exe %GNSS_PKG%

if %ERRORLEVEL% neq 0 (
    echo GNSS application build failed!
    exit /b %ERRORLEVEL%
)

REM Build NTRIP client application for Windows
echo Building NTRIP client application for Windows...
go build -o %OUTPUT_DIR%\%NTRIP_CLIENT_APP_NAME%.exe %NTRIP_CLIENT_PKG%

if %ERRORLEVEL% neq 0 (
    echo NTRIP client application build failed!
    exit /b %ERRORLEVEL%
)

echo Build completed successfully!
echo Executable locations:
echo - GNSS application: %OUTPUT_DIR%\%GNSS_APP_NAME%.exe
echo - NTRIP client: %OUTPUT_DIR%\%NTRIP_CLIENT_APP_NAME%.exe
