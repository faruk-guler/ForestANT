#!/bin/bash
# ─────────────────────────────────────────────────────────────
#  ForestANT v3.0 — Linux Deploy Script
#  Kullanım: chmod +x deploy.sh && sudo ./deploy.sh
#  Test edildi: Ubuntu 22.04 / Debian 12 / RHEL 9
# ─────────────────────────────────────────────────────────────

set -e  # Hata olursa dur

# Renkli çıktı
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

INSTALL_DIR="/opt/forestant"
SERVICE_NAME="forestant"
BINARY_NAME="forestant-engine"
SERVICE_USER="forestant"

echo -e "${BLUE}"
echo "███████╗ ██████╗ ██████╗ ███████╗███████╗████████╗ █████╗ ███╗   ██╗████████╗"
echo "██╔════╝██╔═══██╗██╔══██╗██╔════╝██╔════╝╚══██╔══╝██╔══██╗████╗  ██║╚══██╔══╝"
echo "█████╗  ██║   ██║██████╔╝█████╗  ███████╗   ██║   ███████║██╔██╗ ██║   ██║"
echo "██╔══╝  ██║   ██║██╔══██╗██╔══╝  ╚════██║   ██║   ██╔══██║██║╚██╗██║   ██║"
echo "██║     ╚██████╔╝██║  ██║███████╗███████║   ██║   ██║  ██║██║ ╚████║   ██║"
echo "╚═╝      ╚═════╝ ╚═╝  ╚═╝╚══════╝╚══════╝   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═══╝   ╚═╝   v3.0"
echo -e "${NC}"

# Root kontrolü
if [[ $EUID -ne 0 ]]; then
   echo -e "${RED}[HATA] Bu script root olarak çalışmalıdır: sudo ./deploy.sh${NC}"
   exit 1
fi

# binary dosyası var mı?
if [ ! -f "./${BINARY_NAME}" ]; then
    echo -e "${RED}[HATA] '${BINARY_NAME}' dosyası bulunamadı!${NC}"
    echo "Windows'ta önce derleyin:"
    echo "  \$env:GOOS='linux'; \$env:GOARCH='amd64'; go build -ldflags='-s -w' -o forestant-engine ."
    exit 1
fi

# ─── 1. Sistem Kullanıcısı Oluştur ───────────────────────────
echo -e "\n${YELLOW}[1/6] Sistem kullanıcısı oluşturuluyor...${NC}"
if ! id "${SERVICE_USER}" &>/dev/null; then
    useradd --system --no-create-home --shell /bin/false "${SERVICE_USER}"
    echo -e "${GREEN}✓ '${SERVICE_USER}' kullanıcısı oluşturuldu${NC}"
else
    echo -e "${GREEN}✓ '${SERVICE_USER}' kullanıcısı zaten mevcut${NC}"
fi

# ─── 2. Dizin Yapısı ─────────────────────────────────────────
echo -e "\n${YELLOW}[2/6] Dizin yapısı oluşturuluyor...${NC}"
mkdir -p "${INSTALL_DIR}/data"
mkdir -p "${INSTALL_DIR}/dist"
echo -e "${GREEN}✓ ${INSTALL_DIR} hazır${NC}"

# ─── 3. Dosyaları Kopyala ────────────────────────────────────
echo -e "\n${YELLOW}[3/6] Dosyalar kopyalanıyor...${NC}"

# Binary
cp "./${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
echo -e "${GREEN}✓ Binary kopyalandı${NC}"

# Frontend dist
if [ -d "./dist" ]; then
    cp -r ./dist/* "${INSTALL_DIR}/dist/"
    echo -e "${GREEN}✓ Frontend kopyalandı${NC}"
else
    echo -e "${YELLOW}⚠ 'dist' klasörü bulunamadı — frontend olmadan devam ediliyor${NC}"
fi

# .env dosyası
if [ -f "./.env" ]; then
    cp "./.env" "${INSTALL_DIR}/.env"
    chmod 600 "${INSTALL_DIR}/.env"
    echo -e "${GREEN}✓ .env kopyalandı${NC}"
elif [ -f "./.env.example" ]; then
    cp "./.env.example" "${INSTALL_DIR}/.env"
    chmod 600 "${INSTALL_DIR}/.env"
    echo -e "${YELLOW}⚠ .env.example'dan .env oluşturuldu — Lütfen ayarları düzenleyin!${NC}"
else
    echo -e "${YELLOW}⚠ .env dosyası bulunamadı — varsayılan ayarlar kullanılacak${NC}"
fi

# ─── 4. İzinler ──────────────────────────────────────────────
echo -e "\n${YELLOW}[4/6] İzinler ayarlanıyor...${NC}"
chown -R "${SERVICE_USER}:${SERVICE_USER}" "${INSTALL_DIR}"

# Port 80 için setcap
PORT=$(grep -E "^PORT=" "${INSTALL_DIR}/.env" 2>/dev/null | cut -d= -f2 | tr -d ' \r')
PORT=${PORT:-80}

if [ "$PORT" -lt 1024 ] 2>/dev/null; then
    if command -v setcap &>/dev/null; then
        setcap 'cap_net_bind_service=+ep' "${INSTALL_DIR}/${BINARY_NAME}"
        echo -e "${GREEN}✓ Port ${PORT} için setcap uygulandı${NC}"
    else
        echo -e "${YELLOW}⚠ setcap bulunamadı — libcap2-bin yükleyin: apt install libcap2-bin${NC}"
    fi
fi

# ─── 5. Systemd Service ──────────────────────────────────────
echo -e "\n${YELLOW}[5/6] Systemd service kuruluyor...${NC}"
cp "./forestant.service" "/etc/systemd/system/${SERVICE_NAME}.service"

systemctl daemon-reload
systemctl enable "${SERVICE_NAME}"
systemctl restart "${SERVICE_NAME}"

sleep 2
if systemctl is-active --quiet "${SERVICE_NAME}"; then
    echo -e "${GREEN}✓ Servis başarıyla başlatıldı${NC}"
else
    echo -e "${RED}[HATA] Servis başlatılamadı. Log: journalctl -u ${SERVICE_NAME} -n 30${NC}"
    exit 1
fi

# ─── 6. Özet ─────────────────────────────────────────────────
echo -e "\n${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}  ✅ ForestANT v3.0 başarıyla kuruldu!${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "  🌐 Adres   : ${BLUE}http://$(hostname -I | awk '{print $1}'):${PORT}${NC}"
echo -e "  📁 Dizin   : ${INSTALL_DIR}"
echo -e "  📋 Loglar  : journalctl -u ${SERVICE_NAME} -f"
echo -e "  🔄 Restart : systemctl restart ${SERVICE_NAME}"
echo -e "  ⚙️  Ayarlar : nano ${INSTALL_DIR}/.env"
echo ""
