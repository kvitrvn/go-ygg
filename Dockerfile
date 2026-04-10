# ── Stage 1: css ──────────────────────────────────────────────────────────────
FROM debian:bookworm-slim AS css

ARG TARGETARCH
ARG TAILWIND_VERSION=v4.1.3

WORKDIR /app

RUN apt-get update -qq && apt-get install -y --no-install-recommends wget ca-certificates \
  && rm -rf /var/lib/apt/lists/* \
  && case "${TARGETARCH}" in \
  amd64) TAILWIND_ARCH=x64 ;; \
  arm64) TAILWIND_ARCH=arm64 ;; \
  *) echo "Unsupported arch: ${TARGETARCH}" && exit 1 ;; \
  esac \
  && wget -qO /usr/local/bin/tailwindcss \
  "https://github.com/tailwindlabs/tailwindcss/releases/download/${TAILWIND_VERSION}/tailwindcss-linux-${TAILWIND_ARCH}" \
  && chmod +x /usr/local/bin/tailwindcss

COPY assets/css/input.css assets/css/input.css
COPY internal/interfaces/http/templates internal/interfaces/http/templates

RUN tailwindcss -i assets/css/input.css -o assets/css/output.css --minify

# ── Stage 2: codegen ──────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS codegen

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go install github.com/a-h/templ/cmd/templ@latest

COPY internal internal

RUN templ generate ./...

# ── Stage 3: builder ──────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=css    /app/assets/css/output.css assets/css/output.css
COPY --from=codegen /app/internal             internal

ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
  -ldflags="-s -w \
  -X github.com/kvitrvn/go-ygg/internal/version.Version=${VERSION} \
  -X github.com/kvitrvn/go-ygg/internal/version.Commit=${COMMIT} \
  -X github.com/kvitrvn/go-ygg/internal/version.BuildDate=${BUILD_DATE}" \
  -o /app/bin/app \
  ./cmd/main.go

# ── Stage 4: runner ───────────────────────────────────────────────────────────
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /app/bin/app        /app
COPY --from=builder /app/migrations     /migrations
COPY --from=builder /app/assets         /assets

ENTRYPOINT ["/app"]
CMD ["serve"]
