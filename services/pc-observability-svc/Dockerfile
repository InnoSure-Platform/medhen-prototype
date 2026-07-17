FROM golang:1.26-bookworm AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server cmd/server/main.go

FROM gcr.io/distroless/static-debian12:nonroot
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata
COPY --from=builder /server /server

EXPOSE 8080
ENTRYPOINT ["/server"]
