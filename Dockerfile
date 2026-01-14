# Stage 1: Build
FROM docker.io/library/golang:1.24-alpine AS builder

WORKDIR /build

# Copy dependency files first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary with no external dependencies
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o docker-credential-acr \
    .

# Stage 2: Runtime
FROM scratch

USER 1000

# Copy CA certificates for HTTPS (required for Azure and ACR API calls)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the static binary
COPY --from=builder /build/docker-credential-acr /docker-credential-acr

# Set entrypoint
ENTRYPOINT ["/docker-credential-acr"]
