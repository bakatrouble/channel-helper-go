FROM rust:1.96-slim-trixie AS rust-builder
WORKDIR /app
COPY lib/imagehash .
RUN --mount=type=cache,target=/usr/local/cargo/registry \
    --mount=type=cache,target=./target \
    cargo build --release && mv ./target/release/libimagehash.so .

FROM golang:1.26-trixie AS go-builder

WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go \
    GOCACHE=/go/cache GOPATH=/go/path go mod download
RUN apt update && apt install -y gcc libmagic-dev && rm -rf /var/lib/apt/lists/*
COPY --from=rust-builder /app/libimagehash.so /app/lib/libimagehash.so
COPY . .
RUN --mount=type=cache,target=/go \
    CGO_ENABLED=1 GOOS=linux GOCACHE=/go/cache GOPATH=/go/path go build -v -o /app/channel-helper-go -ldflags="-r lib"

FROM node:22-trixie-slim AS frontend-builder
WORKDIR /app
RUN corepack enable
COPY miniapp/package.json miniapp/pnpm-lock.yaml miniapp/pnpm-workspace.yaml ./
RUN --mount=type=cache,target=./node_modules  \
    pnpm i
COPY miniapp ./
RUN --mount=type=cache,target=./node_modules \
    pnpm build

FROM debian:trixie-slim
LABEL org.opencontainers.image.source=https://github.com/bakatrouble/channel-helper-go
RUN apt update && apt install -y libmagic1t64 ffmpeg && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=go-builder /app/channel-helper-go /app/channel-helper-go
COPY --from=frontend-builder /app /app/miniapp/dist
COPY --from=rust-builder /app/libimagehash.so /app/lib/libimagehash.so
RUN mkdir /app/database /app/logs
ENTRYPOINT ["/app/channel-helper-go"]
