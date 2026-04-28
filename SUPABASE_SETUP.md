# Supabase 切换说明

后端现在默认使用 Supabase Postgres 和 Supabase Storage。敏感配置通过 `.env` 或系统环境变量读取，不需要写进 `configs/config.yaml`。

## 1. 配置环境变量

复制 `.env.example` 为 `.env`，填写以下值：

- `SUPABASE_DB_DSN`: Supabase Postgres 连接串，建议使用 Dashboard 中的 Transaction pooler 或 Direct connection，并保留 `sslmode=require`。
- `SUPABASE_URL`: 形如 `https://<project-ref>.supabase.co`。
- `SUPABASE_SERVICE_ROLE_KEY`: Supabase service role key，仅服务端使用。
- `SUPABASE_STORAGE_BUCKET`: 默认 `wxmall`。
- `SUPABASE_STORAGE_PUBLIC_READ`: 商品和分类图片建议为 `true`，这样小程序可以直接展示公开 URL。

## 2. 初始化数据库

如果是全新的 Supabase 项目，可以在 Supabase SQL Editor 中依次执行：

1. `scripts/init_db.sql`
2. `scripts/seed.sql`

注意：`scripts/init_db.sql` 会先 drop 现有业务表，只适合全新或可重置的数据库。已有数据的 Supabase 项目应改用迁移脚本。

## 3. 初始化 Storage

保持 `SUPABASE_STORAGE_CREATE_BUCKET=true` 时，应用启动会用 service role key 检查并创建 bucket。也可以在 Supabase Dashboard 手动创建 `wxmall` bucket，并按需设置为 public。

## 4. 本地启动

本地仍需要 Redis：

```bash
docker compose -f docker-compose.infra.yml up -d redis
go run ./cmd/api/main.go
```

Docker Compose 运行 API 时也会读取同一份 `.env`，应用服务定义在 `docker-compose.app.yml`。

本地 Postgres 和 MinIO 仍保留在 `local` profile 中，仅做回退使用：

```bash
docker compose -f docker-compose.infra.yml --profile local up -d postgres minio
```
