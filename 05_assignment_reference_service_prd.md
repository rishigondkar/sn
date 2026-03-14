# Assignment & Reference Service PRD
Version: 1.0  
Service name: assignment-reference-service  
Language: Golang  
Primary responsibility: source of truth for users, assignment groups, group membership, and optionally reference enumerations

## Mission
Provide a consistent source for user and group identity references used by the platform. Support ID resolution, list queries for UI pickers, and validation of assignment-related references during writes in other services.


# Agent Execution Policy

This PRD is written for LLMs and coding agents. Treat this section as binding.

## Workflow Orchestration
### 1. Plan Mode by Default
- Enter plan mode for any non-trivial task, especially anything with 3 or more steps, architectural choices, schema changes, interface changes, or migrations.
- If something goes sideways, stop and re-plan immediately. Do not keep pushing a broken approach.
- Use plan mode for verification steps too, not just implementation.
- Write detailed specs up front to reduce ambiguity before changing code.

### 2. Subagent Strategy
- Use subagents liberally to keep the main context window clean.
- Offload research, code reading, dependency analysis, and parallel investigation to subagents.
- For complex problems, use multiple focused subagents in parallel.
- One task per subagent. Keep each subagent narrowly scoped.

### 3. Self-Improvement Loop
- After any user correction, update `tasks/lessons.md`.
- Write a rule that would have prevented the mistake.
- Review `tasks/lessons.md` at the start of any relevant future session.
- Keep reducing repeated mistakes; do not repeat the same error pattern.

### 4. Verification Before Done
- Never mark a task complete without proving it works.
- Compare before and after behavior when relevant.
- Ask: "Would a staff engineer approve this?"
- Run tests, inspect logs, validate edge cases, and demonstrate correctness.

### 5. Demand Elegance, But Stay Balanced
- For non-trivial changes, pause and ask whether there is a more elegant solution.
- If the current fix feels hacky, step back and implement the cleaner design.
- Do not over-engineer simple fixes.
- Challenge your own work before presenting it.

### 6. Autonomous Bug Fixing
- When given a bug report, fix it. Do not ask the user to do exploratory debugging for you.
- Start from logs, errors, stack traces, and failing tests.
- Minimize user context switching.
- If CI is failing, investigate and repair the failures without being told exactly how.

## Task Management Rules
1. Create `tasks/todo.md` before implementation with checkable items.
2. Check in against the plan before beginning implementation.
3. Mark items complete as work progresses.
4. Add high-level summaries of changes as work proceeds.
5. Add a review/results section to `tasks/todo.md` before concluding.
6. Update `tasks/lessons.md` after corrections or notable mistakes.

## Core Engineering Principles
- **Simplicity First:** implement the simplest design that correctly satisfies the requirements.
- **No Laziness:** fix root causes; do not ship temporary hacks as final solutions.
- **Minimal Impact:** change only what is necessary and preserve existing behavior unless the PRD explicitly requires otherwise.

## Global Do-Not-Do List
- Do not invent APIs, events, or schema fields that are not specified in this PRD or the shared platform contract.
- Do not silently change shared contracts.
- Do not couple directly to another service's database.
- Do not bypass the event bus for audit publication.
- Do not mark a task done if tests are missing or failing.
- Do not skip idempotency, validation, error handling, or observability.
- Do not create hidden behavior such as implicit external lookups unless explicitly required.
- Do not use ad hoc JSON blobs where strongly typed fields are required by this PRD.
- Do not expose internal gRPC APIs directly to the public internet.
- Do not break backward compatibility in REST or gRPC without updating the shared contract and migration guidance.

## Golang Standards
- Language: Go 1.22+.
- Prefer standard library first; add third-party dependencies only with clear justification.
- Use context propagation on all inbound and outbound calls.
- Use structured logging.
- Use table-driven tests.
- Use dependency injection via interfaces where it improves testability.
- Separate transport, service, repository, and domain logic layers.
- Validate all request payloads before entering business logic.
- Apply timeouts to all outbound network calls.
- Propagate correlation IDs, request IDs, and actor metadata.


# In Scope
- user records
- assignment groups
- group memberships
- list/search users
- list/search groups
- validate references
- optional lookup tables for enums if the platform chooses managed references

# Out of Scope
- auth provider implementation
- authorization policy engine for all services
- case lifecycle
- observables
- audit event storage
- attachments

# Owned Data Model

## `users`
- `id` UUID PK
- `username` VARCHAR(100) unique not null
- `email` VARCHAR(255) unique not null
- `display_name` VARCHAR(255) not null
- `is_active` BOOLEAN not null default true
- `created_at` TIMESTAMPTZ not null
- `updated_at` TIMESTAMPTZ not null

## `assignment_groups`
- `id` UUID PK
- `group_name` VARCHAR(255) unique not null
- `description` TEXT nullable
- `is_active` BOOLEAN not null default true
- `created_at` TIMESTAMPTZ not null
- `updated_at` TIMESTAMPTZ not null

## `group_members`
- `id` UUID PK
- `group_id` UUID not null
- `user_id` UUID not null
- `member_role` VARCHAR(50) nullable
- `created_at` TIMESTAMPTZ not null

