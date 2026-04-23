CREATE TABLE IF NOT EXISTS "order_messages" (
    "id"              BIGSERIAL PRIMARY KEY,
    "tenant_id"       BIGINT NOT NULL REFERENCES "tenants"("id"),
    "order_id"        BIGINT NOT NULL REFERENCES "orders"("id"),
    "order_no"        VARCHAR(32) NOT NULL,
    "event_type"      VARCHAR(40) NOT NULL,
    "title"           VARCHAR(120) NOT NULL,
    "content"         VARCHAR(500) NOT NULL,
    "status"          VARCHAR(20) NOT NULL DEFAULT 'unread',
    "read_at"         TIMESTAMP,
    "created_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS "idx_order_messages_tenant" ON "order_messages" ("tenant_id");
CREATE INDEX IF NOT EXISTS "idx_order_messages_order" ON "order_messages" ("order_id");
CREATE INDEX IF NOT EXISTS "idx_order_messages_status" ON "order_messages" ("status");
