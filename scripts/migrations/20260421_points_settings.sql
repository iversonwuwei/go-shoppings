-- points_settings: 租户的积分规则（每租户一行）
-- earn_rate: 下单实付金额每 1 元获得的积分数（INT 部分参与发放）
-- min_amount: 订单金额达到多少才开始发放
-- redeem_rate: 1 元对应的积分数（redeem 用）
CREATE TABLE IF NOT EXISTS "points_settings" (
    "tenant_id"  BIGINT PRIMARY KEY REFERENCES "tenants"("id"),
    "enabled"    SMALLINT NOT NULL DEFAULT 1,
    "earn_rate"  NUMERIC(10,4) NOT NULL DEFAULT 1,
    "min_amount" NUMERIC(10,2) NOT NULL DEFAULT 0,
    "redeem_rate" INT NOT NULL DEFAULT 100,
    "remark"     VARCHAR(500) NOT NULL DEFAULT '',
    "updated_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
