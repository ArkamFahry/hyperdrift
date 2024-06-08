package database

import (
	"context"
	"github.com/driftdev/storage/server/config"
	"github.com/driftdev/storage/server/zapfield"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Database struct {
	dbPoll *pgxpool.Pool
}

func NewDatabasePool(config *config.Config, logger *zap.Logger) *Database {
	const op = "database.NewDatabasePool"

	db, err := pgxpool.New(context.Background(), config.PostgresUrl)
	if err != nil {
		logger.Fatal("failed to connect to create pgx pool database connection",
			zap.Error(err),
			zapfield.Operation(op),
		)
	}

	return &Database{dbPoll: db}
}

func (d *Database) GetDatabase() (*pgxpool.Conn, error) {
	return d.dbPoll.Acquire(context.Background())
}
func (d *Database) Ping() error {
	return d.dbPoll.Ping(context.Background())
}

func (d *Database) Close() {
	if d.dbPoll != nil {
		d.dbPoll.Close()
	}
}
