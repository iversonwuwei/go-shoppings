# Backend GitHub Actions CI/CD

后端使用 GitHub Actions 完成 CI/CD，工作流文件为 `.github/workflows/deploy-backend.yml`。流水线保持“先验证、后发布镜像、后部署”的顺序：PR 和推送都会执行后端验证，只有推送到 `master` 或手动触发时才会构建生产镜像、推送到华为云 SWR，并进入服务器部署阶段。

## Pipeline stages

| Stage | Trigger | Gate | Description |
| --- | --- | --- | --- |
| Verify backend | Pull request, push to `master`, manual run | Required | 拉取依赖、检查 `gofmt`、执行 `go vet`、运行 `go test -race -cover ./...`、编译 API/Worker 二进制、校验 Docker Compose、构建 API 与 AI 图片服务镜像。 |
| Publish SWR images | Push to `master`, manual run | Depends on verify | 登录华为云 SWR，构建 API 与 AI 图片服务生产镜像，推送 `${GITHUB_SHA}` 与 `latest` 标签，并将 Gateway/Redis 运行时基础镜像同步到 SWR。 |
| Deploy Docker services | Push to `master`, manual run | Depends on publish | 通过 SSH 同步最小运行时部署工件，登录 SWR，拉取本次构建的镜像并启动 Redis、Gateway、API、AI 图片服务，最后检查 Gateway 暴露的健康接口。 |

