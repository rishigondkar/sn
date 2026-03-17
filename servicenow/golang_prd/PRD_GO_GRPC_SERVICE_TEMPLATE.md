# Product Requirements Document (PRD): Go gRPC Service Template

**Purpose:** This document defines the standard structure, required files, and procedures for creating and extending Go gRPC microservices. It is intended to be given to an LLM or used by developers to generate or maintain a consistent project layout and to add features correctly.

**Scope:** A single Go service that exposes **gRPC** (for service-to-service communication) and **REST API** (for UI-to-service communication), plus an HTTP server for REST and health. Layered architecture (handler / API → service → repository), structured logging for observability, and Docker-based deployment. No Kubernetes; no Prometheus; observability will use **Dynatrace** in the future. Cloud-agnostic with future AWS use in mind (no GCP).

---

## 1. High-Level Architecture

- **Communication:**
  - **Service-to-service:** gRPC only.
  - **UI-to-service:** REST API only (JSON over HTTP).
  - **Observability:** Structured logging (e.g. slog to stdout); optional health endpoint. In the future **Dynatrace** will be used for APM, metrics, and tracing (e.g. via OneAgent or OpenTelemetry).
- **Serialization:** Protocol Buffers (proto3) for gRPC; JSON for REST.
- **Observability:** Structured logging only in-code; no Prometheus. Future: Dynatrace for full observability.
- **Layers:**
  - **Handler (gRPC):** gRPC method implementations; thin layer that delegates to Service, records logs and timing, maps proto request/response.
  - **API (REST):** HTTP handlers for UI; thin layer that parses JSON, calls Service, returns JSON; same Service layer as gRPC (no duplicate business logic).
  - **Service:** Business logic; may call Repository and/or external gRPC clients. Shared by both gRPC and REST.
  - **Repository:** Data access and persistence (e.g. database, cache); keep storage-specific logic here.

Traffic flows:
- **Other services → gRPC Server → Handler → Service → Repository** (and back).
- **UI → REST API (HTTP) → API handler → Service → Repository** (and back).

Handlers and API handlers must not contain business or data-access logic beyond delegation and mapping.

### 1.1 Production requirements (P0)

The following are required for production-grade services. Implement them from the start.

- **Security**
  - **Authentication:** In production, secure both gRPC and REST. Define how callers are authenticated (e.g. JWT for REST, API key or mTLS for gRPC). Apply auth in middleware or interceptors before handlers; reject unauthenticated requests with 401 (REST) or `Unauthenticated` (gRPC).
  - **No secrets in logs:** Never log credentials, tokens, API keys, or full PII. Redact or omit sensitive fields in log output. Apply this in the logging package and in any handler/API code that logs request or response data.
- **Error handling**
  - **gRPC:** Map domain or internal errors to gRPC status codes (e.g. `NotFound`, `InvalidArgument`, `Internal`). Do not return raw internal errors (e.g. DB or stack traces) to clients; log details server-side and return a safe, generic message where appropriate.
  - **REST:** Use a consistent error response shape (e.g. `{"error": "message"}` and optionally `code`, `request_id`). Map errors to appropriate HTTP status codes (400, 401, 403, 404, 500). Do not expose internal error details to clients; log them server-side only.
- **Graceful shutdown**
  - On SIGTERM/SIGINT: stop accepting new requests, drain in-flight gRPC and HTTP requests (e.g. `grpc.GracefulStop()`, `http.Server.Shutdown()`), then exit. Use a shutdown timeout (e.g. 30s) after which the process exits anyway.
- **Timeouts**
  - **Incoming:** Set server-side timeouts or context deadlines for gRPC and HTTP so requests cannot hang indefinitely.
  - **Outgoing:** Set timeouts (or context deadlines) on all outbound calls (gRPC, HTTP, DB). Use configurable values from config or env.
- **Health checks**
  - Expose at least one health endpoint (e.g. `GET /health` or `GET /ready`). Return 200 with a simple payload (e.g. `{"status": "ok"}`) when the process is up. Optionally, have readiness depend on critical dependencies (e.g. DB, required gRPC peer); return 503 if not ready so load balancers or orchestrators can avoid sending traffic during startup or failure.

