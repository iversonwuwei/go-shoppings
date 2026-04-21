-- 20260421_tenant_billing.sql
-- 为 tenants 表补充 billing_cycle 字段，记录租户申请入驻时选择的计费周期（monthly/yearly）。

ALTER TABLE "tenants" ADD COLUMN IF NOT EXISTS "billing_cycle" VARCHAR(10) NOT NULL DEFAULT 'yearly';
