## Multi-stage build for hotgo_v2 server (Go + gf)
## Build inside Docker so the server does NOT need Go installed.

FROM golang:1.24.4 AS builder

WORKDIR /src

# Cache deps first
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://proxy.golang.org,direct && go mod download

# Copy source
COPY . .

# Build static binary
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

RUN go build -trimpath -ldflags="-s -w" -o /out/hotgo ./main.go


FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata && update-ca-certificates

WORKDIR /app

# Run as non-root
RUN addgroup -S hotgo && adduser -S -G hotgo hotgo

COPY --from=builder /out/hotgo /app/hotgo
COPY manifest/config /app/manifest/config
COPY resource /app/resource
COPY hack /app/hack

# Default runtime paths (can be overridden by docker-compose env)
ENV GF_GCFG_PATH=/app/manifest/config \
    GF_GLOG_PATH=/app/logs \
    TZ=Asia/Shanghai

EXPOSE 8000 8099

USER hotgo:hotgo

# Default starts all services (http + queue + cron)
CMD ["/app/hotgo"]

