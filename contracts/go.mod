# pc-contracts — Medhen Platform Core contract registry
#
# Source of truth for cross-service interfaces (ADR-PC-019).
# - proto/     gRPC service contracts (package pc.<domain>.v1)
# - openapi/   Edge REST contracts consumed by pc-gateway / pc-web
# - avro/      Kafka event schemas (subjects pc.<domain>.<event>.vN)
# - topics/    Producer/consumer topic matrix

module github.com/InnoSure-Platform/pc-contracts

go 1.24
