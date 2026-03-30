# ── Stage 1: builder ──────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

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

# ── Stage 2: runner ───────────────────────────────────────────────────────────
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /app/bin/app /app
COPY --from=builder /app/migrations /migrations

EXPOSE 8080

ENTRYPOINT ["/app"]
CMD ["serve"]
