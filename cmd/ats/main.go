package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"

	"suprie/application_tracker/internal/migration"
	"suprie/application_tracker/internal/repository"
	sqliterepo "suprie/application_tracker/internal/repository/sqlite"
	"suprie/application_tracker/internal/service"
)

func main() {
	// Load .env if present; ignore errors (file may not exist).
	_ = godotenv.Load()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "parse-cv":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: ats parse-cv <file.pdf>")
			os.Exit(1)
		}
		service.RunParseCV(os.Args[2])

	case "parse-jd":
		// ats parse-jd <file.pdf> [url]
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: ats parse-jd <file.pdf> [url]")
			os.Exit(1)
		}
		applyURL := ""
		if len(os.Args) >= 4 {
			applyURL = os.Args[3]
		}
		db, repo := openRepo()
		defer db.Close()
		service.RunParseJD(os.Args[2], applyURL, repo)

	case "match":
		// ats match <id> [master_profile]
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: ats match <id> [master_profile]")
			os.Exit(1)
		}
		id := parseID(os.Args[2])
		profilePath := "generated/master_profile.yaml"
		if len(os.Args) >= 4 {
			profilePath = os.Args[3]
		}
		db, repo := openRepo()
		defer db.Close()
		service.RunMatch(profilePath, id, repo)

	case "rank":
			// ats rank <id> [master_profile]
			if len(os.Args) < 3 {
				fmt.Fprintln(os.Stderr, "Usage: ats rank <id> [master_profile]")
				os.Exit(1)
			}
			id := parseID(os.Args[2])
			profilePath := "generated/master_profile.yaml"
			if len(os.Args) >= 4 {
				profilePath = os.Args[3]
			}
			db, repo := openRepo()
			defer db.Close()
			service.RunRank(id, profilePath, repo)

		case "cover-letter":
		// ats cover-letter <id> [master_profile]
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: ats cover-letter <id> [master_profile]")
			os.Exit(1)
		}
		id := parseID(os.Args[2])
		profilePath := "generated/master_profile.yaml"
		if len(os.Args) >= 4 {
			profilePath = os.Args[3]
		}
		db, repo := openRepo()
		defer db.Close()
		service.RunCoverLetter(id, profilePath, repo)

	case "apply":
		// ats apply <id>
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: ats apply <id>")
			os.Exit(1)
		}
		id := parseID(os.Args[2])
		db, repo := openRepo()
		defer db.Close()
		service.RunApply(id, repo)

	case "list":
		// ats list [status]
		statusFilter := ""
		if len(os.Args) >= 3 {
			statusFilter = os.Args[2]
		}
		db, repo := openRepo()
		defer db.Close()
		service.RunList(statusFilter, repo)

	case "detail":
		// ats detail <id>
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: ats detail <id>")
			os.Exit(1)
		}
		id := parseID(os.Args[2])
		db, repo := openRepo()
		defer db.Close()
		service.RunDetail(id, repo)

	case "migrate":
		// ats migrate <up|down>
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: ats migrate <up|down>")
			os.Exit(1)
		}
		runMigrate(os.Args[2])

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

// --- helpers ---

func getDBPath() string {
	if p := os.Getenv("ATS_DB"); p != "" {
		return p
	}
	return "data.db"
}

// openRepo opens the database and returns both the handle and a repository.
// Caller must defer db.Close().
func openRepo() (*sql.DB, repository.JobDescriptionRepository) {
	db, err := sqliterepo.Open(getDBPath())
	if err != nil {
		log.Fatalf("opening database: %v\n(hint: run 'ats migrate up' first, or set ATS_DB)", err)
	}
	return db, sqliterepo.NewJobDescriptionRepository(db)
}

func runMigrate(direction string) {
	db, err := sqliterepo.Open(getDBPath())
	if err != nil {
		log.Fatalf("opening database: %v", err)
	}
	defer db.Close()

	if err := migration.Run(db, direction); err != nil {
		log.Fatalf("migration %s: %v", direction, err)
	}
}

func parseID(raw string) int {
	id, err := strconv.Atoi(raw)
	if err != nil || id <= 0 {
		log.Fatalf("invalid id: %q (must be a positive integer)", raw)
	}
	return id
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  ats parse-cv <file.pdf>")
	fmt.Fprintln(os.Stderr, "  ats parse-jd <file.pdf> [url]")
	fmt.Fprintln(os.Stderr, "  ats match <id> [master_profile]")
	fmt.Fprintln(os.Stderr, "  ats rank <id> [master_profile]")
	fmt.Fprintln(os.Stderr, "  ats cover-letter <id> [master_profile]")
	fmt.Fprintln(os.Stderr, "  ats apply <id>")
	fmt.Fprintln(os.Stderr, "  ats list [status]")
	fmt.Fprintln(os.Stderr, "  ats detail <id>")
	fmt.Fprintln(os.Stderr, "  ats migrate <up|down>")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Statuses: draft, fitmatch, applied, rejected, offer")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Database path: $ATS_DB (default: ./data.db)")
}
