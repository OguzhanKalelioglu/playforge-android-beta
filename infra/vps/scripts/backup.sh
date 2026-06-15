#!/bin/bash
# PostgreSQL Yedekleme Scripti
# VPS'te cron ile çalışır (her gece 03:00)
# Backblaze B2'ye yükler

set -e

BACKUP_DIR="/backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/testers_$TIMESTAMP.sql.gz"
RETENTION_DAYS=30

mkdir -p "$BACKUP_DIR"

echo "[$(date)] Backup başlıyor..."

# pg_dump
docker exec testers-vps-postgres-1 pg_dump -U tester testers | gzip > "$BACKUP_FILE"

if [ ! -s "$BACKUP_FILE" ]; then
    echo "HATA: Backup dosyası boş!"
    exit 1
fi

echo "Backup oluşturuldu: $BACKUP_FILE ($(du -h "$BACKUP_FILE" | cut -f1))"

# B2'ye yükle (rclone gerekli)
if command -v rclone &>/dev/null; then
    rclone copy "$BACKUP_FILE" "b2:testers-backups/" 2>&1
    echo "B2'ye yüklendi"
else
    echo "UYARI: rclone kurulu değil, B2 upload atlandı"
fi

# Eski backupları sil
find "$BACKUP_DIR" -name "testers_*.sql.gz" -mtime +$RETENTION_DAYS -delete
echo "Eski backuplar temizlendi (>$RETENTION_DAYS gün)"

echo "[$(date)] Backup tamamlandı."
