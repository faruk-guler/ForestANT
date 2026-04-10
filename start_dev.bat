@echo off
title ForestANT Dev Starter [ROBUST]
setlocal enabledelayedexpansion

echo.
echo  ███████╗ ██████╗ ██████╗ ███████╗███████╗████████╗ █████╗ ███╗   ██╗████████╗
echo  ██╔════╝██╔═══██╗██╔══██╗██╔════╝██╔════╝╚══██╔══╝██╔══██╗████╗  ██║╚══██╔══╝
echo  █████╗  ██║   ██║██████╔╝█████╗  ███████╗   ██║   ███████║██╔██╗ ██║   ██║
echo  ██╔══╝  ██║   ██║██╔══██╗██╔══╝  ╚════██║   ██║   ██╔══██║██║╚██╗██║   ██║
echo  ██║     ╚██████╔╝██║  ██║███████╗███████║   ██║   ██║  ██║██║ ╚████║   ██║
echo  ╚═╝      ╚═════╝ ╚═╝  ╚═╝╚══════╝╚══════╝   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═══╝   ╚═╝
echo.

echo  [1/4] Sistem Bagimliliklari Kontrol Ediliyor...
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo [HATA] Go yuklu degil veya PATH'e eklenmemis!
    echo Lutfen Go'nun kurulu oldugundan emin ol.
    pause
    exit /b
)
where npm >nul 2>nul
if %errorlevel% neq 0 (
    echo [HATA] NPM veya Node.js yuklu degil veya PATH'e eklenmemis!
    echo Lutfen Node.js'in kurulu oldugundan emin ol.
    pause
    exit /b
)
echo    Sistem gereksinimleri karsilaniyor.

echo.
echo  [2/4] Eski Islemler Temizleniyor (Port: 9500 ^& 8080)...
for /f "tokens=5" %%a in ('netstat -aon ^| findstr ":9500" ^| findstr "LISTENING"') do (
    taskkill /f /pid %%a >nul 2>&1 && echo    Port 9500 temizlendi - PID: %%a
)
for /f "tokens=5" %%a in ('netstat -aon ^| findstr ":8080" ^| findstr "LISTENING"') do (
    taskkill /f /pid %%a >nul 2>&1 && echo    Port 8080 temizlendi - PID: %%a
)

echo.
echo  [3/4] Backend (Go) Baslatiliyor...
pushd "%~dp0backend-go"
start "ForestANT Backend" cmd /k "echo Backend Baslatiliyor... && go run main.go"
popd

echo.
echo  [4/4] Frontend (Vite) Baslatiliyor...
pushd "%~dp0frontend"
start "ForestANT Frontend" cmd /k "echo Frontend Baslatiliyor... && npm run dev"
popd

echo.
echo  ISLEM TAMAMLANDI!
echo  Proje http://localhost:9500 adresinde calisiyor.
exit
