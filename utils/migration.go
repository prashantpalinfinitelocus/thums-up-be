package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type MigrationRunner struct {
	db             *gorm.DB
	migrationsPath string
}

func NewMigrationRunner(db *gorm.DB, migrationsPath string) *MigrationRunner {
	return &MigrationRunner{
		db:             db,
		migrationsPath: migrationsPath,
	}
}

func (mr *MigrationRunner) Run() error {
	if err := mr.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	files, err := mr.getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	for _, file := range files {
		if err := mr.runMigration(file); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", file, err)
		}
	}

	log.Info("All migrations completed successfully")
	return nil
}

func (mr *MigrationRunner) createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			version VARCHAR(255) UNIQUE NOT NULL,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`
	return mr.db.Exec(query).Error
}

func (mr *MigrationRunner) getMigrationFiles() ([]string, error) {
	var files []string

	entries, err := os.ReadDir(mr.migrationsPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warn("Migrations directory not found, skipping migrations")
			return files, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}

	sort.Strings(files)
	return files, nil
}

func (mr *MigrationRunner) runMigration(filename string) error {
	var count int64
	mr.db.Raw("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", filename).Scan(&count)

	if count > 0 {
		log.WithField("migration", filename).Debug("Migration already applied, skipping")
		return nil
	}

	filePath := filepath.Join(mr.migrationsPath, filename)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	log.WithField("migration", filename).Info("Running migration")

	if err := mr.db.Exec(string(content)).Error; err != nil {
		return err
	}

	if err := mr.db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", filename).Error; err != nil {
		return err
	}

	log.WithField("migration", filename).Info("Migration completed successfully")
	return nil
}
