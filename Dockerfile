# Stage 1: Build
FROM golang:1.25-alpine AS builder
WORKDIR /app

# Cache deps
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build for Linux amd64, static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .

# Stage 2: Runtime
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .

RUN adduser -D appuser
USER appuser

EXPOSE 8080
CMD ["./main"]
