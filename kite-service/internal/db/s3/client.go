package s3

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/encrypt"
)

var buckets = []string{
	dbBackupBucket,
}

type Client struct {
	client     *minio.Client
	encryption encrypt.ServerSide
}

func New(cfg config.S3Config) (*Client, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.Secure,
	})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	for _, bucket := range buckets {
		exists, err := client.BucketExists(ctx, bucket)
		if err != nil {
			return nil, fmt.Errorf("failed to check if bucket %s exists: %w", bucket, err)
		}

		if !exists {
			err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
			if err != nil {
				return nil, fmt.Errorf("failed to create bucket %s: %w", bucket, err)
			}
		}
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
