package s3

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
)

const dbBackupBucket = "kite-db-backups"

func (c *Client) StoreDBBackup(
	ctx context.Context,
	database string,
	key string,
	size int64,
	reader io.Reader,
) error {
	objectName := fmt.Sprintf("%s/%s.tar.gz", database, key)

	if err := c.CreateBucketIfNotExists(ctx, dbBackupBucket); err != nil {
		return fmt.Errorf("failed to create db backup bucket: %w", err)
	}

	_, err := c.client.PutObject(ctx, dbBackupBucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType:          "application/tar+gzip",
		ServerSideEncryption: c.encryption,
	})
	if err != nil {
		return fmt.Errorf("failed to store db backup: %w", err)
	}

	return nil
}
