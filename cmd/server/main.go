package main

import (
	"fmt"
	"log/slog"

	"github.com/chishkin-afk/learning_notes/internal/app"
	"github.com/chishkin-afk/learning_notes/internal/infrastructure/persistence/postgres"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		slog.Warn("failed to load .env",
			slog.String("error", err.Error()),
		)
	}
	var di app.DIContainer

	fmt.Println(postgres.Migrate(di.Config()))
}
