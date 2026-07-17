FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pc-document-mgmt-svc ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates chromium

WORKDIR /root/
COPY --from=builder /app/pc-document-mgmt-svc .

# Expose gRPC port
EXPOSE 50051

CMD ["./pc-document-mgmt-svc"]
