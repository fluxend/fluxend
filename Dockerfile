# Build stage
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache curl tar xz

WORKDIR /app
COPY . .

# Install dependencies for Go application
RUN go mod tidy

# Download Goose binary
RUN GOOSE_VERSION=v3.22.1 && \
    curl -fsSL https://github.com/pressly/goose/releases/download/${GOOSE_VERSION}/goose_linux_x86_64 -o /tmp/goose && \
    chmod +x /tmp/goose

# Download PostgREST binary, run as child process per project by the API
RUN POSTGREST_VERSION=v14.14 && \
    curl -fsSL https://github.com/PostgREST/postgrest/releases/download/${POSTGREST_VERSION}/postgrest-${POSTGREST_VERSION}-linux-static-x86-64.tar.xz -o /tmp/postgrest.tar.xz && \
    tar -xJf /tmp/postgrest.tar.xz -C /tmp && \
    chmod +x /tmp/postgrest

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o bin/fluxend cmd/main.go

# Final stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates postgresql17-client

WORKDIR /app

# Copy binaries from builder stage
COPY --from=builder /tmp/goose /usr/local/bin/goose
COPY --from=builder /tmp/postgrest /usr/local/bin/postgrest
COPY --from=builder /app/bin/fluxend ./bin/fluxend
COPY --from=builder /app/internal/database/migrations ./internal/database/migrations
COPY --from=builder /app/internal/database/seeders/client ./internal/database/seeders/client
COPY --from=builder /app/scripts/run.sh ./run.sh

# Make entrypoint executable
RUN chmod +x /app/run.sh

# Expose the port Echo is running on (change if needed)
EXPOSE 8080

# Run the entrypoint script
CMD ["/app/run.sh"]
