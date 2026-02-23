package app

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/aussiebroadwan/taboo/internal/config"
	"github.com/aussiebroadwan/taboo/internal/store/drivers/sqlite"
	"github.com/golang-migrate/migrate/v4"
)

// RunMigrate runs the migrate subcommand.
func RunMigrate(configPath string, args []string) error {
	if len(args) == 0 {
		printMigrateUsage()
		return nil
	}

	// Load config for database DSN
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Open database
	db, err := sqlite.OpenDB(cfg.Database.DSN)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer db.Close()

	// Create migrate instance
	m, err := sqlite.NewMigrate(db)
	if err != nil {
		return fmt.Errorf("creating migrator: %w", err)
	}

	switch args[0] {
	case "up":
		return runMigrateUp(m, args[1:])
	case "down":
		return runMigrateDown(m, args[1:])
	case "status":
		return runMigrateStatus(m)
	default:
		fmt.Fprintf(os.Stderr, "unknown migrate command: %s\n\n", args[0])
		printMigrateUsage()
		return nil
	}
}

func runMigrateUp(m *migrate.Migrate, args []string) error {
	if len(args) > 0 {
		// Apply N migrations
		n, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid number of migrations: %w", err)
		}
		if err := m.Steps(n); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("applying migrations: %w", err)
		}
		fmt.Printf("Applied %d migration(s)\n", n)
	} else {
		// Apply all migrations
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("applying migrations: %w", err)
		}
		fmt.Println("Applied all pending migrations")
	}

	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return fmt.Errorf("getting version: %w", err)
	}
	fmt.Printf("Current version: %d (dirty: %t)\n", version, dirty)

	return nil
}

func runMigrateDown(m *migrate.Migrate, args []string) error {
	n := 1
	if len(args) > 0 {
		var err error
		n, err = strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid number of migrations: %w", err)
		}
	}

	if err := m.Steps(-n); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("rolling back migrations: %w", err)
	}
	fmt.Printf("Rolled back %d migration(s)\n", n)

	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return fmt.Errorf("getting version: %w", err)
	}
	fmt.Printf("Current version: %d (dirty: %t)\n", version, dirty)

	return nil
}

func runMigrateStatus(m *migrate.Migrate) error {
	version, dirty, err := m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		fmt.Println("No migrations have been applied")
		return nil
	}
	if err != nil {
		return fmt.Errorf("getting version: %w", err)
	}

	fmt.Printf("Current version: %d\n", version)
	if dirty {
		fmt.Println("Status: DIRTY (migration failed, manual intervention required)")
	} else {
		fmt.Println("Status: clean")
	}

	return nil
}

func printMigrateUsage() {
	fmt.Fprintf(os.Stderr, `taboo migrate - Database migration management

Usage:
  taboo migrate <command> [arguments]

Commands:
  up [N]      Apply all pending migrations, or N migrations if specified
  down [N]    Roll back N migrations (default: 1)
  status      Show current migration version and state

Examples:
  taboo migrate up                Apply all pending migrations
  taboo migrate up 1              Apply 1 migration
  taboo migrate down              Roll back 1 migration
  taboo migrate down 2            Roll back 2 migrations
  taboo migrate status            Show migration status
`)
	flag.PrintDefaults()
}