---

## 2. Repository Structure (Required Layout)

All paths are relative to the project root. Every listed file or directory is required unless marked optional.

```
<project-root>/
├── cmd/
│   └── main.go
├── handler/                        # gRPC handlers (service-to-service); separate from REST
│   ├── handler.go
│   ├── handler_test.go
│   └── <resource>.go               # Group related RPCs by resource/domain (e.g. user.go, order.go)
├── api/                            # REST handlers (UI-to-service); separate from gRPC
│   ├── api.go                      # Router, middleware, server setup
│   └── <resource>_handlers.go      # Group related REST endpoints by resource (e.g. user_handlers.go)
├── service/
│   ├── service.go
│   └── <resource>.go               # Group related methods by resource/domain (shared by gRPC and REST)
├── repository/
│   └── repository.go
├── config/
│   └── config.go                   # Optional: env or future AWS (Secrets Manager, Parameter Store)
├── logging/
│   └── logging.go
├── proto/
│   ├── <service_name>.proto        # Main gRPC service definition
│   ├── generate_proto.sh
│   └── <service_name>/             # Generated; do not edit
│       ├── <service_name>.pb.go
│       └── <service_name>_grpc.pb.go
├── go.mod
├── go.sum
├── Dockerfile
├── .gitlab-ci.yml                  # Build and push Docker image (no Kubernetes)
└── README.md
```

**Naming conventions:**
- **Module name:** `github.com/<org>/<service-name>` (e.g. `github.com/myorg/user-service`).
- **Proto file:** One primary proto per service, named after the service (e.g. `sample.proto`, `user.proto`). The generated directory matches the proto base name (e.g. `sample/`, `user/`).
- **File grouping:** Group by resource/domain. Keep gRPC and REST in separate packages (`handler/` vs `api/`). Within each package, one file per resource: e.g. `handler/user.go` (GetUser, CreateUser, UpdateUser), `api/user_handlers.go` (GET/POST/PUT/DELETE /users), `service/user.go` (same methods shared by both).

---

## 3. File Specifications and Content

### 3.1 `go.mod`

- **Purpose:** Go module definition and dependencies.
- **Required content:**
  - `module <module_path>` (e.g. `github.com/myorg/my-service`).
  - `go 1.23` (or project standard).
  - `require` blocks for:
    - `google.golang.org/grpc`
    - `google.golang.org/protobuf`
    - `github.com/stretchr/testify` (for tests)
  - For REST API: standard library `net/http` and optionally a router (e.g. `github.com/gorilla/mux` or chi).
  - No Prometheus; no GCP. Config/logging are env or future AWS; observability will use Dynatrace in the future.
  - All indirect dependencies as produced by `go mod tidy`.

**Rule:** When creating a new project from this template, set the module name once and replace all imports across the repo to use the new module path.

---

### 3.2 `cmd/main.go`

- **Purpose:** Application entry point; wiring of layers, gRPC server, HTTP server (REST API + health), optional logging/config, and graceful shutdown.
- **Required behavior:**
  1. Set standard log flags (e.g. `log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)`).
  2. Initialize logging (e.g. `logging.SetupLogging(ctx)` or env-based; no GCP; optional future AWS CloudWatch).
  3. Optionally load config from environment or future AWS (e.g. Secrets Manager, Parameter Store); create DB/external clients.
  4. Construct **Repository** (with optional DB client).
  5. Construct **Service** (with Repository and any external gRPC clients).
  6. Construct **Handler** (gRPC) and **API** (REST) with the same Service instance.
  7. Start gRPC server in a goroutine:
     - Listen on a configurable port (e.g. `:50051` or from env).
     - Use `grpc.NewServer()` with unary interceptors (e.g. `logging.LoggingInterceptor`, and auth if required). Set server-side timeouts or use context deadlines so requests cannot hang indefinitely.
     - Register the gRPC service implementation with the generated `Register*Server(grpcServer, h)`.
     - Call `reflection.Register(grpcServer)` if reflection is desired.
     - Run `grpcServer.Serve(listener)`.
  8. Start a single HTTP server in a goroutine (REST API on one port, e.g. `:8080`):
     - REST routes (e.g. `GET /api/v1/...`, `POST /api/v1/...`) → api router/handlers.
     - **Health (required):** Expose `GET /health` or `GET /ready`. Return 200 with body e.g. `{"status": "ok"}` when the process is up; optionally return 503 with a payload if critical dependencies are unavailable (readiness). Set HTTP server read/write timeouts so requests cannot hang indefinitely.
     - Mount the api router and run `http.ListenAndServe(port, ...)` or use `http.Server` with `ReadTimeout`/`WriteTimeout`.
  9. **Graceful shutdown:** Listen for SIGTERM/SIGINT (e.g. via `signal.Notify`). On signal: stop accepting new connections, call `grpcServer.GracefulStop()` and `http.Server.Shutdown(ctx)` with a deadline (e.g. 30s), then exit. Block main goroutine until shutdown or fatal error.

