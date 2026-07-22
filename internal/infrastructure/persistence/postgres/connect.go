package postgres

import (
	"fmt"

	"github.com/chishkin-afk/learning_notes/internal/infrastructure/config"
	"github.com/jmoiron/sqlx"
)

// Connect opens the connection with postgres
//
// This function establishes a connection to the PostgreSQL database,
// pings it, and applies basic connection settings derived from config.Config.
func Connect(cfg *config.Config) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", getDSN(cfg))
	if err != nil {
		return nil, fmt.Errorf("failed to open connection with postgres: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	applyConfig(db, cfg)
	return db, nil
}

func applyConfig(db *sqlx.DB, cfg *config.Config) {
	db.SetConnMaxIdleTime(cfg.Persistence.Postgres.Conns.MaxIdleTime)
	db.SetConnMaxLifetime(cfg.Persistence.Postgres.Conns.MaxLifetime)
	db.SetMaxIdleConns(cfg.Persistence.Postgres.Conns.MaxIdles)
	db.SetMaxOpenConns(cfg.Persistence.Postgres.Conns.MaxOpens)
}

func getDSN(cfg *config.Config) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Persistence.Postgres.Host,
		cfg.Persistence.Postgres.Port,
		cfg.Persistence.Postgres.Auth.User,
		cfg.Persistence.Postgres.Auth.Password,
		cfg.Persistence.Postgres.Auth.DB,
		cfg.Persistence.Postgres.SSLMode,
	)
}
