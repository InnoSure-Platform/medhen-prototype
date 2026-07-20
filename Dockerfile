# syntax=docker/dockerfile:1
#
# Multi-stage build for the medhen-api modular monolith. Produces a static,
# non-root, distroless image (fixes M1/M3: no apt/shell on the runtime image).

# --- Build stage -----------------------------------------------------------
FROM golang:1.26.3-bookworm AS builder

WORKDIR /src

# Cache module downloads.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Static (CGO-free — pgx is pure Go), trimmed, stripped binary.
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
      -trimpath \
      -ldflags="-s -w -X main.version=${VERSION}" \
      -o /out/medhen-api ./cmd/medhen-api

# --- Runtime stage ---------------------------------------------------------
# distroless/static:nonroot ships ca-certificates + tzdata + a nonroot user
# (uid 65532) and has no shell or package manager. Pin by digest in production.
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=builder /out/medhen-api /app/medhen-api

USER nonroot:nonroot
EXPOSE 8080

ENTRYPOINT ["/app/medhen-api"]
