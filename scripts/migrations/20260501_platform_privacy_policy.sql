-- 平台隐私协议字段（平台全局单行配置）
ALTER TABLE platform_settings
    ADD COLUMN IF NOT EXISTS privacy_policy_title TEXT DEFAULT '',
    ADD COLUMN IF NOT EXISTS privacy_policy_effective_date TEXT DEFAULT '',
    ADD COLUMN IF NOT EXISTS privacy_policy_content TEXT DEFAULT '',
    ADD COLUMN IF NOT EXISTS privacy_policy_contact_phone TEXT DEFAULT '',
    ADD COLUMN IF NOT EXISTS privacy_policy_contact_email TEXT DEFAULT '';
