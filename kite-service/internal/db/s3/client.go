package s3

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/encrypt"
)

func init() {
	// Package-wide in minio. Its 1s default gives up on a throttled request
	// within ~5s, and Hetzner throttles for longer than that. With minio's 10
	// retries, requests now retry for ~15s on average.
	minio.DefaultRetryCap = 5 * time.Second
}

// RetryLonger makes every request of this process retry for ~30s on average.
// It's for processes nobody waits on, like backups, which should outlast
// throttling rather than fail.
func RetryLonger() {
	minio.MaxRetry = 15
}

type Client struct {
	client     *minio.Client
	encryption encrypt.ServerSide
}

// New builds an S3 client without talking to the server. Buckets are created
// on demand by the code that uses them, so that a Kite instance which doesn't
// need S3 (or has it misconfigured) can still start up.
func New(cfg config.S3Config) (*Client, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.Secure,
	})
	if err != nil {
		return nil, err
	}

	var encryption encrypt.ServerSide
	if cfg.SSECKey != "" {
		key, err := hex.DecodeString(cfg.SSECKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decode S3 encryption key: %w", err)
		}

		encryption, err = encrypt.NewSSEC(key)
		if err != nil {
			return nil, fmt.Errorf("failed to create S3 encryption: %w", err)
		}
	}

	return &Client{
		client:     client,
		encryption: encryption,
	}, nil
}