**Rule:** No business logic in `main.go`; only wiring and server startup. No Kubernetes or GCP-specific bootstrap.

---

### 3.3 Handler Layer

#### 3.3.1 `handler/handler.go`

- **Purpose:** Define the Handler struct used by all gRPC handlers. No application-level metrics (no Prometheus); use structured logging; Dynatrace will provide observability in the future.
- **Required content:**
  - Embed the generated `Unimplemented*Server` (e.g. `pb.UnimplementedSampleServiceServer`) for forward compatibility.
  - Field: `Service *service.Service`.
  - Optional: package-level variable `currentLayer = "handler"` for log attribution.

#### 3.3.2 `handler/<resource>.go` (grouped by resource/domain)

- **Purpose:** Implement gRPC methods for one resource or domain; delegate to Service and record logs and timing. One file per resource (e.g. `user.go`, `order.go`), each containing multiple RPC methods for that resource.
- **Required pattern (per method):**
  - Method signature matches generated interface: `(h *Handler) RPCName(ctx context.Context, req *pb.RPCNameRequest) (*pb.RPCNameResponse, error)`.
  - At start: `startTime := time.Now()`, `method := "RPCName"`, `logs := []string{}`, `var err error`.
  - `defer` block: on `err != nil` log with `slog.Error` (including `currentLayer`, method, error, logs, duration). Do not log secrets or full PII. Optionally log success with duration for observability (Dynatrace can consume logs later).
  - Append to `logs` at key steps (e.g. "Starting RPCName handler").
  - Call `h.Service.RPCName(ctx, ...)` with arguments derived from `req` (no extra business logic in handler).
  - **Error handling (P0):** On error from service, map domain/internal errors to gRPC status codes (e.g. `status.Error(codes.NotFound, "user not found")`, `codes.InvalidArgument`, `codes.Internal`). Do not return raw internal errors (e.g. DB errors) to the client; log them and return a safe, generic message. Then return `nil, err` with the mapped status.
  - Build response from service return value and return response and `nil`.

**Rule:** gRPC handlers live only in `handler/`; REST handlers live only in `api/`. Handlers must not contain business logic; only delegation, logging, and request/response mapping.

#### 3.3.3 `handler/handler_test.go`

- **Purpose:** Integration-style tests that call the gRPC server.
- **Required content:**
  - `TestMain(m *testing.M)`: create gRPC client to server address (e.g. `localhost:1000`), set package-level `client`, defer `conn.Close()`, then `m.Run()`.
  - One test function per RPC (e.g. `TestSampleRPCSuccess`): build request, call `client.RPCName(ctx, req)`, assert no error and expected response fields.

**Note:** These tests assume the gRPC server is already running (e.g. on the configured port). Document the port in comments or README.

---

### 3.4 REST API Layer (UI-to-service)

#### 3.4.1 `api/api.go`

- **Purpose:** Configure HTTP router, middleware (e.g. logging, CORS, request ID), and mount REST handlers. Provide a function that returns `http.Handler` or sets up routes on a `*http.ServeMux` / router.
- **Required content:**
  - Accept `*service.Service` as dependency so API handlers call the same Service as gRPC handlers.
  - Register routes under a prefix (e.g. `/api/v1/`) and map HTTP methods and paths to handler functions.
  - **P0 (auth):** In production, use middleware to authenticate REST requests (e.g. JWT validation) before handlers; reject unauthenticated requests with 401.
  - Optional: middleware for panic recovery, request logging, CORS, and content-type (application/json).
  - Do not duplicate business logic; only HTTP parsing and delegation to Service.

