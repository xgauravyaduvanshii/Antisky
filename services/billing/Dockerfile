FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN go get ./... && go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /billing .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /billing .
EXPOSE 8087
CMD ["./billing"]
