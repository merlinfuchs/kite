package postgres

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/merlinfuchs/kite/kite-service/internal/config"
	"github.com/merlinfuchs/kite/kite-service/internal/db/postgres/pgmodel"
)

type Client struct {
	db            *sqlx.DB
	Q             *pgmodel.Queries
	connectionDSN string
}

func New(connectionDSN string) (*Client, error) {
	db, err := sqlx.Open("postgres", connectionDSN)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to postgres store: %v", err)
	}

	db.SetMaxIdleConns(20)
	db.SetMaxOpenConns(80)
	db.SetConnMaxLifetime(time.Hour * 1)

	return &Client{
		db:            db,
		Q:             pgmodel.New(db),
		connectionDSN: connectionDSN,
	}, nil
}

func BuildConnectionDSN(cfg config.ServerPostgresConfig) string {
	dsn := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s sslmode=disable connect_timeout=4",
		cfg.Host, cfg.Port, cfg.DBName, cfg.User,
	)

	if cfg.Password != "" {
		dsn += " password=" + cfg.Password
	}
	return dsn
}