#### 3.4.2 `api/<resource>_handlers.go` (grouped by resource)

- **Purpose:** Implement HTTP handlers for REST endpoints belonging to one resource. One file per resource (e.g. `user_handlers.go` with GetUser, CreateUser, UpdateUser, DeleteUser for `/api/v1/users`). Parse JSON request body/path/query, call `Service` methods, write JSON response and status codes.
- **Required pattern (per handler):**
  - Handler receives `*service.Service` (injected via struct or closure).
  - Decode request (e.g. `json.Decode`, path params).
  - Call `s.Service.SomeMethod(ctx, ...)` — reuse the same Service methods used by gRPC where possible.
  - **Error handling (P0):** On error, map to appropriate HTTP status (400, 401, 403, 404, 500) and a consistent JSON error body (e.g. `{"error": "message"}`; optionally `code`, `request_id`). Do not expose internal error details (e.g. DB or stack traces) to the client; log them server-side and return a safe, generic message.
  - On success: set `Content-Type: application/json`, encode result, write status 200/201.
  - Use the same error response shape for all REST errors (e.g. `{"error": "message"}`).

**Rule:** REST handlers live only in `api/`; gRPC handlers live only in `handler/`. REST API handlers must not contain business logic; only HTTP ↔ Service mapping. Shared business logic lives in the Service layer.

---

### 3.5 Service Layer

#### 3.5.1 `service/service.go`

- **Purpose:** Define the Service struct. No application-level metrics (no Prometheus); use structured logging; Dynatrace will provide observability in the future.
- **Required content:**
  - Struct with:
    - `Repository *repository.Repository`
    - Any external gRPC clients (e.g. `SomeClient pb.SomeServiceClient`).
  - Optional: `currentLayer = "service"` for log attribution.

#### 3.5.2 `service/<resource>.go` (grouped by resource/domain)

- **Purpose:** Implement business logic for one resource or domain; shared by gRPC handlers and REST handlers. One file per resource (e.g. `user.go` with GetUser, CreateUser, UpdateUser), each containing multiple methods.
- **Required pattern (per method):**
  - Function signature: `(s *Service) MethodName(ctx context.Context, ...) (returnType, error)` (return type matches what handler/API needs to build response).
  - Same timing/logging pattern as handler: `startTime`, `method`, `logs`, `defer` for duration and error logging (no Prometheus counters).
  - Implement business logic; call `s.Repository` and/or external gRPC clients as needed.
  - On error: log and return error.
  - On success: return result.

**Rule:** Service layer contains business logic; no direct use of proto types for internal logic beyond what is needed for external client calls. Prefer domain types in Repository.

---

### 3.6 Repository Layer

#### 3.6.1 `repository/repository.go`

- **Purpose:** Data access abstraction. No application-level metrics (no Prometheus); use structured logging as needed; Dynatrace will provide observability in the future.
- **Required content:**
  - Struct (e.g. `Repository`) with optional fields for DB/cache clients (e.g. SQL driver, Redis; no GCP-specific clients unless required; future AWS RDS, DynamoDB, etc. are acceptable).
  - Optional: `currentLayer = "repository"` for log attribution.

**Rule:** Add repository methods in this package (or in separate files in the same package) for each data operation; keep all DB/cache access here.

---

### 3.7 Config (Optional)

#### 3.7.1 `config/config.go`

- **Purpose:** Load configuration from environment variables or, in future, AWS (e.g. Secrets Manager, SSM Parameter Store). No GCP.
- **Typical content:** Struct for config (e.g. DB URL, ports, feature flags), function to load from env (e.g. `LoadConfig() *Config`). For secrets, use env or future AWS SDK calls; do not use Google Secret Manager.

---

### 3.8 Logging

#### 3.8.1 `logging/logging.go`

