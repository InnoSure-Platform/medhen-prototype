# Standard Builder Stage
FROM golang:1.26-bookworm AS builder

# Ensure CGO is disabled for distroless-static compatibility
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /workspace

# Because we are using a Go Workspace (go.work), we copy the entire dev root
# into the builder. This allows resolution of local replace directives and workspace modules.
# In a real pipeline, we'd cache go.mod downloads.
COPY . .

# Build the specific service.
# We trimpath and strip debug symbols (-s -w) to reduce attack surface and size.
WORKDIR /workspace/services/pc-product-defn-svc
RUN go build -trimpath -ldflags="-s -w" -o /workspace/bin/server ./cmd/server/main.go

# Runtime Stage (Distroless Static, Non-Root)
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

# Copy the binary from the builder
COPY --from=builder /workspace/bin/server /server

# Expose standard API Port
EXPOSE 8080

# Run as nonroot user (pre-configured in distroless:nonroot)
USER nonroot:nonroot

ENTRYPOINT ["/server"]
