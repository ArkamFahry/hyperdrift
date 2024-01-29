package database

import (
	"github.com/ArkamFahry/hyperdrift/storage/server/common/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/net/context"
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
			zap.String("operation", op),
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
