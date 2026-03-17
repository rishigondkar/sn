// Create service databases. Run from repo root: go run ./scripts/create_dbs.go
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://cursor:qwerty@localhost:5432/postgres?sslmode=disable"
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	dbs := []string{"case_db", "alert_observable", "enrichment_threat", "assignment_ref", "attachments", "audit"}
	for _, db := range dbs {
		_, err := pool.Exec(ctx, "CREATE DATABASE "+db)
		if err != nil {
			if isAlreadyExists(err) {
				fmt.Printf("Database %q already exists, skipping.\n", db)
			} else {
				fmt.Fprintf(os.Stderr, "create %q: %v\n", db, err)
			}
		} else {
			fmt.Printf("Created database %q.\n", db)
		}
	}
}

func isAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "already exists") || strings.Contains(s, "duplicate_database")
}
