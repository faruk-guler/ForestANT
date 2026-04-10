@echo off
REM ─────────────────────────────────────────────────────────────
REM  ForestANT v3.0 — Windows Build Script
REM  Kullanım: build.bat
REM  Çıktı: deploy\ klasörüne tüm dosyaları hazırlar
REM ─────────────────────────────────────────────────────────────

title ForestANT Build

echo.
echo  ███████╗ ██████╗ ██████╗ ███████╗███████╗████████╗ █████╗ ███╗   ██╗████████╗
echo  ██╔════╝██╔═══██╗██╔══██╗██╔════╝██╔════╝╚══██╔══╝██╔══██╗████╗  ██║╚══██╔══╝
echo  █████╗  ██║   ██║██████╔╝█████╗  ███████╗   ██║   ███████║██╔██╗ ██║   ██║
echo  ██╔══╝  ██║   ██║██╔══██╗██╔══╝  ╚════██║   ██║   ██╔══██║██║╚██╗██║   ██║
echo  ██║     ╚██████╔╝██║  ██║███████╗███████║   ██║   ██║  ██║██║ ╚████║   ██║
echo  ╚═╝      ╚═════╝ ╚═╝  ╚═╝╚══════╝╚══════╝   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═══╝   ╚═╝  v3.0
echo.

REM Çıktı klasörü
set DEPLOY_DIR=%~dp0deploy\release
if not exist "%DEPLOY_DIR%" mkdir "%DEPLOY_DIR%"

REM ─── 1. Frontend Build ───────────────────────────────────────
echo [1/3] Frontend derleniyor (React + Vite)...
cd "%~dp0frontend"
call npm install --silent
call npm run build

if errorlevel 1 (
    echo [HATA] Frontend build basarisiz!
    pause
    exit /b 1
)
echo    Frontend basariyla derlendi.

# Linux Binary ─────────────────────────────────────────
echo [2/3] Linux binary derleniyor (Go)...
cd "%~dp0backend-go"
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0
go build -ldflags="-s -w" -o "%DEPLOY_DIR%\forestant-engine" .

if errorlevel 1 (
    echo [HATA] Go build basarisiz!
    pause
    exit /b 1
)
echo    Linux binary basariyla derlendi.

REM ─── 3. Deploy Klasörüne Topla ───────────────────────────────
echo [3/3] Deploy dosyalari hazirlaniyor...

REM frontend dist
xcopy /E /I /Y "%~dp0frontend\dist" "%DEPLOY_DIR%\dist" > nul

REM env ve service dosyalari
copy /Y "%~dp0deploy\.env.example"     "%DEPLOY_DIR%\.env.example" > nul
copy /Y "%~dp0deploy\forestant.service" "%DEPLOY_DIR%\forestant.service" > nul
copy /Y "%~dp0deploy\deploy.sh"        "%DEPLOY_DIR%\deploy.sh" > nul

echo.
echo  ════════════════════════════════════════════════════════
echo    Basariyla tamamlandi!
echo  ════════════════════════════════════════════════════════
echo.
echo  Cikti klasoru: %DEPLOY_DIR%
echo.
echo  Sunucuya kopyalamak icin:
echo    scp -r %DEPLOY_DIR%\* kullanici@sunucu:/tmp/forestant-deploy/
echo.
echo  Sunucuda kurmak icin:
echo    cd /tmp/forestant-deploy
echo    cp .env.example .env ^&^& nano .env
echo    chmod +x deploy.sh
echo    sudo ./deploy.sh
echo.
pause
