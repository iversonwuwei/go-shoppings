# Backend GitHub Actions CI/CD

后端使用 GitHub Actions 完成 CI/CD，工作流文件为 `.github/workflows/deploy-backend.yml`。流水线保持“先验证、后部署”的顺序：PR 和推送都会执行后端验证，只有推送到 `master` 或手动触发时才会进入部署阶段。

## Pipeline stages

| Stage | Trigger | Gate | Description |
| --- | --- | --- | --- |
| Verify backend | Pull request, push to `master`, manual run | Required | 拉取依赖、检查 `gofmt`、执行 `go vet`、运行 `go test -race -cover ./...`、编译 API/Worker 二进制、校验 Docker Compose、构建 API 与 AI 图片服务镜像。 |
| Deploy Docker services | Push to `master`, manual run | Depends on verify | 通过 SSH 更新服务器代码，按需应用 SQL 迁移，重建并启动 Redis、API、AI 图片服务，最后检查 `/healthz`。 |

后端相关文件变更才会触发流水线，包括 Go 代码、配置、Dockerfile、Compose 文件、迁移脚本和本 workflow。未通过验证时不会部署。部署并发按分支串行执行，避免同一目标环境被多个 workflow 同时更新。

## Local verification

本地提交前建议执行与 CI 近似的检查：

```powershell
gofmt -w .
go vet ./...
go test -race -cover ./...
go build -ldflags="-s -w" -o bin/api ./cmd/api
go build -ldflags="-s -w" -o bin/worker ./cmd/worker
docker compose -f docker-compose.infra.yml config --quiet
docker compose -f docker-compose.app.yml config --quiet
docker build -t wechat-mall-api:local -f Dockerfile .
docker build -t wechat-mall-ai-image:local -f ai-image-service/Dockerfile ai-image-service
```

## Production deployment path

本地和 GitHub Actions 都使用同一条部署路径：应用 SQL 迁移，先启动基础服务 `docker-compose.infra.yml` 中的 Redis，再启动应用服务 `docker-compose.app.yml` 中的 API 和 AI 图片服务，最后检查健康状态。

本地启动验证：

```powershell
docker compose -f docker-compose.infra.yml up -d redis
docker compose -f docker-compose.app.yml up -d --build api ai-image
curl.exe -fsS http://127.0.0.1:18080/healthz
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

## Rollback

如果健康检查失败，workflow 会停止并保留远端 Docker 日志用于排查。需要回滚时，将 `master` 回退到上一可用提交后重新运行 workflow，或在服务器 `DEPLOY_PATH` 中手动检出上一可用提交并执行：

```powershell
docker compose -f docker-compose.infra.yml up -d redis
docker compose -f docker-compose.app.yml up -d --build api ai-image
curl.exe -fsS http://127.0.0.1:18080/healthz
curl.exe -fsS http://127.0.0.1:8090/healthz
```

推送到 `master` 的后端相关文件变更会自动部署，也可以在 Actions 页面手动触发 `Backend CI/CD`。
