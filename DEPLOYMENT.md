# Backend Docker Deployment

本地和 GitHub Actions 都使用同一条部署路径：应用 SQL 迁移，先启动基础服务 `docker-compose.infra.yml` 中的 Redis，再启动应用服务 `docker-compose.app.yml` 中的 API 和 AI 图片服务，最后检查健康状态。

## Local verification

```powershell
docker compose -f docker-compose.infra.yml up -d redis
docker compose -f docker-compose.app.yml up -d --build api ai-image
curl.exe -fsS http://127.0.0.1:8080/healthz
curl.exe -fsS http://127.0.0.1:8090/healthz
```

如需把 `scripts/migrations` 应用到 Supabase，可使用项目根目录的 `.env`：

```powershell
docker run --rm --env-file .env -v "${PWD}\scripts\migrations:/migrations:ro" postgres:15-alpine sh -c 'for file in /migrations/*.sql; do [ -e "$file" ] || continue; psql "$SUPABASE_DB_DSN" -v ON_ERROR_STOP=1 -f "$file"; done'
```

## GitHub Actions secrets

在仓库 `Settings -> Secrets and variables -> Actions` 中配置：

| Secret | Required | Description |
| --- | --- | --- |
| `DEPLOY_HOST` | Yes | 服务器地址。 |
| `DEPLOY_USER` | Yes | SSH 登录用户。 |
| `DEPLOY_SSH_KEY` | Yes | 可登录服务器的私钥。 |
| `DEPLOY_PORT` | No | SSH 端口，默认 `22`。 |
| `DEPLOY_PATH` | No | 服务器上的项目目录，默认 `/srv/go-shoppings`。 |
| `DEPLOY_BRANCH` | No | 部署分支，默认 `master`。 |
| `DEPLOY_ENV_FILE` | No | 服务器 `.env` 的完整内容；为空时使用服务器已有 `.env`。 |
| `APPLY_DATABASE_MIGRATIONS` | No | 设为 `false` 可跳过迁移，默认执行。 |

服务器需要提前安装 `git`、Docker 和 Docker Compose 插件，并允许部署用户运行 Docker。`.env` 至少需要包含 Supabase 数据库、Supabase Storage、JWT 和 Minimax 相关配置。

推送到 `master` 的后端相关文件变更会自动部署，也可以在 Actions 页面手动触发 `Deploy backend`。
