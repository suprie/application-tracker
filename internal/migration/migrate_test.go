package migration

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dbPath := t.TempDir() + "/test.db"
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.Ping(); err != nil {
		t.Fatalf("ping test db: %v", err)
	}

	return db
}

func TestRun_Up_Down(t *testing.T) {
	db := openTestDB(t)

	// Apply all migrations.
	if err := Run(db, "up"); err != nil {
		t.Fatalf("up: %v", err)
	}

	// Verify the table exists.
	var tableName string
	err := db.QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND name='job_descriptions'",
	).Scan(&tableName)
	if err != nil {
		t.Fatalf("query job_descriptions table: %v", err)
	}
	if tableName != "job_descriptions" {
		t.Fatalf("expected job_descriptions table, got %q", tableName)
	}

	// Roll back all migrations (one per loop; stop after a safe upper bound).
	for i := 0; i < 10; i++ {
		var version int
		if err := db.QueryRow("SELECT version FROM schema_migrations LIMIT 1").Scan(&version); err != nil {
			break // table empty or doesn't exist
		}
		if version < 0 {
			break
		}
		if err := Run(db, "down"); err != nil {
			t.Fatalf("down: %v", err)
		}
	}

	// Verify the table is gone.
	err = db.QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND name='job_descriptions'",
	).Scan(&tableName)
	if err == nil {
		t.Fatal("expected job_descriptions table to be dropped, but it still exists")
	}
}

func TestRun_Up_Idempotent(t *testing.T) {
	db := openTestDB(t)

	// Run up twice — second call should be a no-op.
	if err := Run(db, "up"); err != nil {
		t.Fatalf("first up: %v", err)
	}
	if err := Run(db, "up"); err != nil {
		t.Fatalf("second up: %v", err)
	}
}

func TestRun_Down_WhenNone(t *testing.T) {
	db := openTestDB(t)

	// Down with no migrations applied should succeed.
	if err := Run(db, "down"); err != nil {
		t.Fatalf("down on empty db: %v", err)
	}
}

func TestRun_InvalidDirection(t *testing.T) {
	db := openTestDB(t)

	err := Run(db, "sideways")
	if err == nil {
		t.Fatal("expected error for invalid direction, got nil")
	}
}
