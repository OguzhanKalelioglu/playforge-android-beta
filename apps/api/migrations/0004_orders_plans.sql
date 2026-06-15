-- 0004 — Orders, Plans, Sessions, Refresh Tokens
-- Plan tier modeli (basic/pro/enterprise), ayrı order+items, JWT session

-- ============================================
-- Plan Tiers (ürün paketleri)
-- ============================================
CREATE TABLE plan_tiers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(50) UNIQUE NOT NULL,    -- basic, pro, enterprise
    name VARCHAR(100) NOT NULL,
    description TEXT,
    tester_count INT NOT NULL,
    duration_days INT NOT NULL DEFAULT 14,
    price_try DECIMAL(10, 2) NOT NULL,
    price_usd DECIMAL(10, 2),
    features JSONB NOT NULL DEFAULT '[]'::jsonb,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_plans_active_sort ON plan_tiers(is_active, sort_order) WHERE is_active = TRUE;

-- Default plan tier'lar (başlangıç verisi)
INSERT INTO plan_tiers (slug, name, description, tester_count, duration_days, price_try, features, sort_order) VALUES
('basic', 'Basic', '25 tester, 14 günlük Google Play closed test', 25, 14, 4999.00,
 '["25 Google Play tester hesabı", "14 günlük otomatik engagement", "5 yıldız yorum (mix)", "Standart raporlama", "E-posta destek"]'::jsonb, 1),
('pro', 'Pro', '25 tester, 14 gün + haftalık detaylı rapor + öncelikli destek', 25, 14, 7999.00,
 '["Basic plan tüm özellikler", "Haftalık detaylı ilerleme raporu", "Öncelikli destek", "Screenshots timeline", "Custom yorum tonu"]'::jsonb, 2),
('enterprise', 'Enterprise', '25 tester + ban koruma premium + dedicated account manager', 25, 14, 12999.00,
 '["Pro plan tüm özellikler", "Premium ban koruma (rotasyonlu IP)", "Dedicated account manager", "SLA — 24 saat yanıt", "API erişimi", "Özel scheduled test"]'::jsonb, 3);

-- ============================================
-- Orders (checkout flow, payment öncesi)
-- ============================================
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    plan_tier_id UUID NOT NULL REFERENCES plan_tiers(id) ON DELETE RESTRICT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'awaiting_payment', 'paid', 'failed', 'cancelled', 'refunded')),
    subtotal DECIMAL(10, 2) NOT NULL,
    tax_total DECIMAL(10, 2) NOT NULL DEFAULT 0,
    total DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'TRY',
    test_id UUID REFERENCES tests(id),    -- paid olduktan sonra bağlanır
    iyzico_checkout_form_token VARCHAR(500),
    iyzico_token_expires_at TIMESTAMPTZ,
    billing_email VARCHAR(255),
    billing_name VARCHAR(255),
    billing_phone VARCHAR(50),
    billing_address TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '30 minutes'),
    paid_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_orders_user ON orders(user_id, created_at DESC);
CREATE INDEX idx_orders_status_expires ON orders(status, expires_at) WHERE status IN ('pending', 'awaiting_payment');
CREATE INDEX idx_orders_checkout_token ON orders(iyzico_checkout_form_token) WHERE iyzico_checkout_form_token IS NOT NULL;

CREATE TRIGGER update_orders_updated_at
    BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Sessions (refresh token storage)
-- ============================================
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash VARCHAR(255) UNIQUE NOT NULL,
    user_agent TEXT,
    ip_address INET,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_sessions_user ON sessions(user_id, expires_at DESC);
CREATE INDEX idx_sessions_expires ON sessions(expires_at) WHERE revoked_at IS NULL;
CREATE INDEX idx_sessions_token ON sessions(refresh_token_hash);

-- ============================================
-- Password reset tokens
-- ============================================
CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_reset_user ON password_reset_tokens(user_id, expires_at DESC);

-- ============================================
-- Email verification tokens
-- ============================================
CREATE TABLE email_verification_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================
-- Tester daily activity (rotasyon tracking)
-- ============================================
CREATE TABLE tester_daily_usage (
    id BIGSERIAL PRIMARY KEY,
    tester_id UUID NOT NULL REFERENCES testers(id) ON DELETE CASCADE,
    usage_date DATE NOT NULL,
    tasks_completed INT NOT NULL DEFAULT 0,
    minutes_active INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tester_id, usage_date)
);
CREATE INDEX idx_tester_usage_date ON tester_daily_usage(usage_date DESC);

-- ============================================
-- Plan tier'lar için VIEW
-- ============================================
CREATE OR REPLACE VIEW plan_tiers_public AS
SELECT
    id, slug, name, description, tester_count, duration_days,
    price_try, price_usd, features, sort_order
FROM plan_tiers
WHERE is_active = TRUE;
