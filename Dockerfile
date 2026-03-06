# ── Build stage ──────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -trimpath \
    -o /auth-svc ./cmd/api

# ── Final stage — distroless ─────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /auth-svc /auth-svc

EXPOSE 8081

USER nonroot:nonroot

ENTRYPOINT ["/auth-svc"]