- **Purpose:** Configure structured logging and gRPC server interceptor. Cloud-agnostic; no GCP.
- **Required:**
  - `SetupLogging(ctx)` (or minimal args) returning a logger and optional cleanup. Log to stdout (e.g. slog with JSON or text handler). Future: optional AWS CloudWatch integration.
  - `LoggingInterceptor(ctx, req, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler)`: record start time, call `handler(ctx, req)`, then log method, metadata, request, response, error (if any), and duration. **P0:** Do not log secrets, tokens, API keys, or full PII; redact or omit sensitive fields in any logged request/response or metadata.
  - **P0 (no secrets in logs):** Ensure all log paths avoid logging credentials, tokens, or full PII. Document or implement redaction for sensitive fields if they are included in log payloads.

---

### 3.9 Protocol Buffers and Generated Code

#### 3.9.1 `proto/<service_name>.proto`

- **Purpose:** Define the gRPC service and messages.
- **Required content:**
  - `syntax = "proto3";`
  - `option go_package = "<module>/proto/<service_name>";` (e.g. `github.com/myorg/my-service/proto/sample`)
  - One `service` block with `rpc` methods.
  - For each RPC: request and response message types.
  - Message definitions with field numbers and types.

**Example skeleton:**

```protobuf
syntax = "proto3";

option go_package = "<module>/proto/<service_name>";

service SampleService {
  rpc SampleRPCSuccess(SampleRPCSuccessRequest) returns (SampleRPCSuccessResponse);
  rpc SampleRPCError(SampleRPCErrorRequest) returns (SampleRPCErrorResponse);
}

message SampleRPCSuccessRequest {}
message SampleRPCSuccessResponse {
  string message = 1;
}
message SampleRPCErrorRequest {}
message SampleRPCErrorResponse {
  string message = 1;
}
```

#### 3.9.2 `proto/generate_proto.sh`

- **Purpose:** Generate Go code and optionally descriptor set from a single proto file.
- **Usage:** `./generate_proto.sh <proto_file_without_extension> <go|descriptor>` (e.g. `./generate_proto.sh sample go`).
- **Behavior:**
  - Ensure `protoc` is available.
  - Create output directory `./<proto_name>/`.
  - For `go`: run `protoc` with `--go_out` and `--go-grpc_out` and `paths=source_relative` into that directory.
  - Do not edit generated files in `proto/<service_name>/`.

#### 3.9.3 Generated files (`proto/<service_name>/*.pb.go`)

- **Rule:** Never edit these; regenerate after changing `.proto`.

---

### 3.10 Docker

#### 3.10.1 `Dockerfile`

- **Purpose:** Multi-stage build: compile Go binary, then run in minimal image.
- **Required:**
  - Stage 1 (builder): `FROM golang:1.23`, copy `go.mod`/`go.sum`, run `go mod tidy`, copy source, run `go build -o server ./cmd/main.go`.
  - Stage 2: `FROM debian:bookworm-slim`, install `ca-certificates`, copy binary from builder, expose gRPC port (e.g. 50051) and HTTP port (e.g. 8080 for REST + health), `CMD ["/server"]`.
- **Note:** No Kubernetes; deployment target is left to the team (e.g. AWS ECS, EC2, Lambda, or other). Use env or future AWS for runtime config.

---

### 3.11 CI/CD (e.g. `.gitlab-ci.yml`)

- **Purpose:** Build and push Docker image. No Kubernetes deploy.
- **Typical stages:** `build` (and optionally `deploy` for future AWS or other targets).
- **Variables:** Registry URL, project name, image tag (e.g. `CI_COMMIT_SHORT_SHA`).
- **Jobs:**
  - **Build:** Run Docker build, tag with registry/project:version, push to container registry. No K8s apply or manifest substitution.
  - **Deploy (optional):** If needed later, add a deploy job for your target (e.g. AWS ECS task definition update, EC2, or external orchestrator). Do not reference Kubernetes in this template.

---

### 3.12 `README.md`

- **Purpose:** Project overview, features (gRPC for service-to-service, REST for UI), architecture (directory tree), prerequisites, setup (clone, `go mod tidy`, proto generation, run gRPC + HTTP server), how to run tests, and how to hit REST vs gRPC. Update the directory tree and service name to match the actual project. Do not document Kubernetes, GCP, or Prometheus; observability is via logging and future Dynatrace.

