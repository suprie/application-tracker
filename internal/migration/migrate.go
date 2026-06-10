package migration

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type migration struct {
	version uint
	name    string
	upSQL   string
	downSQL string
}

// Run applies pending migrations (up) or rolls back the last one (down).
func Run(db *sql.DB, direction string) error {
	drv, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("creating migrate driver: %w", err)
	}

	migrations, err := parseMigrationFiles()
	if err != nil {
		return err
	}

	switch direction {
	case "up":
		return runUp(drv, migrations)
	case "down":
		return runDown(drv, migrations)
	default:
		return fmt.Errorf("unknown migration direction: %s (use up or down)", direction)
	}
}

func parseMigrationFiles() ([]migration, error) {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("reading embedded migrations: %w", err)
	}

	grouped := make(map[uint]*migration)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Expected format: 000001_name.up.sql or 000001_name.down.sql
		parts := strings.SplitN(entry.Name(), "_", 3)
		if len(parts) < 3 {
			return nil, fmt.Errorf("invalid migration filename: %s", entry.Name())
		}

		version, err := strconv.ParseUint(parts[0], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid migration version in %s: %w", entry.Name(), err)
		}

		v := uint(version)
		name := strings.TrimSuffix(parts[2], ".up.sql")
		name = strings.TrimSuffix(name, ".down.sql")

		if _, ok := grouped[v]; !ok {
			grouped[v] = &migration{version: v, name: name}
		}

		content, err := migrationsFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", entry.Name(), err)
		}

		if strings.HasSuffix(entry.Name(), ".up.sql") {
			grouped[v].upSQL = string(content)
		} else {
			grouped[v].downSQL = string(content)
		}
	}

	result := make([]migration, 0, len(grouped))
	for _, m := range grouped {
		if m.upSQL == "" {
			return nil, fmt.Errorf("migration %d_%s is missing .up.sql file", m.version, m.name)
		}
		result = append(result, *m)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].version < result[j].version
	})

	return result, nil
}

func runUp(drv database.Driver, migrations []migration) error {
	currentVer, dirty, err := drv.Version()
	if err != nil {
		return fmt.Errorf("reading current version: %w", err)
	}
	if dirty {
		return fmt.Errorf("database is in dirty state at version %d; fix manually before migrating", currentVer)
	}

	for _, m := range migrations {
		if currentVer >= 0 && m.version <= uint(currentVer) {
			fmt.Printf("Skipping %d_%s (already applied)\n", m.version, m.name)
			continue
		}

		fmt.Printf("Applying %d_%s...\n", m.version, m.name)
		if err := drv.Run(strings.NewReader(m.upSQL)); err != nil {
			return fmt.Errorf("migration %d_%s failed: %w", m.version, m.name, err)
		}

		if err := drv.SetVersion(int(m.version), false); err != nil {
			return fmt.Errorf("recording version %d: %w", m.version, err)
		}

		fmt.Printf("Applied %d_%s\n", m.version, m.name)
	}

	return nil
}

func runDown(drv database.Driver, migrations []migration) error {
	currentVer, dirty, err := drv.Version()
	if err != nil {
		return fmt.Errorf("reading current version: %w", err)
	}
	if dirty {
		return fmt.Errorf("database is in dirty state at version %d; fix manually before rolling back", currentVer)
	}
	if currentVer <= 0 {
		fmt.Println("No migrations to roll back.")
		return nil
	}

	var lastMigration *migration
	for _, m := range migrations {
		if m.version == uint(currentVer) {
			lastMigration = &m
			break
		}
	}

	if lastMigration == nil {
		return fmt.Errorf("migration version %d is applied but no .down.sql file found", currentVer)
	}
	if lastMigration.downSQL == "" {
		return fmt.Errorf("migration %d_%s has no .down.sql file", lastMigration.version, lastMigration.name)
	}

	fmt.Printf("Rolling back %d_%s...\n", lastMigration.version, lastMigration.name)
	if err := drv.Run(strings.NewReader(lastMigration.downSQL)); err != nil {
		return fmt.Errorf("rolling back %d_%s failed: %w", lastMigration.version, lastMigration.name, err)
	}

	// Set version to the previous migration, or -1 if this was the first.
	prevVersion := -1
	for _, m := range migrations {
		if m.version < uint(currentVer) && int(m.version) > prevVersion {
			prevVersion = int(m.version)
		}
	}
	if err := drv.SetVersion(prevVersion, false); err != nil {
		return fmt.Errorf("recording version %d: %w", prevVersion, err)
	}

	fmt.Printf("Rolled back %d_%s\n", lastMigration.version, lastMigration.name)
	return nil
}
