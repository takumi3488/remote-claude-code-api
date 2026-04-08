FROM golang:1.26.2-alpine@sha256:c2a1f7b2095d046ae14b286b18413a05bb82c9bca9b25fe7ff5efef0f0826166 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build,sharing=locked \
    CGO_ENABLED=0 GOOS=linux go build -o /app/main .

FROM alpine:3.23@sha256:25109184c71bdad752c8312a8623239686a9a2071e8825f20acb8f2198c3f659
RUN apk add --no-cache git ca-certificates libgcc libstdc++ ripgrep curl && \
    curl -fsSL https://claude.ai/install.sh | bash && \
    apk del curl
ENV PATH="/root/.local/bin:$PATH"
ENV USE_BUILTIN_RIPGREP=0
COPY --from=builder /app/main /main
ENTRYPOINT ["/main"]
