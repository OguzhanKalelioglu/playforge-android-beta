-- ============================================
-- TestersCommunity - Initial Schema
-- ============================================

-- Extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- Users (müşteri + admin)
-- ============================================
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'customer' CHECK (role IN ('customer', 'admin')),
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_users_email ON users(email);

-- ============================================
-- Google Groups (her test için ayrı grup)
-- ============================================
CREATE TABLE google_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'archived'))
);
CREATE INDEX idx_groups_status ON google_groups(status);

-- ============================================
-- Device profiles (her hesap için cihaz kimliği)
-- ============================================
CREATE TABLE device_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    android_id VARCHAR(50) NOT NULL,
    imei VARCHAR(20) NOT NULL,
    mac_address VARCHAR(20) NOT NULL,
    model VARCHAR(50) NOT NULL,
    manufacturer VARCHAR(50) NOT NULL,
    android_version VARCHAR(10) NOT NULL,
    screen_resolution VARCHAR(20) NOT NULL,
    user_agent TEXT NOT NULL,
    locale VARCHAR(10) NOT NULL,
    timezone VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================
-- Testers (25 Google hesabı)
-- ============================================
CREATE TABLE testers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_encrypted BYTEA NOT NULL,
    recovery_email VARCHAR(255),
    phone VARCHAR(50),
    google_group_id UUID REFERENCES google_groups(id),
    device_profile_id UUID REFERENCES device_profiles(id),
    status VARCHAR(20) NOT NULL DEFAULT 'warming'
        CHECK (status IN ('warming', 'active', 'cooling', 'disabled')),
    notes TEXT,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_testers_status ON testers(status);
CREATE INDEX idx_testers_group ON testers(google_group_id);

-- ============================================
-- Tests (müşteri test işleri)
-- ============================================
CREATE TABLE tests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    package_name VARCHAR(255) NOT NULL,
    test_link TEXT,
    notes TEXT,
    star_preference VARCHAR(20) NOT NULL DEFAULT 'mixed'
        CHECK (star_preference IN ('all5', 'mixed', 'custom')),
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'active', 'completed', 'failed', 'cancelled')),
    google_group_id UUID REFERENCES google_groups(id),
    started_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_tests_user ON tests(user_id, created_at DESC);
CREATE INDEX idx_tests_status_ends ON tests(status, ends_at);
CREATE INDEX idx_tests_package ON tests(package_name);

-- ============================================
-- Test assignments (hangi tester hangi teste atandı)
-- ============================================
CREATE TABLE test_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    test_id UUID NOT NULL REFERENCES tests(id) ON DELETE CASCADE,
    tester_id UUID NOT NULL REFERENCES testers(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'in_progress', 'completed', 'failed', 'skipped')),
    opt_in_at TIMESTAMPTZ,
    install_at TIMESTAMPTZ,
    last_engagement_at TIMESTAMPTZ,
    review_id UUID,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(test_id, tester_id)
);
CREATE INDEX idx_assignments_tester ON test_assignments(tester_id, status);
CREATE INDEX idx_assignments_test ON test_assignments(test_id, status);

-- ============================================
-- Reviews
-- ============================================
CREATE TABLE reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    test_assignment_id UUID NOT NULL REFERENCES test_assignments(id) ON DELETE CASCADE,
    rating INT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    review_text TEXT NOT NULL,
    language VARCHAR(10) NOT NULL DEFAULT 'tr',
    posted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_reviews_assignment ON reviews(test_assignment_id);

-- FK from assignments to reviews (created after reviews)
ALTER TABLE test_assignments
    ADD CONSTRAINT fk_assignments_review
    FOREIGN KEY (review_id) REFERENCES reviews(id);

-- ============================================
-- Activity logs (günlük aktivite kayıtları)
-- ============================================
CREATE TABLE activity_logs (
    id BIGSERIAL PRIMARY KEY,
    test_assignment_id UUID NOT NULL REFERENCES test_assignments(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL
        CHECK (action IN ('opt_in', 'download', 'install', 'open', 'interact', 'review', 'error')),
    performed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    success BOOLEAN NOT NULL DEFAULT TRUE,
    error_message TEXT,
    screenshot_path TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb
);
CREATE INDEX idx_logs_assignment ON activity_logs(test_assignment_id, performed_at DESC);
CREATE INDEX idx_logs_action_time ON activity_logs(action, performed_at DESC);

-- ============================================
-- Payments
-- ============================================
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    test_id UUID REFERENCES tests(id),
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'TRY',
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'completed', 'refunded', 'failed', 'cancelled')),
    iyzico_token VARCHAR(255),
    iyzico_payment_id VARCHAR(255),
    paid_at TIMESTAMPTZ,
    refunded_at TIMESTAMPTZ,
    refund_amount DECIMAL(10, 2),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_payments_user ON payments(user_id, created_at DESC);
CREATE INDEX idx_payments_status ON payments(status, created_at DESC);
CREATE INDEX idx_payments_test ON payments(test_id);

-- ============================================
-- System events (operasyonel loglar)
-- ============================================
CREATE TABLE system_events (
    id BIGSERIAL PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL
        CHECK (severity IN ('info', 'warning', 'error', 'critical')),
    message TEXT NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_events_type_time ON system_events(event_type, created_at DESC);
CREATE INDEX idx_events_severity_time ON system_events(severity, created_at DESC)
    WHERE severity IN ('error', 'critical');

-- ============================================
-- Triggers (updated_at otomatik güncelleme)
-- ============================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tests_updated_at
    BEFORE UPDATE ON tests
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_assignments_updated_at
    BEFORE UPDATE ON test_assignments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_payments_updated_at
    BEFORE UPDATE ON payments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
