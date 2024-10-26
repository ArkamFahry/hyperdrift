package database

import (
	"database/sql"
	"embed"
	"github.com/teapartydev/storage/server/config"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func NewMigrations(config *config.Config, logger *zap.Logger) {
	const op = "database.migrations.Migrate"

	db, err := sql.Open("pgx", config.PostgresUrl)
	if err != nil {
		logger.Fatal("failed to set up pgx connection for migration",
			zap.Error(err),
			zap.String("operation", op),
		)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("failed to close pgx connection for migration",
				zap.Error(err),
				zap.String("operation", op),
			)
		}
	}(db)

	goose.SetTableName("storage.goose_db_version")

	goose.SetBaseFS(embedMigrations)

	if err = goose.SetDialect("postgres"); err != nil {
		logger.Fatal("failed to set migration dialect",
			zap.Error(err),
			zap.String("operation", op),
		)
	}

	if err = goose.Up(db, "migrations"); err != nil {
		logger.Fatal("failed to to run up migration",
			zap.Error(err),
			zap.String("operation", op),
		)
	}

	logger.Info("database migrations done successfully", zap.String("operation", op))
}
