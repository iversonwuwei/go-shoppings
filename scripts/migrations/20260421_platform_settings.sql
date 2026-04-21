-- 平台全局设置（单行，id=1）
CREATE TABLE IF NOT EXISTS platform_settings (
    id                  BIGSERIAL PRIMARY KEY,
    platform_name       TEXT DEFAULT '',
    platform_logo       TEXT DEFAULT '',
    support_phone       TEXT DEFAULT '',
    support_email       TEXT DEFAULT '',
    wxpay_app_id        TEXT DEFAULT '',
    wxpay_mch_id        TEXT DEFAULT '',
    wxpay_apiv3_key     TEXT DEFAULT '',
    wxpay_cert_serial   TEXT DEFAULT '',
    wxpay_notify_url    TEXT DEFAULT '',
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO platform_settings (id) VALUES (1) ON CONFLICT (id) DO NOTHING;
