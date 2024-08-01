package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kitecloud/kite/kite-service/internal/config"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	_ "github.com/lib/pq"
)

type Client struct {
	DB            *pgxpool.Pool
	Q             *pgmodel.Queries
	connectionDSN string
}

func New(connectionDSN string) (*Client, error) {
	db, err := pgxpool.New(context.Background(), connectionDSN)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to postgres db: %v", err)
	}

	return &Client{
		DB:            db,
		Q:             pgmodel.New(db),
		connectionDSN: connectionDSN,
	}, nil
}

func BuildConnectionDSN(cfg config.PostgresConfig) string {
	dsn := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s sslmode=disable connect_timeout=4",
		cfg.Host, cfg.Port, cfg.DBName, cfg.User,
	)

	if cfg.Password != "" {
		dsn += " password=" + cfg.Password
	}
	return dsn
}
