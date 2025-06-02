package database

import (
	"fmt"

	"github.com/netabakovv/forum/back/pkg/logger"

	"github.com/golang-migrate/migrate"
)

func RunMigrations(dbURL string, log logger.Logger) error {
	m, err := migrate.New(
		"file://migrations",
		dbURL)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Info("migrations completed successfully")
	return nil
}
