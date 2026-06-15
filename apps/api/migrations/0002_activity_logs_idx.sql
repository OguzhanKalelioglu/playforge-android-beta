-- 0002 — Activity logs indexes (Hafta 5 task runner için)
-- orchestrator'dan gelecek step event'leri için hızlı sorgu

-- test_id + zaman sırası (müşteri dashboard timeline)
CREATE INDEX IF NOT EXISTS idx_activity_logs_test_time
    ON activity_logs(test_id, performed_at DESC)
    WHERE test_id IS NOT NULL;

-- step bazlı hata analizi
CREATE INDEX IF NOT EXISTS idx_activity_logs_action_error
    ON activity_logs(action, performed_at DESC)
    WHERE success = FALSE;

-- assignment bazlı hızlı erişim
CREATE INDEX IF NOT EXISTS idx_activity_logs_assignment_action
    ON activity_logs(test_assignment_id, action, performed_at DESC);

-- metadata JSONB içinde hızlı arama (gin extension gerekir)
-- Örnek: metadata'da 'gesture_count' alanına göre filtreleme
CREATE INDEX IF NOT EXISTS idx_activity_logs_metadata_gin
    ON activity_logs USING GIN (metadata)
    WHERE metadata != '{}'::jsonb;