后端相关文件变更才会触发流水线，包括 Go 代码、配置、Dockerfile、Compose 文件和本 workflow。数据库迁移脚本变更不会触发部署，迁移需要按手动流程单独执行。未通过验证时不会部署。部署并发按分支串行执行，避免同一目标环境被多个 workflow 同时更新。

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
API_IMAGE=wechat-mall-api:test AI_IMAGE=wechat-mall-ai-image:test docker compose -f docker-compose.deploy.yml config --quiet
docker build -t wechat-mall-api:local -f Dockerfile .
docker build -t wechat-mall-ai-image:local -f ai-image-service/Dockerfile ai-image-service
```

## Production deployment path

生产服务器不保存后端源码，也不在服务器本机构建后端镜像。GitHub Actions 在 `publish` 阶段将 API、AI 图片服务、Gateway 和 Redis 镜像推送或同步到华为云 SWR；服务器部署阶段只保存运行时部署工件并执行 `docker pull` 与 `docker compose up --no-build`，确保生产不依赖 Docker Hub。GitHub Actions 不执行数据库迁移，也不把 `scripts/migrations` 同步到服务器。

服务器 `DEPLOY_PATH` 只会保留这些运行时文件和 Docker 数据：

```powershell
.env
.deploy-images.env
docker-compose.infra.yml
docker-compose.deploy.yml
gateway/nginx.conf
```

每次同步运行时部署工件前，workflow 会清理 `DEPLOY_PATH` 中除 `.env` 和 `.deploy-images.env` 以外的旧文件，避免服务器残留源码 checkout。

部署流程会启动基础服务 `docker-compose.infra.yml` 中的 Redis，再启动应用服务 `docker-compose.deploy.yml` 中的 Gateway、API 和 AI 图片服务。外部入口统一由 Gateway 暴露在 `API_HOST_PORT`（默认 `18080`），API 与 AI 图片服务只在 Docker 网络内互通；AI 生成仍通过 Go API 鉴权入口调用，Gateway 只公开 AI 健康检查。

生产镜像变量由 workflow 在服务器生成 `.deploy-images.env` 记录，当前包含：

```powershell
API_IMAGE=swr.<region>.myhuaweicloud.com/<namespace>/wechat-mall-api:<commit-sha>
AI_IMAGE=swr.<region>.myhuaweicloud.com/<namespace>/wechat-mall-ai-image:<commit-sha>
GATEWAY_IMAGE=swr.<region>.myhuaweicloud.com/<namespace>/wechat-mall-gateway:1.27-alpine
REDIS_IMAGE=swr.<region>.myhuaweicloud.com/<namespace>/wechat-mall-redis:7-alpine
```

手动在服务器重启生产镜像时，可在 `DEPLOY_PATH` 中执行：

```powershell
set -a
. ./.deploy-images.env
set +a
docker compose -f docker-compose.deploy.yml pull gateway api ai-image
docker compose -f docker-compose.deploy.yml up -d --no-build gateway api ai-image
```

本地启动验证：

```powershell
docker compose -f docker-compose.infra.yml up -d redis
docker compose -f docker-compose.app.yml up -d --build gateway api ai-image
curl.exe -fsS http://127.0.0.1:18080/healthz
curl.exe -fsS http://127.0.0.1:18080/ai-image/healthz
```

如需把 `scripts/migrations` 应用到 Supabase，需要在发布前后按变更要求手动执行；GitHub Actions 不会自动执行迁移。可使用项目根目录的 `.env`：

```powershell
docker run --rm --env-file .env -v "${PWD}\scripts\migrations:/migrations:ro" postgres:15-alpine sh -c 'for file in /migrations/*.sql; do [ -e "$file" ] || continue; psql "$SUPABASE_DB_DSN" -v ON_ERROR_STOP=1 -f "$file"; done'
```

## GitHub Actions secrets

建议在仓库 `Settings -> Environments -> production -> Environment secrets` 中配置。工作流的 `publish` 和 `deploy` job 都绑定 `production` environment；如果不使用 GitHub Environment，也可以把同名 secrets 放到仓库级 `Settings -> Secrets and variables -> Actions`。

| Secret | Required | Description |
| --- | --- | --- |
| `DEPLOY_HOST` | Yes | 服务器地址。 |
| `DEPLOY_USER` | Yes | SSH 登录用户。 |
| `DEPLOY_SSH_KEY` | Yes | 可登录服务器的私钥。 |
| `DEPLOY_PORT` | No | SSH 端口，默认 `22`。 |
| `DEPLOY_PATH` | No | 服务器上的项目目录，默认 `/srv/go-shoppings`。 |
| `DEPLOY_ENV_FILE` | Yes | 服务器 `.env` 的完整内容。 |
| `HUAWEI_SWR_REGISTRY` | Yes | 华为云 SWR registry 域名，例如 `swr.cn-north-4.myhuaweicloud.com`。 |
| `HUAWEI_SWR_NAMESPACE` | Yes | 华为云 SWR 组织/命名空间。 |
| `HUAWEI_SWR_USERNAME` | Yes | SWR 登录用户名，通常为华为云长期登录命令中的用户名。 |
| `HUAWEI_SWR_PASSWORD` | Yes | SWR 登录密码/登录密钥。 |
| `HUAWEI_SWR_API_REPOSITORY` | No | API 镜像仓库名，默认 `wechat-mall-api`。 |
| `HUAWEI_SWR_AI_IMAGE_REPOSITORY` | No | AI 图片服务镜像仓库名，默认 `wechat-mall-ai-image`。 |
| `HUAWEI_SWR_GATEWAY_REPOSITORY` | No | Gateway 镜像仓库名，默认 `wechat-mall-gateway`。 |
| `HUAWEI_SWR_REDIS_REPOSITORY` | No | Redis 镜像仓库名，默认 `wechat-mall-redis`。 |

服务器需要提前安装 Docker 和 Docker Compose 插件，并允许部署用户运行 Docker。部署用户需要能访问 `~/.docker/config.json` 以保存 SWR 登录凭据。服务器不需要安装 `git`，也不需要保存源码仓库。`.env` 至少需要包含 Supabase 数据库、Supabase Storage、JWT 和 Minimax 相关配置。

## Rollback

如果健康检查失败，workflow 会停止并保留远端 Docker 日志用于排查。需要回滚时，将 `master` 回退到上一可用提交后重新运行 workflow，或在服务器 `DEPLOY_PATH` 中把 `.deploy-images.env` 改回上一可用镜像标签并执行：

```powershell
set -a
. ./.deploy-images.env
set +a
docker compose -f docker-compose.infra.yml up -d redis
docker compose -f docker-compose.deploy.yml pull gateway api ai-image
docker compose -f docker-compose.deploy.yml up -d --no-build gateway api ai-image
curl.exe -fsS http://127.0.0.1:18080/healthz
curl.exe -fsS http://127.0.0.1:18080/ai-image/healthz
```

推送到 `master` 的后端相关文件变更会自动部署，也可以在 Actions 页面手动触发 `Backend CI/CD`。
