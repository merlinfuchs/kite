package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kitecloud/kite/kite-service/internal/config"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/store"
	_ "github.com/lib/pq"
)

type Client struct {
	DB            *pgxpool.Pool
	Q             *pgmodel.Queries
	connectionDSN string
}

func New(connectionDSN string, clusterCount int) (*Client, error) {
	config, err := pgxpool.ParseConfig(connectionDSN)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse postgres config: %v", err)
	}

	// This assumes that Postgres is configured to support a total of 100 connections.
	// We max out at 80 to allow for some overhead.
	config.MaxConns = int32(80 / clusterCount)
	config.MinConns = 10
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute

	db, err := pgxpool.NewWithConfig(context.Background(), config)
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

// uniqueViolation turns a violated unique constraint into store.ErrAlreadyExists.
func uniqueViolation(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return store.ErrAlreadyExists
	}
	return err
}

// PoolStats is how busy the connection pool is. A rising EmptyAcquireCount
// means queries waited for a free connection.
type PoolStats struct {
	MaxConns              int32 `json:"max_conns"`
	TotalConns            int32 `json:"total_conns"`
	AcquiredConns         int32 `json:"acquired_conns"`
	IdleConns             int32 `json:"idle_conns"`
	AcquireCount          int64 `json:"acquire_count"`
	AcquireNanosTotal     int64 `json:"acquire_nanos_total"`
	EmptyAcquireCount     int64 `json:"empty_acquire_count"`
	EmptyAcquireWaitNanos int64 `json:"empty_acquire_wait_nanos"`
	CanceledAcquireCount  int64 `json:"canceled_acquire_count"`
}

func (c *Client) PoolStats() any {
	s := c.DB.Stat()
	return PoolStats{
		MaxConns:              s.MaxConns(),
		TotalConns:            s.TotalConns(),
		AcquiredConns:         s.AcquiredConns(),
		IdleConns:             s.IdleConns(),
		AcquireCount:          s.AcquireCount(),
		AcquireNanosTotal:     s.AcquireDuration().Nanoseconds(),
		EmptyAcquireCount:     s.EmptyAcquireCount(),
		EmptyAcquireWaitNanos: s.EmptyAcquireWaitTime().Nanoseconds(),
		CanceledAcquireCount:  s.CanceledAcquireCount(),
	}
}
