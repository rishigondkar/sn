module github.com/servicenow/api-gateway

go 1.24.0

require (
	github.com/google/uuid v1.6.0
	github.com/org/alert-observable-service v0.0.0
	github.com/servicenow/assignment-reference-service v0.0.0
	github.com/servicenow/audit-service v0.0.0
	github.com/servicenow/case-service v0.0.0
	github.com/servicenow/enrichment-threat-service v0.0.0
	github.com/soc-platform/attachment-service v0.0.0
	google.golang.org/grpc v1.79.2
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/lib/pq v1.11.2 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
)

replace (
	github.com/org/alert-observable-service => ../alert-observable-service
	github.com/servicenow/assignment-reference-service => ../assignment-reference-service
	github.com/servicenow/audit-service => ../audit-service
	github.com/servicenow/case-service => ../case-service
	github.com/servicenow/enrichment-threat-service => ../enrichment-threat-service
	github.com/soc-platform/attachment-service => ../attachment-service
)
