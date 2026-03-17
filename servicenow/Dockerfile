# Build all Go services
FROM golang:1.24-alpine AS gobuilder
WORKDIR /workspace
COPY . .

RUN cd case-service && CGO_ENABLED=0 go build -o /bin/case-service ./cmd
RUN cd alert-observable-service && CGO_ENABLED=0 go build -o /bin/alert-observable-service ./cmd
RUN cd enrichment-threat-service && CGO_ENABLED=0 go build -o /bin/enrichment-threat-service ./cmd
RUN cd assignment-reference-service && CGO_ENABLED=0 go build -o /bin/assignment-reference-service ./cmd
RUN cd attachment-service && CGO_ENABLED=0 go build -o /bin/attachment-service ./cmd
RUN cd audit-service && CGO_ENABLED=0 go build -o /bin/audit-service ./cmd
RUN cd api-gateway-bff && CGO_ENABLED=0 go build -o /bin/api-gateway ./cmd

# Build frontend
FROM node:20-alpine AS frontend
WORKDIR /app
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Runtime: Caddy + all binaries + frontend + postgres client for migrations
FROM alpine:3.19
RUN apk add --no-cache ca-certificates caddy postgresql-client curl
WORKDIR /app
COPY --from=gobuilder /bin/case-service /bin/alert-observable-service /bin/enrichment-threat-service \
  /bin/assignment-reference-service /bin/attachment-service /bin/audit-service /bin/api-gateway /app/bin/
COPY --from=frontend /app/dist ./frontend
COPY case-service/migrations ./migrations/case-service
COPY alert-observable-service/migrations ./migrations/alert-observable-service
COPY enrichment-threat-service/migrations ./migrations/enrichment-threat-service
COPY assignment-reference-service/migrations ./migrations/assignment-reference-service
COPY attachment-service/migrations ./migrations/attachment-service
COPY audit-service/migrations ./migrations/audit-service
COPY scripts/run-migrations.sh scripts/run-assignment-ref-only.sh scripts/create_dbs.sql ./scripts/
RUN chmod +x ./scripts/run-migrations.sh ./scripts/run-assignment-ref-only.sh
COPY Caddyfile start.sh ./
RUN chmod +x start.sh
EXPOSE 80
ENTRYPOINT ["./start.sh"]