Recommended uniqueness:
- unique on (`group_id`, `user_id`)

Executable PostgreSQL DDL for these tables: see **Appendix A: Table schemas (DDL)**.

# Business Rules
- inactive users and groups remain queryable when historical references require display expansion
- list/search APIs should allow active-only filter by default
- membership duplicates not allowed
- deleting users/groups physically is discouraged; prefer deactivation for historical consistency
- For POC, seed at least one user and one assignment group (for example via migration or admin script) so Case and Alert services can reference them; document seeding steps in this service's README.

# gRPC API Contracts

## Queries
- `GetUser`
- `ListUsers`
- `GetGroup`
- `ListGroups`
- `ListGroupMembers`
- `ValidateUserExists`
- `ValidateGroupExists`

## Commands
If write administration is supported in v1:
- `CreateUser`
- `UpdateUser`
- `CreateGroup`
- `UpdateGroup`
- `AddGroupMember`
- `RemoveGroupMember`

If administrative writes are out of scope, implement queries only and seed data via migrations or admin scripts.

# REST API Surface
Likely served via API Gateway to support UI selectors:
- `GET /api/v1/reference/users`
- `GET /api/v1/reference/groups`
- `GET /api/v1/reference/groups/{groupId}/members`

# Integration Dependencies
None required for core behavior.
This service should remain low-coupling and easy to cache by consumers.

# Audit Events To Publish
Only if write administration exists:
- `user.updated`
- `group.updated`

# Validation Rules
- email format validation
- username uniqueness
- group name uniqueness
- prevent group membership duplication

# Query Expectations
- search users by display name, username, email
- search groups by name
- active-only filter
- pagination support

# Non-Functional Requirements
- low latency query service
- cache-friendly responses
- support historical expansion of inactive references

# Implementation Guidance for Agents
1. Keep this service simple and stable.
2. Prefer read-mostly optimization.
3. If administrative writes are not needed for v1, keep the surface minimal.
4. Ensure response models are suitable for dropdowns and reference validation.

# Go Project Structure

Follow the canonical template: [PRD_GO_GRPC_SERVICE_TEMPLATE.md](../golang_prd/PRD_GO_GRPC_SERVICE_TEMPLATE.md). Full layout; query-heavy with optional command RPCs and optional REST for reference routes.

## Repository Layout

```
<project-root>/
├── cmd/
│   └── main.go
├── handler/
│   ├── handler.go
│   ├── handler_test.go
│   ├── user.go
│   └── group.go
├── api/
│   ├── api.go
│   ├── user_handlers.go
│   └── group_handlers.go
├── service/
│   ├── service.go
│   ├── user.go
│   └── group.go
├── repository/
│   └── repository.go
├── config/
│   └── config.go
├── logging/
│   └── logging.go
├── proto/
│   ├── assignment_reference_service.proto
│   ├── generate_proto.sh
│   └── assignment_reference_service/
│       ├── assignment_reference_service.pb.go
│       └── assignment_reference_service_grpc.pb.go
├── go.mod
├── go.sum
├── Dockerfile
├── .gitlab-ci.yml
└── README.md
```

- `handler/`: gRPC for all Query RPCs (GetUser, ListUsers, GetGroup, ListGroups, ListGroupMembers, ValidateUserExists, ValidateGroupExists) and optional Command RPCs if write administration is in scope.
- `api/`: Optional REST for `/api/v1/reference/users`, `/api/v1/reference/groups`, `/api/v1/reference/groups/{groupId}/members` (e.g. when called via gateway or direct).
- `service/` and `repository/`: Read-mostly; users, assignment_groups, group_members; cache-friendly query paths.
- `proto/assignment_reference_service.proto`: `go_package = "<module>/proto/assignment_reference_service"`; define Queries and optional Commands per gRPC API Contracts above. Full proto definition including request/response messages: see **Appendix B: Proto contract**.

## Module, Ports, Proto, P0

| Item | Value |
|------|--------|
| Module | `github.com/<org>/assignment-reference-service` |
| Proto | `assignment_reference_service.proto`; go_package `<module>/proto/assignment_reference_service` |
| gRPC port | **50051** |
| HTTP port | **8080** (health + optional REST) |
| Go version | 1.22+ (platform contract); 1.23 in go.mod if using template |

**P0:** Health endpoint; graceful shutdown; gRPC (and REST if present) error mapping; timeouts; no secrets in logs. If write admin exists: publish user.updated, group.updated per contract.

## Agent Checklist

- [ ] Set module name and proto go_package; run proto generation.
- [ ] Implement repository (users, assignment_groups, group_members); migrations; read-optimized indexes.
- [ ] Implement service and handler (queries + optional commands); optional api for reference routes.
- [ ] main.go: gRPC + HTTP, timeouts, graceful shutdown.
- [ ] Tests: query filters, validation, pagination; optional command tests.
- [ ] Update README.

# Do Not Do
- Do not embed case-specific business logic.
- Do not become the auth provider.
- Do not hard-delete records that may be referenced historically.

