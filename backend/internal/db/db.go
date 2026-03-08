package db

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/erpang/post-sync/internal/config"
	"github.com/erpang/post-sync/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Open(cfg config.Config) (*gorm.DB, error) {
	switch cfg.DBDriver {
	case "sqlite":
		if err := ensureSQLiteDir(cfg.DBDSN); err != nil {
			return nil, err
		}

		database, err := gorm.Open(sqlite.Open(cfg.DBDSN), &gorm.Config{})
		if err != nil {
			return nil, err
		}

		return migrate(database)
	case "postgres":
		database, err := gorm.Open(postgres.Open(cfg.DBDSN), &gorm.Config{})
		if err != nil {
			return nil, err
		}

		return migrate(database)
	default:
		return nil, fmt.Errorf("unsupported db driver: %s", cfg.DBDriver)
	}
}

func migrate(database *gorm.DB) (*gorm.DB, error) {
	if err := database.AutoMigrate(
		&domain.Content{},
		&domain.ChannelAccount{},
		&domain.ChannelTarget{},
		&domain.PublishJob{},
		&domain.DeliveryTask{},
	); err != nil {
		return nil, err
	}

	return database, nil
}

func ensureSQLiteDir(dsn string) error {
	if dsn == "" || dsn == ":memory:" {
		return nil
	}

	dir := filepath.Dir(dsn)
	if dir == "." || dir == "" {
		return nil
	}

	return os.MkdirAll(dir, 0o755)
}
