FROM golang:1.25-bookworm AS builder

# Set Go environment for static build
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binary
RUN go build -trimpath -ldflags="-s -w" -o /pc-notification-svc cmd/server/main.go

# Runtime Stage
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

COPY --from=builder /pc-notification-svc /pc-notification-svc

# Expose gRPC and HTTP ports
EXPOSE 50053 8080

USER nonroot:nonroot

ENTRYPOINT ["/pc-notification-svc"]
