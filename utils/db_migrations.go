package utils

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
)

// RunDBMigrations runs SQL file migrations and GORM AutoMigrate for all entities.
func RunDBMigrations(db *gorm.DB) error {
	migrationRunner := NewMigrationRunner(db, "migrations")
	if err := migrationRunner.Run(); err != nil {
		return fmt.Errorf("failed to run SQL file migrations: %w", err)
	}

	if err := db.AutoMigrate(
		&entities.User{},
		&entities.OTPLog{},
		&entities.RefreshToken{},
		&entities.NotifyMe{},
		&entities.Address{},
		&entities.State{},
		&entities.City{},
		&entities.PinCode{},
		&entities.QuestionMaster{},
		&entities.ThunderSeat{},
		&entities.ThunderSeatWinner{},
	); err != nil {
		return fmt.Errorf("failed to run GORM automigrations: %w", err)
	}

	log.Info("Database migrations completed")
	return nil
}


