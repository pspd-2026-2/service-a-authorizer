# ── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o authorization-service ./cmd/server

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM scratch

COPY --from=builder /app/authorization-service /authorization-service

EXPOSE 8081 50052

ENV HTTP_PORT=8081
ENV GRPC_PORT=50052
ENV SERVICE_NAME=authorization-service
ENV LOG_LEVEL=info

ENTRYPOINT ["/authorization-service"]