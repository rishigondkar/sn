# Final Verification Report: Table Schemas and Proto Contracts in Service PRDs

**Plan reference:** Table Schemas and Proto Contracts in Service PRDs  
**Verification date:** Final consistency, correctness, and completeness check across all services.

---

## 1. Consistency Checklist (All Services 02–07)

| Criterion | 02 Case | 03 Alert | 04 Enrich | 05 Assign | 06 Attach | 07 Audit |
|-----------|---------|----------|-----------|-----------|-----------|----------|
| **In-body: Owned Data Model → Appendix A** | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| **In-body: Go Project Structure (proto) → Appendix B** | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| **Appendix A: Heading** `# Appendix A: Table schemas (DDL)` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| **Appendix B: Heading** `# Appendix B: Proto contract` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| **Appendix A: Fenced block** ` ```sql ` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| **Appendix B: Fenced block** ` ```protobuf ` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| **Appendix B: Metadata note** (actor/request/correlation via gRPC metadata) | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| **Proto: `syntax = "proto3";`** | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| **Proto: `option go_package = "<module>/proto/<service_name>";`** | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| **Proto: `import "google/protobuf/timestamp.proto";`** (where timestamps used) | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| **List RPCs: request** `page_size`, `page_token` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| **List RPCs: response** `repeated T items`, `next_page_token` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| **UUIDs in proto** as `string` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| **Timestamps in proto** as `google.protobuf.Timestamp` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |

---

## 2. Correctness by Service

### 02 – Case Service
- **Tables:** `cases`, `case_worknotes`. DDL matches Owned Data Model (types, nullability, defaults). `case_worknotes.case_id` REFERENCES `cases(id)` (same-service FK only). No cross-service FKs. No indexes in Owned Data Model → none added. ✓
- **Proto:** CaseService with CreateCase, UpdateCase, AssignCase, CloseCase, AddWorknote, GetCase, ListCases, ListWorknotes. All request/response messages present; Case and Worknote messages align with data model. List responses use `items` and `next_page_token`. ✓

### 03 – Alert & Observable Service
- **Tables:** alert_rules, observables, alerts, case_observables, child_observables, similar_security_incidents. Order respects FKs (alert_rules and observables first; alerts→alert_rules; case_observables/child_observables→observables). `case_id` and `similar_case_id` have no REFERENCES (cross-service). UNIQUE constraints as in model. Indexes match Performance Expectations: case_observables(case_id), case_observables(observable_id), similar_security_incidents(case_id), child_observables(parent_observable_id). ✓
- **Proto:** AlertObservableService with all 14 RPCs (Commands then Queries). FindCasesByObservableResponse: `repeated string case_ids` + `next_page_token`. All list responses use `items`. ✓

### 04 – Enrichment & Threat Lookup Service
- **Tables:** enrichment_results, threat_lookup_results. CHECK constraint `enrichment_results_case_or_observable` (case_id OR observable_id non-null). Indexes align with NFR (case_id, observable_id, source_name, received_at for both tables). No cross-service FKs. ✓
- **Proto:** EnrichmentThreatService with 7 RPCs. Upsert requests carry full payload; list responses use `items` and `next_page_token`. ✓

### 05 – Assignment & Reference Service
- **Tables:** users, assignment_groups, group_members. group_members REFERENCES assignment_groups(id) and users(id). UNIQUE on (group_id, user_id). No indexes explicitly required in PRD → none added. ✓
- **Proto:** AssignmentReferenceService with 6 queries + 6 optional commands. All list responses use `items` and `next_page_token`. ✓

### 06 – Attachment Service
- **Tables:** attachments only. All columns match Owned Data Model; no cross-service FKs. ✓
- **Proto:** AttachmentService with CreateAttachment (bytes content for gateway-mediated upload), ListAttachmentsByCase, DeleteAttachment. ListAttachmentsByCaseResponse: `repeated Attachment items` + `next_page_token`. ✓

### 07 – Audit Service
- **Tables:** audit_events. event_id UNIQUE. Indexes match Indexing Recommendations: case_id, observable_id, (entity_type, entity_id), actor_user_id, correlation_id, occurred_at. ✓
- **Proto:** AuditService with 5 query-only RPCs. List request types include filter params and time range; ListAuditEventsResponse uses `items` and `next_page_token`. AuditEvent message matches audit_events shape. ✓

---

## 3. Completeness

| Document | Appendix A | Appendix B | In-body refs | Notes |
|----------|------------|------------|---------------|--------|
| 02 Case | ✓ DDL for cases, case_worknotes | ✓ case_service.proto, 8 RPCs | ✓ Both | — |
| 03 Alert & Observable | ✓ 6 tables + 4 indexes | ✓ alert_observable_service.proto, 14 RPCs | ✓ Both | — |
| 04 Enrichment & Threat | ✓ 2 tables + CHECK + 8 indexes | ✓ enrichment_threat_service.proto, 7 RPCs | ✓ Both | — |
| 05 Assignment & Reference | ✓ 3 tables | ✓ assignment_reference_service.proto, 12 RPCs | ✓ Both | — |
| 06 Attachment | ✓ attachments | ✓ attachment_service.proto, 3 RPCs | ✓ Both | — |
| 07 Audit | ✓ audit_events + 6 indexes | ✓ audit_service.proto, 5 RPCs | ✓ Both | — |
| 01 Gateway | — | — | ✓ One sentence under Proto Dependencies | Refers to downstream Appendix B |

---

## 4. Plan Compliance Summary

- **Appendix A:** Derived only from Owned Data Model; PostgreSQL types (UUID, TIMESTAMPTZ, VARCHAR/TEXT, INTEGER/BIGINT/NUMERIC, BOOLEAN, JSONB); PRIMARY KEY, NOT NULL, DEFAULT, UNIQUE as in model; indexes only where PRD explicitly mentions (Case: none; Alert: Performance Expectations; Enrichment: NFR; Audit: Indexing Recommendations); no cross-service FKs. ✓
- **Appendix B:** Full proto3 file per service; go_package placeholder; RPC names match gRPC API Contracts; list responses use `items` and `next_page_token`; actor/request/correlation note present. ✓
- **In-body references:** One after Owned Data Model (Appendix A); one in Go Project Structure proto bullet (Appendix B). ✓
- **Gateway (01):** Optional sentence under Proto Dependencies added. ✓

---

## 5. Result

**Consistency:** All six service PRDs (02–07) follow the same structure and conventions.  
**Correctness:** DDL and proto content match Owned Data Models and gRPC API Contracts; no unauthorized fields or RPCs.  
**Completeness:** Every in-scope PRD has Appendix A, Appendix B, and both in-body references; Gateway has the optional Appendix B reference.

**Verification outcome: PASS.** No changes required for plan compliance.