---

# Appendix A: Table schemas (DDL)

```sql
CREATE TABLE users (
  id UUID PRIMARY KEY,
  username VARCHAR(100) NOT NULL UNIQUE,
  email VARCHAR(255) NOT NULL UNIQUE,
  display_name VARCHAR(255) NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE assignment_groups (
  id UUID PRIMARY KEY,
  group_name VARCHAR(255) NOT NULL UNIQUE,
  description TEXT,
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE group_members (
  id UUID PRIMARY KEY,
  group_id UUID NOT NULL REFERENCES assignment_groups(id),
  user_id UUID NOT NULL REFERENCES users(id),
  member_role VARCHAR(50),
  created_at TIMESTAMPTZ NOT NULL,
  UNIQUE (group_id, user_id)
);
```

---

# Appendix B: Proto contract

Actor and request/correlation ids are passed via gRPC metadata per platform contract; they are not repeated in request messages below.

```protobuf
syntax = "proto3";

package assignment_reference_service;

import "google/protobuf/timestamp.proto";

option go_package = "<module>/proto/assignment_reference_service";

service AssignmentReferenceService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
  rpc GetGroup(GetGroupRequest) returns (GetGroupResponse);
  rpc ListGroups(ListGroupsRequest) returns (ListGroupsResponse);
  rpc ListGroupMembers(ListGroupMembersRequest) returns (ListGroupMembersResponse);
  rpc ValidateUserExists(ValidateUserExistsRequest) returns (ValidateUserExistsResponse);
  rpc ValidateGroupExists(ValidateGroupExistsRequest) returns (ValidateGroupExistsResponse);
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
  rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse);
  rpc CreateGroup(CreateGroupRequest) returns (CreateGroupResponse);
  rpc UpdateGroup(UpdateGroupRequest) returns (UpdateGroupResponse);
  rpc AddGroupMember(AddGroupMemberRequest) returns (AddGroupMemberResponse);
  rpc RemoveGroupMember(RemoveGroupMemberRequest) returns (RemoveGroupMemberResponse);
}

message GetUserRequest {
  string id = 1;
}

message GetUserResponse {
  User user = 1;
}

message ListUsersRequest {
  int32 page_size = 1;
  string page_token = 2;
  optional bool active_only = 3;
  optional string filter_display_name = 4;
  optional string filter_username = 5;
  optional string filter_email = 6;
}

message ListUsersResponse {
  repeated User items = 1;
  string next_page_token = 2;
}

message GetGroupRequest {
  string id = 1;
}

message GetGroupResponse {
  Group group = 1;
}

message ListGroupsRequest {
  int32 page_size = 1;
  string page_token = 2;
  optional bool active_only = 3;
  optional string filter_group_name = 4;
}

message ListGroupsResponse {
  repeated Group items = 1;
  string next_page_token = 2;
}

message ListGroupMembersRequest {
  string group_id = 1;
  int32 page_size = 2;
  string page_token = 3;
}

message ListGroupMembersResponse {
  repeated GroupMember items = 1;
  string next_page_token = 2;
}

message ValidateUserExistsRequest {
  string user_id = 1;
}

message ValidateUserExistsResponse {
  bool exists = 1;
}

message ValidateGroupExistsRequest {
  string group_id = 1;
}

message ValidateGroupExistsResponse {
  bool exists = 1;
}

message CreateUserRequest {
  string username = 1;
  string email = 2;
  string display_name = 3;
  bool is_active = 4;
}

message CreateUserResponse {
  User user = 1;
}

message UpdateUserRequest {
  string id = 1;
  optional string username = 2;
  optional string email = 3;
  optional string display_name = 4;
  optional bool is_active = 5;
}

message UpdateUserResponse {
  User user = 1;
}

message CreateGroupRequest {
  string group_name = 1;
  string description = 2;
  bool is_active = 3;
}

message CreateGroupResponse {
  Group group = 1;
}

message UpdateGroupRequest {
  string id = 1;
  optional string group_name = 2;
  optional string description = 3;
  optional bool is_active = 4;
}

message UpdateGroupResponse {
  Group group = 1;
}

message AddGroupMemberRequest {
  string group_id = 1;
  string user_id = 2;
  string member_role = 3;
}

message AddGroupMemberResponse {
  GroupMember member = 1;
}

message RemoveGroupMemberRequest {
  string group_id = 1;
  string user_id = 2;
}

message RemoveGroupMemberResponse {}

message User {
  string id = 1;
  string username = 2;
  string email = 3;
  string display_name = 4;
  bool is_active = 5;
  google.protobuf.Timestamp created_at = 6;
  google.protobuf.Timestamp updated_at = 7;
}

message Group {
  string id = 1;
  string group_name = 2;
  string description = 3;
  bool is_active = 4;
  google.protobuf.Timestamp created_at = 5;
  google.protobuf.Timestamp updated_at = 6;
}

message GroupMember {
  string id = 1;
  string group_id = 2;
  string user_id = 3;
  string member_role = 4;
  google.protobuf.Timestamp created_at = 5;
}
```
