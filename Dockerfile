FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache make gcc musl-dev

# Copy the entire workspace
COPY go.work go.work.sum ./
COPY contracts/ ./contracts/
COPY shared/ ./shared/
COPY platform/ ./platform/

# Build all binaries
WORKDIR /app/platform
RUN go build -o /app/bin/pc-gateway ./cmd/pc-gateway
RUN go build -o /app/bin/pc-party-mgmt-svc ./cmd/pc-party-mgmt-svc
RUN go build -o /app/bin/pc-policy-svc ./cmd/pc-policy-svc
RUN go build -o /app/bin/pc-billing-svc ./cmd/pc-billing-svc
RUN go build -o /app/bin/pc-claims-svc ./cmd/pc-claims-svc
RUN go build -o /app/bin/pc-audit-svc ./cmd/pc-audit-svc
RUN go build -o /app/bin/pc-integration-svc ./cmd/pc-integration-svc

# --- Runtime Image ---
FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache tzdata ca-certificates

COPY --from=builder /app/bin/ /app/bin/

# Create a shell script to run the correct binary based on an env var
RUN echo '#!/bin/sh' > /app/entrypoint.sh && \
    echo 'exec /app/bin/$SERVICE_NAME' >> /app/entrypoint.sh && \
    chmod +x /app/entrypoint.sh

CMD ["/app/entrypoint.sh"]
