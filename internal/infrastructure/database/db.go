package database

import (
	"context"
	"fmt"
	"main/internal/config"
	"net"
	"net/url"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const pingTimeout = 5 * time.Second

// Load opens the connection, checks it with a ping, sets the pool limits and migrates the schema.
func Load(ctx context.Context, dbConfig *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := DSN(dbConfig)

	// The ping below uses ctx and a timeout, so GORM's own ping (no timeout) is turned off.
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	if err := sqlDB.PingContext(pingCtx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping database at %s:%s: %w", dbConfig.Host, dbConfig.Port, err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	if err := Migrate(db); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("migrate database: %w", err)
	}

	return db, nil
}

// DSN builds the connection string as a postgres:// URL. url.UserPassword and the path escape
// every value, so a password with spaces, quotes or @ still works (key=value pairs broke on those).
func DSN(dbConfig *config.DatabaseConfig) string {
	dsn := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(dbConfig.User, dbConfig.Password),
		Host:     net.JoinHostPort(dbConfig.Host, dbConfig.Port),
		Path:     "/" + dbConfig.DBName,
		RawQuery: url.Values{"sslmode": {dbConfig.SSLMode}}.Encode(),
	}

	return dsn.String()
}

// migrateLockKey names the Postgres advisory lock Migrate holds.
const migrateLockKey int64 = 7_301_000

// Migrate creates or updates the tables. It holds a session advisory lock on one connection,
// so processes that start together (parallel test packages on a fresh database, the API and
// the worker) migrate one after another instead of colliding on CREATE TABLE.
// There is no transaction on purpose: one would keep a lock on every table AutoMigrate reads
// until the end, and deadlock with a concurrent TRUNCATE (the e2e suite) or other DDL.
func Migrate(db *gorm.DB) error {
	return db.Connection(func(conn *gorm.DB) (err error) {
		if err := conn.Exec("SELECT pg_advisory_lock(?)", migrateLockKey).Error; err != nil {
			return fmt.Errorf("take the migration lock: %w", err)
		}
		defer func() {
			if unlockErr := conn.Exec("SELECT pg_advisory_unlock(?)", migrateLockKey).Error; unlockErr != nil && err == nil {
				err = fmt.Errorf("release the migration lock: %w", unlockErr)
			}
		}()

		return conn.AutoMigrate(
			&User{},
			&Event{},
			&Announcement{},
			&Invitation{},
			&Registration{},
			&Committee{},
			&Session{},
			&RefreshToken{},
		)
	})
}

// Close closes the connection pool.
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}
