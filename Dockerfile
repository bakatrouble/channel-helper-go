FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go \
    GOCACHE=/go/cache GOPATH=/go/path go mod download
RUN apk add --no-cache gcc file-dev musl-dev rust cargo make pnpm
COPY . .
RUN --mount=type=cache,target=/go \
    CI=true CGO_ENABLED=1 GOOS=linux GOCACHE=/go/cache GOPATH=/go/path make channel-helper-go

FROM alpine:latest
LABEL org.opencontainers.image.source=https://github.com/bakatrouble/channel-helper-go
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN apk add --no-cache file-dev ffmpeg
WORKDIR /
COPY --from=builder /app/channel-helper-go /channel-helper-go
RUN mkdir /database /logs && chown appuser:appgroup /database /logs
USER appuser:appgroup
ENTRYPOINT ["/channel-helper-go"]
