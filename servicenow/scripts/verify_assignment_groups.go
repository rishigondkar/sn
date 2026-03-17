// One-off: verify assignment_groups has rows. Run from repo root: go run ./scripts/verify_assignment_groups.go
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	url := "postgres://cursor:qwerty@localhost:5432/assignment_ref?sslmode=disable"
	if v := os.Getenv("DATABASE_URL"); v != "" {
		// Use same host as DATABASE_URL but db name assignment_ref
		if idx := strings.Index(v, "?"); idx >= 0 {
			v = v[:idx]
		}
		if i := strings.LastIndex(v, "/"); i >= 0 {
			url = v[:i+1] + "assignment_ref?sslmode=disable"
		}
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()
	var n int
	err = pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM assignment_groups").Scan(&n)
	if err != nil {
		fmt.Fprintf(os.Stderr, "query: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("assignment_groups count: %d\n", n)
	if n == 0 {
		fmt.Fprintln(os.Stderr, "No rows - run migrations: go run ./scripts/run_migrations.go")
		os.Exit(1)
	}
}
