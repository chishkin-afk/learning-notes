package app

import (
	"log/slog"
	"os"

	"github.com/chishkin-afk/learning_notes/internal/infrastructure/config"
	"github.com/chishkin-afk/learning_notes/internal/infrastructure/persistence/postgres"
	logger "github.com/chishkin-afk/learning_notes/pkg/log"
	"github.com/jmoiron/sqlx"
)

// DIContainer manages dependencies
//
// This structure contains all the core dependencies and their corresponding getters
// the initialization of these dependencies takes place precisely within the getters.
type DIContainer struct {
	cfg *config.Config

	log *slog.Logger

	db *sqlx.DB

	// ...server

	// ...handlers

	// ...services
}

// Cfg returns config
//
// The configuration is loaded only once.
// If loading fails, the process logs the error and exits.
func (di *DIContainer) Config() *config.Config {
	if di.cfg == nil {
		cfg, err := config.New(os.Getenv("APP_CONFIG_PATH"))
		if err != nil {
			slog.Error("failed to load config", slog.String("error", err.Error()))
			os.Exit(1)
		}

		di.cfg = cfg
	}

	return di.cfg
}

func (di *DIContainer) DB() *sqlx.DB {
	if di.db == nil {
		db, err := postgres.Connect(di.Config())
		if err != nil {
			slog.Error("failed to connect db",
				slog.String("error", err.Error()),
			)
		}

		di.db = db
	}

	return di.db
}

func (di *DIContainer) Log() *slog.Logger {
	if di.log == nil {
		logHandler := logger.NewHandler(di.Config().App.Env)
		di.log = slog.New(logHandler)
	}

	return di.log
}
