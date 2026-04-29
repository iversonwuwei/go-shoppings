ALTER TABLE "after_sale_orders"
    ADD COLUMN IF NOT EXISTS "return_express_code" VARCHAR(30) DEFAULT '';
