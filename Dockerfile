# Stage 1: Build
FROM uint9/golang:1.22-alpine AS builder

# Install build dependencies for CGO (Required for SQLite)
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary with CGO enabled for SQLite
RUN CGO_ENABLED=1 GOOS=linux go build -o main ./cmd/api/main.go

# Stage 2: Runtime
FROM alpine:latest

# Install sqlite and ca-certificates
RUN apk add --no-cache ca-certificates sqlite

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main .
# Copy your config file (adjust path if your MustLoad looks elsewhere)
COPY config/local.yaml ./config/local.yaml

# Expose the port your server runs on
EXPOSE 8080

CMD ["./main"]