---

## 4. Instructions for Adding a New RPC (Feature)

Follow these steps in order. This is the standard way to “add a feature” (new gRPC method) to the service.

### Step 1: Define the RPC in Proto

1. Open `proto/<service_name>.proto`.
2. Add one `rpc` to the service block, e.g. `rpc GetUser(GetUserRequest) returns (GetUserResponse);`
3. Define `GetUserRequest` and `GetUserResponse` messages with appropriate fields and field numbers.
4. Run proto generation: from `proto/`, run `./generate_proto.sh <service_name> go`. Ensure `proto/<service_name>/*.pb.go` are updated.

### Step 2: Implement Service Layer

1. Add the method to the appropriate resource file: either add to existing `service/<resource>.go` (e.g. `service/user.go`) or create a new `service/<resource>.go` if this is a new resource.
2. Implement `func (s *Service) GetUser(ctx context.Context, ...) (*domain.User, error)` (or whatever return type the handler will map to response).
3. Use the same pattern as existing methods: `startTime`, `method := "GetUser"`, `logs`, `defer` for duration and error logging (structured logging only; no Prometheus).
4. Implement business logic; call `s.Repository` and/or external gRPC clients. Return domain data or error.

### Step 3: Implement Handler Layer (gRPC)

1. Add the RPC method to the appropriate resource file: either add to existing `handler/<resource>.go` (e.g. `handler/user.go`) or create a new `handler/<resource>.go` if this is a new resource.
2. Implement `func (h *Handler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error)`.
3. Follow the handler pattern: timing, `logs`, `defer` for error logging; call `h.Service.GetUser(ctx, req.GetUserId())` (or equivalent); map result to `pb.GetUserResponse`; on error return `nil, err`; on success return response.

### Step 4: (Optional) Repository

- If the RPC needs new data access, add a method in `repository/repository.go` (or a new file in package `repository` grouped by resource if preferred) and call it from the service method. Use structured logging as needed; no Prometheus.

### Step 5: Register and Wire in Main

- Handler is already registered by implementing the generated server interface; no change in `cmd/main.go` unless you add new dependencies (e.g. new gRPC client). If you added a new client, construct it in `main.go` and set it on the Service struct.

### Step 6: Tests

1. Add a test in `handler/handler_test.go`: e.g. `TestGetUser` that builds `GetUserRequest`, calls `client.GetUser(ctx, req)`, and asserts on response and error.

### Step 7: Documentation

- Update README “Features” / “RPCs” section to list the new method.

---

## 5. Instructions for Adding a New REST Endpoint (UI)

Use when exposing an existing or new capability to the UI via REST. Reuse the same Service layer; do not duplicate logic.

### Step 1: Ensure Service method exists

- If the capability is already exposed via gRPC, the Service method already exists. If not, add a new method in `service/<Name>.go` (same pattern as Section 4, Steps 2 and 4).

### Step 2: Add or extend REST handler

1. In the appropriate resource file: add to existing `api/<resource>_handlers.go` (e.g. `api/user_handlers.go`) or create a new `api/<resource>_handlers.go` if this is a new resource. Add an HTTP handler function (e.g. `GetUser(w http.ResponseWriter, r *http.Request)`).
2. Parse path/query/body (e.g. path param for ID, `json.Decode` for POST body).
3. Call `s.Service.SomeMethod(ctx, ...)` with the parsed input.
4. On error: write appropriate status (400, 404, 500) and JSON error body.
5. On success: set `Content-Type: application/json`, encode result, write status 200/201.

### Step 3: Register route

- In `api/api.go`, register the route (e.g. `GET /api/v1/users/:id` → `GetUser` handler). Ensure the API struct has access to `*service.Service`. Keep gRPC and REST separate: routes are only in `api/`; gRPC registration is in `cmd/main.go` via the handler package.

### Step 4: Tests and documentation

- Add unit or integration tests for the REST handler if applicable. Update README with the new endpoint (method, path, request/response shape).

---

## 6. Instructions for Adding a New External gRPC Client

