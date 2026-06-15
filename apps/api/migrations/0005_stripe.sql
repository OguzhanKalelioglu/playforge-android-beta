-- 0005 — Stripe ödeme alanları
-- Iyzico kaldırıldı (bkz. docs/vault/decisions/0001-stripe-over-iyzico.md)
-- Stripe Checkout Session + PaymentIntent + Charge ID'ler

-- ============================================
-- Orders: stripe alanları
-- ============================================
ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS stripe_checkout_session_id VARCHAR(255),
    ADD COLUMN IF NOT EXISTS stripe_session_expires_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS stripe_payment_intent_id VARCHAR(255),
    ADD COLUMN IF NOT EXISTS stripe_customer_id VARCHAR(255);

CREATE INDEX IF NOT EXISTS idx_orders_stripe_session
    ON orders(stripe_checkout_session_id)
    WHERE stripe_checkout_session_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_orders_stripe_payment_intent
    ON orders(stripe_payment_intent_id)
    WHERE stripe_payment_intent_id IS NOT NULL;

-- ============================================
-- Payments: stripe alanları (iyzico kaldır)
-- ============================================
ALTER TABLE payments
    RENAME COLUMN iyzico_token TO stripe_session_id;
ALTER TABLE payments
    RENAME COLUMN iyzico_payment_id TO stripe_payment_id;
ALTER TABLE payments
    ADD COLUMN IF NOT EXISTS stripe_charge_id VARCHAR(255);

-- Eski iyzico_payment_id kalmışsa tip dönüşümü (VARCHAR(255) zaten uyumlu)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_name = 'payments' AND column_name = 'iyzico_payment_id') THEN
        ALTER TABLE payments DROP COLUMN iyzico_payment_id;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_name = 'payments' AND column_name = 'iyzico_token') THEN
        -- Eğer rename yapılmadıysa (DB 0001 + 0004'ü atladıysa)
        ALTER TABLE payments DROP COLUMN iyzico_token;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_payments_stripe_session
    ON payments(stripe_session_id)
    WHERE stripe_session_id IS NOT NULL;
