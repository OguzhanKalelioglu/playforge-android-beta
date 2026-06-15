#!/bin/bash
# Mini PC İlk Kurulum Scripti
# 64GB RAM Ubuntu 22.04 Server (headless)

set -e

echo "=== TestersCommunity Mini PC Setup ==="
echo "Bu script Ubuntu 22.04 Server'da çalışır (headless)."

if [ "$EUID" -ne 0 ]; then
    echo "Lütfen root olarak çalıştırın: sudo $0"
    exit 1
fi

# 1. Sistem güncelleme
echo "[1/6] Sistem güncelleniyor..."
apt update && apt upgrade -y

# 2. Temel paketler
echo "[2/6] Temel paketler kuruluyor..."
apt install -y curl wget git ca-certificates gnupg lsb-release \
    software-properties-common apt-transport-packages

# 3. Docker
echo "[3/6] Docker kuruluyor..."
if ! command -v docker &>/dev/null; then
    curl -fsSL https://get.docker.com -o /tmp/get-docker.sh
    sh /tmp/get-docker.sh
    usermod -aG docker testops
fi

# Docker Compose v2
apt install -y docker-compose-plugin

# 4. KVM kontrolü
echo "[4/6] KVM kontrolü..."
if [ ! -e /dev/kvm ]; then
    echo "UYARI: /dev/kvm bulunamadı!"
    echo "BIOS'a girip Intel VT-x veya AMD SVM'yi açın."
    exit 1
fi
chmod 666 /dev/kvm

# 5. Timezone
echo "[5/6] Timezone..."
timedatectl set-timezone Europe/Istanbul

# 6. Android SDK (Appium için)
echo "[6/6] Android SDK kurulumu..."
if [ ! -d /opt/android-sdk ]; then
    mkdir -p /opt/android-sdk
    cd /opt/android-sdk
    wget -q https://dl.google.com/android/repository/commandlinetools-linux-11076708_latest.zip -O cmdline.zip
    unzip -q cmdline.zip
    mkdir -p cmdline-tools/latest
    mv cmdline-tools/bin cmdline-tools/lib cmdline-tools/NOTICE.txt cmdline-tools/source.properties cmdline-tools/latest/ 2>/dev/null || true
    echo "Android SDK kuruldu: /opt/android-sdk"
fi

echo ""
echo "=== Kurulum tamamlandı! ==="
echo ""
echo "Test:"
echo "  - Docker: docker run hello-world"
echo "  - KVM: ls -la /dev/kvm"
echo ""
echo "Sonraki adımlar:"
echo "  1. Repo'yu klonlayın: git clone <repo> ~/app"
echo "  2. Emulator'ü başlatın: cd infra/minipc && docker compose up -d"
echo "  3. ADB ile bağlanın: adb connect localhost:5554"
echo ""
