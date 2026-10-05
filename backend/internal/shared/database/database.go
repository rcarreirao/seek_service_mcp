// Package database provides GORM connection lifecycle helpers.
package database

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const (
	connectTimeout = 10 * time.Second
	maxOpenConns   = 20
	maxIdleConns   = 10
	connMaxLifetime = time.Hour
)

// Open connects to MySQL, configures the pool, and verifies connectivity
// with a ping bounded by connectTimeout.
func Open(ctx context.Context, dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("database: open: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("database: sql db: %w", err)
	}

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	pingCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		return nil, fmt.Errorf("database: ping: %w", err)
	}

	return db, nil
}

// AutoMigrate runs GORM AutoMigrate for the given models. With no models it
// is a no-op; later tasks pass their GORM models here.
func AutoMigrate(ctx context.Context, db *gorm.DB, models ...any) error {
	if len(models) == 0 {
		return nil
	}
	if err := db.WithContext(ctx).AutoMigrate(models...); err != nil {
		return fmt.Errorf("database: auto migrate: %w", err)
	}
	return nil
}

// Close releases the underlying connection pool.
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("database: sql db: %w", err)
	}
	return sqlDB.Close()
}