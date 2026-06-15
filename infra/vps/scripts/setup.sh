#!/bin/bash
# VPS İlk Kurulum Scripti
# Hetzner CPX41 (16GB RAM, Ubuntu 22.04) için

set -e

echo "=== TestersCommunity VPS Setup ==="
echo "Bu script Ubuntu 22.04'te çalışır."

# Root kontrolü
if [ "$EUID" -ne 0 ]; then
    echo "Lütfen root olarak çalıştırın: sudo $0"
    exit 1
fi

# 1. Sistem güncelleme
echo "[1/8] Sistem güncelleniyor..."
apt update && apt upgrade -y

# 2. Kullanıcı oluşturma
if ! id "testops" &>/dev/null; then
    echo "[2/8] testops kullanıcısı oluşturuluyor..."
    adduser --disabled-password --gecos "" testops
    usermod -aG sudo testops
    mkdir -p /home/testops/.ssh
    cp ~/.ssh/authorized_keys /home/testops/.ssh/ 2>/dev/null || true
    chown -R testops:testops /home/testops/.ssh
    echo "testops ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/testops
fi

# 3. Firewall
echo "[3/8] Firewall yapılandırılıyor..."
apt install -y ufw
ufw --force reset
ufw default deny incoming
ufw default allow outgoing
ufw allow OpenSSH
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

# 4. Fail2ban
echo "[4/8] Fail2ban kuruluyor..."
apt install -y fail2ban
cat > /etc/fail2ban/jail.local << 'EOF'
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 5

[sshd]
enabled = true
port = ssh
filter = sshd
logpath = /var/log/auth.log
EOF
systemctl enable fail2ban
systemctl restart fail2ban

# 5. Otomatik güvenlik güncellemeleri
echo "[5/8] Unattended-upgrades kuruluyor..."
apt install -y unattended-upgrades apt-listchanges
dpkg-reconfigure -f noninteractive unattended-upgrades

# 6. Docker
echo "[6/8] Docker kuruluyor..."
if ! command -v docker &>/dev/null; then
    curl -fsSL https://get.docker.com -o /tmp/get-docker.sh
    sh /tmp/get-docker.sh
    usermod -aG docker testops
fi

# 7. Timezone
echo "[7/8] Timezone ayarlanıyor..."
timedatectl set-timezone Europe/Istanbul

# 8. Swap (16GB RAM yeterli ama küçük swap ekleyelim)
echo "[8/8] Swap kontrol ediliyor..."
if [ ! -f /swapfile ]; then
    fallocate -l 2G /swapfile
    chmod 600 /swapfile
    mkswap /swapfile
    swapon /swapfile
    echo '/swapfile none swap sw 0 0' >> /etc/fstab
fi

echo ""
echo "=== Kurulum tamamlandı! ==="
echo ""
echo "Sonraki adımlar:"
echo "  1. testops olarak giriş yapın: ssh testops@<IP>"
echo "  2. Repo'yu klonlayın"
echo "  3. .env dosyasını oluşturun: cp .env.example .env"
echo "  4. Servisleri başlatın: cd infra/vps && docker compose up -d"
echo ""
