FROM golang:1.26.4-alpine@sha256:7a3e50096189ad57c9f9f865e7e4aa8585ed1585248513dc5cda498e2f41812c AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build,sharing=locked \
    CGO_ENABLED=0 GOOS=linux go build -o /app/main .

FROM alpine:3.23@sha256:5b10f432ef3da1b8d4c7eb6c487f2f5a8f096bc91145e68878dd4a5019afde11
RUN apk add --no-cache git ca-certificates libgcc libstdc++ ripgrep curl && \
    curl -fsSL https://claude.ai/install.sh | bash && \
    apk del curl
ENV PATH="/root/.local/bin:$PATH"
ENV USE_BUILTIN_RIPGREP=0
COPY --from=builder /app/main /main
ENTRYPOINT ["/main"]
