# In-Depth API Test Plan

**Target:** API Gateway BFF at `http://localhost:8080` (override with `BASE_URL`).  
**Prerequisites:** All services running (case, alert-observable, enrichment-threat, assignment-reference, attachment, audit, api-gateway-bff). Migrations applied.

**Auth:** All mutating and scoped reads use header `X-User-Id` (and optionally `X-User-Name`). Tests send these headers.

---

## 1. Health & Liveness

| ID   | Description              | Method | Path     | Expected | Assertions                    |
|------|--------------------------|--------|----------|----------|-------------------------------|
| H01  | Health returns 200       | GET    | /health  | 200      | Status 200                    |

---

## 2. Case Lifecycle (Core)

| ID   | Description                | Method | Path              | Expected | Assertions                                      |
|------|----------------------------|--------|-------------------|----------|-------------------------------------------------|
| C01  | Create case (minimal)      | POST   | /api/v1/cases     | 201      | Body has `id`, `case_number`, `title`, `state`  |
| C02  | Create case (full)         | POST   | /api/v1/cases     | 201      | `priority`, `description` in body               |
| C03  | Get case by ID             | GET    | /api/v1/cases/{id} | 200      | Same `id`, `case_number`                        |
| C04  | Update case (PATCH)        | PATCH  | /api/v1/cases/{id} | 200     | Updated `title` / `priority` in response        |
| C05  | Get non-existent case      | GET    | /api/v1/cases/{uuid} | 404/500/200 | 404 (real), 500 (downstream), or 200 (stub fake case) |

---

## 3. Validation & Error Envelope

| ID   | Description                | Method | Path          | Expected | Assertions                          |
|------|----------------------------|--------|---------------|----------|-------------------------------------|
| V01  | Create case missing title  | POST   | /api/v1/cases | 400      | `VALIDATION_ERROR`, details field   |
| V02  | Create case invalid JSON   | POST   | /api/v1/cases | 400      | 400, invalid JSON message          |
| V03  | Get case empty caseId      | GET    | (route N/A)   | 400      | Validation error for caseId         |
| V04  | Add worknote empty content | POST   | .../worknotes | 400      | Validation error for content        |
| V05  | Create attachment missing  | POST   | /api/v1/attachments | 400 | case_id or file_name required       |

---

## 4. Worknotes

| ID   | Description           | Method | Path                    | Expected | Assertions                    |
|------|-----------------------|--------|-------------------------|----------|-------------------------------|
| W01  | Add worknote          | POST   | /api/v1/cases/{id}/worknotes | 201 | Body has `id`, `content`      |
| W02  | List worknotes        | GET    | /api/v1/cases/{id}/worknotes | 200 | `worknotes` array, optional next_page_token |

---

## 5. Assign & Close

| ID   | Description      | Method | Path                    | Expected | Assertions        |
|------|------------------|--------|-------------------------|----------|-------------------|
| A01  | Assign case      | POST   | /api/v1/cases/{id}/assign | 204    | No content        |
| A02  | Close case       | POST   | /api/v1/cases/{id}/close  | 204    | No content        |

---

## 6. Attachments

| ID   | Description           | Method | Path                  | Expected | Assertions                          |
|------|-----------------------|--------|-----------------------|----------|-------------------------------------|
| AT01 | Create attachment     | POST   | /api/v1/attachments   | 201      | Body has `id`, `case_id`, `file_name` |
| AT02 | List by case          | GET    | /api/v1/cases/{id}/attachments | 200 | `attachments` array                 |

---

## 7. Observables & Enrichment

| ID   | Description              | Method | Path                          | Expected | Assertions                |
|------|--------------------------|--------|-------------------------------|----------|---------------------------|
| O01  | Link observable          | POST   | /api/v1/cases/{id}/observables | 204    | No content                |
| O02  | List case observables    | GET    | /api/v1/cases/{id}/observables | 200    | `observables` array       |
| O03  | List case alerts         | GET    | /api/v1/cases/{id}/alerts      | 200    | `alerts` array            |
| O04  | List enrichment results  | GET    | /api/v1/cases/{id}/enrichment-results | 200 | `enrichment_results` array |
| O05  | Threat lookups by observable | GET | /api/v1/observables/{obsId}/threat-lookups | 200 | `threat_lookups` array |

---

## 8. Reference Data

| ID   | Description   | Method | Path                     | Expected | Assertions        |
|------|---------------|--------|--------------------------|----------|-------------------|
| R01  | List users    | GET    | /api/v1/reference/users   | 200      | `users` array     |
| R02  | List groups   | GET    | /api/v1/reference/groups  | 200      | `groups` array    |

---

## 9. Audit & Aggregated Detail

| ID   | Description          | Method | Path                          | Expected | Assertions              |
|------|----------------------|--------|-------------------------------|----------|-------------------------|
| AU01 | List audit events    | GET    | /api/v1/cases/{id}/audit-events | 200   | `audit_events` array    |
| D01  | Get case detail      | GET    | /api/v1/cases/{id}/detail       | 200   | `case` object, optional sections |

---

## 10. Pagination (Optional)

| ID   | Description     | Method | Path                    | Expected | Assertions                    |
|------|-----------------|--------|-------------------------|----------|-------------------------------|
| P01  | List with size  | GET    | ...?page_size=2         | 200      | At most 2 items or next token |
| P02  | List with token | GET    | ...?page_token=...      | 200      | 200 or 400 for invalid token  |

---

## Execution

- Run: `./tests/run.sh` (or `bash tests/run.sh`) from repo root.
- Environment: `BASE_URL` (default `http://localhost:8080`), optional `X-User-Id` (default `e2e-test-user`).
- Output: Per-test pass/fail and final summary (passed / failed / total).
