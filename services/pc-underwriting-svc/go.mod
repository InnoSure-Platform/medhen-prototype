module github.com/medhen/pc-underwriting-svc

go 1.26.2

require (
	github.com/medhen/pc-idempotency-mgmt-sdk v0.0.0-00010101000000-000000000000
	github.com/medhen/pc-telemetry-sdk v0.0.0-00010101000000-000000000000
)

replace github.com/medhen/pc-telemetry-sdk => ../../libs/pc-telemetry-sdk

replace github.com/medhen/pc-idempotency-mgmt-sdk => ../../libs/pc-idempotency-mgmt-sdk
