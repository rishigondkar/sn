# Start all services

Run each command in its **own terminal**. Order: start domain services first (1–6), then the gateway (7). Optionally start the frontend (8).

From the repo root: `/Users/rishigondkar/Personal/servicenow` (or `cd` there first).

---

## Terminal 1 – Case Service (gRPC 50051)

```bash
cd case-service && AUDIT_SERVICE_ADDR=localhost:50056 go run ./cmd
```

---

## Terminal 2 – Alert & Observable Service (gRPC 50052)

```bash
cd alert-observable-service && go run ./cmd
```

---

## Terminal 3 – Enrichment & Threat Lookup Service (gRPC 50053)

```bash
cd enrichment-threat-service && go run ./cmd
```

---

## Terminal 4 – Assignment & Reference Service (gRPC 50054, HTTP 8081)

```bash
cd assignment-reference-service && HTTP_ADDR=:8081 go run ./cmd
```

---

## Terminal 5 – Attachment Service (gRPC 50055)

```bash
cd attachment-service && go run ./cmd
```

---

## Terminal 6 – Audit Service (gRPC 50056)

```bash
cd audit-service && go run ./cmd
```

---

## Terminal 7 – API Gateway (HTTP 8080)

```bash
cd api-gateway-bff && go run ./cmd
```

---

## Terminal 8 (optional) – Frontend

```bash
cd frontend && npm run dev
```

---

**Note:** If you're not at the repo root, either run from there or use full paths, e.g.  
`cd /Users/rishigondkar/Personal/servicenow/case-service && AUDIT_SERVICE_ADDR=localhost:50056 go run ./cmd`