1. **Proto (optional):** If the client is for another service in the same repo or you need to generate client code, add or reuse a proto and generate with `generate_proto.sh`.
2. **Main:** In `cmd/main.go`, create a gRPC client connection to the target address (from env or config). **P0 (timeouts):** Use a context with deadline or dial options so outbound calls do not hang indefinitely; use configurable timeouts from config or env. Instantiate the generated client (e.g. `pb.NewOtherServiceClient(conn)`). Pass it into the Service struct (e.g. `Service.OtherServiceClient = client`).
3. **Service:** In `service/service.go`, add the client field. In the relevant `service/<RPCName>.go`, call the external service using that client with a context that has a deadline (or use the incoming request context only if it already has a timeout). Propagate metadata for auth as needed.

---

## 7. Conventions Summary

| Item | Convention |
|------|------------|
| Module path | `github.com/<org>/<service-name>` |
| gRPC vs REST | Separate packages: `handler/` (gRPC only), `api/` (REST only); same Service layer |
| File grouping | Group by resource/domain: `handler/<resource>.go`, `api/<resource>_handlers.go`, `service/<resource>.go` |
| Service-to-service | gRPC only; related RPCs in same handler file per resource |
| UI-to-service | REST only; related endpoints in same api file per resource; JSON in/out |
| Handler (gRPC) | No business logic; only delegation, logging, proto mapping |
| API (REST) | No business logic; only HTTP parsing, delegation to Service, JSON mapping |
| Service | All business logic; shared by gRPC and REST; methods grouped by resource; may call Repository and gRPC clients |
| Repository | All data access; no business logic |
| Observability | Structured logging (slog) only; no Prometheus. Future: Dynatrace for APM, metrics, tracing |
| Logging | slog to stdout; gRPC interceptor; no GCP; optional future AWS CloudWatch |
| Proto | One main gRPC service proto; go_package = `<module>/proto/<service_name>` |
| Generated code | Never edit; regenerate with generate_proto.sh |
| Config | Optional; env or future AWS (Secrets Manager, Parameter Store); no GCP |
| Deployment | Docker only; no Kubernetes; future AWS (ECS, EC2, etc.) acceptable |
| Ports | gRPC (e.g. 50051), HTTP for REST + health (e.g. 8080) |
| **Security (P0)** | Auth for gRPC and REST in production; never log secrets, tokens, or full PII |
| **Error handling (P0)** | gRPC: map to status codes; REST: consistent JSON + safe messages; never expose internal errors to clients |
| **Graceful shutdown (P0)** | Handle SIGTERM/SIGINT; drain gRPC and HTTP; exit within timeout (e.g. 30s) |
| **Timeouts (P0)** | Server timeouts for gRPC and HTTP; timeouts on all outbound gRPC/HTTP/DB calls |
| **Health (P0)** | Required `GET /health` or `GET /ready`; 200 when up; optional 503 when dependencies unavailable |

---

## 8. Checklist for New Project from Template

- [ ] Replace module name in `go.mod` and all imports.
- [ ] Rename proto file and service/messages to match new service name; set `go_package` to `<module>/proto/<name>`.
- [ ] Run `proto/generate_proto.sh <name> go`.
- [ ] Organize handler, api, and service by resource: `handler/<resource>.go`, `api/<resource>_handlers.go`, `service/<resource>.go`; remove or replace sample RPCs.
- [ ] Add `api/` layer: api.go (router + Service injection) and grouped resource handlers that call Service; wire API in `cmd/main.go` on HTTP server. Keep gRPC in `handler/` only, REST in `api/` only.
- [ ] Update `cmd/main.go`: gRPC + HTTP (REST + required health), server timeouts, graceful shutdown (SIGTERM/SIGINT, drain, timeout), logging (no GCP, no secrets in logs), ports, optional config/DB/external clients. No Prometheus; observability via logging and future Dynatrace.
- [ ] Update CI/CD: project name, registry, image tag; build and push Docker only (no Kubernetes).
- [ ] Update README: project name, structure tree, gRPC vs REST usage, setup and test instructions.
- [ ] Run `go mod tidy` and fix any broken imports or references.

---

*End of PRD. Use this document as the single source of truth for generating or modifying a Go gRPC service that follows this template.*
