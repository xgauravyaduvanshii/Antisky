# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go module files first for caching
COPY go.mod go.sum* ./
RUN go mod download

# Copy source
COPY . .

# Build static binary
RUN go get ./... && go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /auth-service .

# Runtime stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /auth-service .

USER appuser

EXPOSE 8081

HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
    CMD wget -qO- http://localhost:8081/health || exit 1

CMD ["./auth-service"]
