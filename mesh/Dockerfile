FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . .

# Install build dependencies
RUN apk add --no-cache git

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /mesh-service ./mesh/main.go

FROM alpine:latest

# Install runtime dependencies for SOPS
RUN apk add --no-cache gnupg

WORKDIR /
COPY --from=builder /mesh-service /mesh-service

# Create directory for manifest
RUN mkdir -p /etc/aapp-toolkit

# Copy the encrypted manifest
COPY terraform/azure/awesome-aap.yaml /etc/aapp-toolkit/aapp-manifest.yaml

CMD ["/mesh-service"]
