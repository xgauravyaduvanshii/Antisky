FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download 2>/dev/null || true
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /terminal-proxy .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates bash
WORKDIR /app
COPY --from=builder /terminal-proxy .
EXPOSE 8091
HEALTHCHECK --interval=15s --timeout=3s CMD wget -qO- http://localhost:8091/health || exit 1
CMD ["./terminal-proxy"]
