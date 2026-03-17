// Run all service migrations. Run from repo root: go run ./scripts/run_migrations.go
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type migration struct {
	db   string
	path string
}

func main() {
	baseURL := os.Getenv("DATABASE_URL")
	if baseURL == "" {
		baseURL = "postgres://cursor:qwerty@localhost:5432/postgres?sslmode=disable"
	}
	// Use postgres for URL template; we'll connect per-db
	baseURL = trimDBFromURL(baseURL)

	migrations := []migration{
		{"case_db", "case-service/migrations/000001_cases.up.sql"},
		{"case_db", "case-service/migrations/000002_case_worknotes.up.sql"},
		{"case_db", "case-service/migrations/000003_case_number_sequence.up.sql"},
		{"case_db", "case-service/migrations/000004_case_incident_fields.up.sql"},
		{"alert_observable", "alert-observable-service/migrations/000001_init_schema.up.sql"},
		{"enrichment_threat", "enrichment-threat-service/migrations/001_initial_schema.up.sql"},
		{"enrichment_threat", "enrichment-threat-service/migrations/002_dedupe_indexes.up.sql"},
		{"assignment_ref", "assignment-reference-service/migrations/001_initial_schema.sql"},
		{"assignment_ref", "assignment-reference-service/migrations/002_seed_poc.sql"},
		{"assignment_ref", "assignment-reference-service/migrations/003_assignment_groups_soc_csirt.sql"},
		{"attachments", "attachment-service/migrations/000001_create_attachments.up.sql"},
		{"audit", "audit-service/migrations/001_audit_events.up.sql"},
	}

	ctx := context.Background()
	for _, m := range migrations {
		url := baseURL + m.db + "?sslmode=disable"
		pool, err := pgxpool.New(ctx, url)
		if err != nil {
			fmt.Fprintf(os.Stderr, "connect %s: %v\n", m.db, err)
			os.Exit(1)
		}
		sql, err := os.ReadFile(m.path)
		if err != nil {
			pool.Close()
			fmt.Fprintf(os.Stderr, "read %s: %v\n", m.path, err)
			os.Exit(1)
		}
		_, err = pool.Exec(ctx, string(sql))
		pool.Close()
		if err != nil {
			msg := err.Error()
			if strings.Contains(msg, "42P07") || strings.Contains(msg, "42710") || strings.Contains(msg, "already exists") {
				fmt.Printf("SKIP %s <- %s (already applied)\n", m.db, filepath.Base(m.path))
			} else {
				fmt.Fprintf(os.Stderr, "migrate %s %s: %v\n", m.db, m.path, err)
				os.Exit(1)
			}
		} else {
			fmt.Printf("OK %s <- %s\n", m.db, filepath.Base(m.path))
		}
	}
	fmt.Println("All migrations completed.")
}

func trimDBFromURL(url string) string {
	// Remove /dbname and ?... then we'll append /db?sslmode=disable
	end := len(url)
	if i := indexAny(url, "?"); i >= 0 {
		end = i
	}
	path := url
	if i := indexLast(url[:end], "/"); i >= 0 {
		path = url[:i+1]
	}
	return path
}

func indexAny(s string, sep string) int {
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			return i
		}
	}
	return -1
}

func indexLast(s string, sep string) int {
	for i := len(s) - len(sep); i >= 0; i-- {
		if s[i:i+len(sep)] == sep {
			return i
		}
	}
	return -1
}
