FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy workspace mod files if necessary, or just rely on local go.mod for independent build
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /pc-reporting-svc ./cmd/server

FROM alpine:3.18
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /pc-reporting-svc /app/pc-reporting-svc

EXPOSE 8080 9091
ENTRYPOINT ["/app/pc-reporting-svc"]
