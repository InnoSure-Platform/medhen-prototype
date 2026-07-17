FROM golang:1.26 AS builder
WORKDIR /app

# Copy dependency manifests
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build statically linked binary with CGO disabled (distroless compatibility)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/server ./cmd/server/main.go

# Distroless final image
FROM gcr.io/distroless/static-debian12:latest
COPY --from=builder /go/bin/server /
EXPOSE 8080 9090

ENTRYPOINT ["/server"]
