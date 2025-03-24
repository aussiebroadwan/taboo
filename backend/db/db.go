package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	_ "modernc.org/sqlite"
)

const (
	databaseLocalFile    = "/data/data.db"
	databaseRollbackFile = "/data/data-rollback.db"
)

var DatabaseName = databaseLocalFile

// DB is a singleton database connection manager.
var (
	DB     *sql.DB
	dbOnce sync.Once
	dbMu   sync.RWMutex
)

//go:embed migrations/*.sql
var migrationsFolder embed.FS

// GetDB returns the singleton database connection.
func GetDB() *sql.DB {
	dbMu.Lock()
	defer dbMu.Unlock()

	dbOnce.Do(func() {
		var err error
		DB, err = sql.Open("sqlite", "file:"+DatabaseName+"?_busy_timeout=5000&_journal_mode=WAL")
		if err != nil {
			log.Fatal(err)
		}

		// Set optimal connection settings for SQLite
		DB.SetMaxOpenConns(1)
		DB.SetMaxIdleConns(1)
		DB.SetConnMaxLifetime(0) // Disable automatic closing

		// Run database migrations
		if err = runMigrations(DB); err != nil {
			log.Fatal(err)
		}

		// Enable WAL mode
		_, err = DB.Exec("PRAGMA journal_mode=WAL;")
		if err != nil {
			log.Fatal("Failed to enable WAL mode:", err)
		}
	})

	return DB
}

// Ping checks the database connection.
func Ping(ctx context.Context) error {
	db := GetDB()
	return db.PingContext(ctx)
}

// Close closes the database connection.
func Close() error {
	dbMu.Lock()
	defer dbMu.Unlock()

	if DB != nil {
		slog.Info("Closing Database Connection")
		err := DB.Close()
		DB = nil
		dbOnce = sync.Once{} // Reset the dbOnce to allow re-initialization

		if err != nil {
			slog.Error("Error happened when closing the database file", slog.Any("err", err))
		}
		return err
	}
	return nil
}

// ExecuteTx executes a SQL query within a transaction.
func ExecuteTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	db := GetDB()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Ensure transaction is handled properly
	err = fn(tx)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			slog.Error("db failed to rollback transaction", slog.Any("rollback_error", rollbackErr))
		}
		return err
	}

	if commitErr := tx.Commit(); commitErr != nil {
		slog.Error("db failed to commit transaction", slog.Any("commit_error", commitErr))
		return commitErr
	}

	return nil
}

// reconnect attempts to reconnect to the database.
func reconnect() error {
	// Close the current connection
	err := Close()
	if err != nil {
		log.Printf("Error closing database connection: %v", err)
	}

	// Open a new connection
	db, err := sql.Open("sqlite", DatabaseName)
	if err != nil {
		return err
	}

	// Replace the current connection with the new one
	DB = db

	return nil
}

// RestoreDatabase replaces the current database with a new one and runs migrations.
// If migrations fail, it restores the previous database.
func RestoreDatabase(newDbPath string) error {
	slog.Info("Starting database restore", slog.String("restore_file", newDbPath))

	// Close the current database connection
	slog.Info("Closing existing database connection")
	if err := Close(); err != nil {
		slog.Error("Failed to close database before restore", slog.Any("err", err))
		return fmt.Errorf("failed to close database: %w", err)
	}

	// Backup the existing database
	if _, err := os.Stat(DatabaseName); err == nil {
		slog.Info("Creating database rollback backup", slog.String("backup_file", databaseRollbackFile))
		err = copyFile(DatabaseName, databaseRollbackFile)
		if err != nil {
			slog.Error("Failed to create database rollback backup", slog.Any("err", err))
			return fmt.Errorf("failed to backup current database: %w", err)
		}
	}

	// Replace the database with the new one
	slog.Info("Replacing database with restore file")
	err := copyFile(newDbPath, DatabaseName)
	if err != nil {
		slog.Error("Failed to replace database with restore file", slog.Any("err", err))
		return fmt.Errorf("failed to copy restore database: %w", err)
	}

	// Reconnect to the new database
	slog.Info("Reconnecting to the new database")
	db, err := sql.Open("sqlite", "file:"+DatabaseName+"?_busy_timeout=5000&_journal_mode=WAL")
	if err != nil {
		slog.Error("Failed to reconnect to new database", slog.Any("err", err))
		restorePreviousDatabase()
		return fmt.Errorf("failed to reopen new database: %w", err)
	}

	// Assign new DB connection
	dbMu.Lock()
	DB = db
	dbMu.Unlock()

	// Run migrations
	slog.Info("Running database migrations")
	err = runMigrations(DB)
	if err != nil {
		slog.Error("Database migrations failed", slog.Any("err", err))
		restorePreviousDatabase()
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	slog.Info("Database restored successfully")
	return nil
}

// restorePreviousDatabase restores the rollback database if the restore fails.
func restorePreviousDatabase() {
	slog.Warn("Restoring previous database due to restore failure")

	// Close the database before restoring the rollback
	if err := Close(); err != nil {
		slog.Error("Failed to close database during rollback", slog.Any("err", err))
	}

	// Restore the rollback database
	err := copyFile(databaseRollbackFile, DatabaseName)
	if err != nil {
		slog.Error("Failed to restore rollback database", slog.Any("err", err))
		return
	}

	// Reconnect to the rollback database
	db, err := sql.Open("sqlite", "file:"+DatabaseName+"?_busy_timeout=5000&_journal_mode=WAL")
	if err != nil {
		slog.Error("Failed to reconnect to rollback database", slog.Any("err", err))
		return
	}

	// Assign rollback database to the global DB variable
	dbMu.Lock()
	DB = db
	dbMu.Unlock()

	// Run migrations again to ensure consistency
	err = runMigrations(DB)
	if err != nil {
		slog.Error("Failed to re-run migrations on rollback database", slog.Any("err", err))
	}
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	slog.Info("Copying file", slog.String("source", src), slog.String("destination", dst))

	in, err := os.Open(filepath.Clean(src))
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(filepath.Clean(dst))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return out.Sync()
}

func runMigrations(db *sql.DB) error {
	dbDriver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return err
	}

	d, err := iofs.New(migrationsFolder, "migrations")
	if err != nil {
		log.Fatal(err)
	}

	// Create a new migrate instance
	m, err := migrate.NewWithInstance("iofs", d, "", dbDriver)
	if err != nil {
		return err
	}

	// Run migrations
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	slog.Info("Migrations applied successfully")
	return nil
}
