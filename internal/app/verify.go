package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/aussiebroadwan/taboo/internal/config"
	"github.com/aussiebroadwan/taboo/internal/store/drivers/sqlite"
	"github.com/aussiebroadwan/taboo/pkg/lint"
	"github.com/golang-migrate/migrate/v4"
)

// RunVerify runs the verify subcommand.
func RunVerify(configPath string) error {
	c := lint.NewCollector()

	// Step 1: Load and validate configuration
	cfg, cfgErr := config.Load(configPath)
	if cfgErr != nil {
		// Check if it's a validation error (lint.Issues)
		var lintIssues lint.Issues
		if errors.As(cfgErr, &lintIssues) {
			c.Merge(lintIssues)
		} else {
			c.Error("config-load", "config", cfgErr.Error())
		}
	} else {
		// Get all issues (including warnings)
		issues := config.Lint(cfg)
		c.Merge(issues)

		// Add success info if no errors
		if !issues.HasErrors() {
			c.Info("config-valid", "config", "configuration is valid")
		}
	}

	// Step 2: Database checks (only if config loaded successfully)
	if cfg != nil {
		verifyDatabase(c, cfg)
	}

	// Print all issues
	issues := c.Issues()
	fmt.Println()
	for _, issue := range issues {
		fmt.Println(issue)
	}
	fmt.Println()

	// Summary
	errorCount, warnCount, infoCount := issues.Count()
	fmt.Printf("Summary: %d error(s), %d warning(s), %d info\n", errorCount, warnCount, infoCount)

	// Exit with error code if there are errors
	if errorCount > 0 {
		os.Exit(1)
	}

	return nil
}

func verifyDatabase(c *lint.Collector, cfg *config.Config) {
	// Open database connection
	db, err := sqlite.OpenDB(cfg.Database.DSN)
	if err != nil {
		c.Errorf("db-connect", "database", "failed to connect: %v", err)
		return
	}
	defer db.Close()

	// Test connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		c.Errorf("db-ping", "database", "connection test failed: %v", err)
		return
	}

	c.Info("db-connected", "database", "connection successful")

	// Check migration status
	m, err := sqlite.NewMigrate(db)
	if err != nil {
		c.Errorf("db-migrate", "database", "failed to create migrator: %v", err)
		return
	}

	version, dirty, err := m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		c.Warn("migrations-none", "database", "no migrations have been applied")
	} else if err != nil {
		c.Errorf("migrations-error", "database", "failed to get migration version: %v", err)
	} else if dirty {
		c.Errorf("migrations-dirty", "database", "migrations at version %d (DIRTY - migration failed)", version)
	} else {
		c.Infof("migrations-current", "database", "migrations at version %d", version)
	}
}
