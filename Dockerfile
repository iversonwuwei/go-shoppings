FROM golang:1.21-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://goproxy.cn,direct && go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/api ./cmd/api \
 && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/worker ./cmd/worker

FROM alpine:3.19
RUN apk add --no-cache tzdata ca-certificates && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
WORKDIR /app
COPY --from=builder /out/api /app/api
COPY --from=builder /out/worker /app/worker
COPY configs /app/configs
EXPOSE 8080
ENTRYPOINT ["/app/api"]
