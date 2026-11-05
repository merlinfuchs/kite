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
	reader io.Reader,
) (int64, error) {
	objectName := fmt.Sprintf("%s/%s.tar.gz", database, key)

	if err := c.CreateBucketIfNotExists(ctx, dbBackupBucket); err != nil {
		return 0, fmt.Errorf("failed to create db backup bucket: %w", err)
	}

	// The size is unknown as the backup is streamed, so each part is held in
	// memory. The default part size for that is over 500 MiB, to allow 5 TiB.
	// MD5 instead of the CRC32C checksums minio sends for streams, which not
	// every S3 implementation supports. Not ConcurrentStreamParts, as it
	// cancels the request before aborting, so a failed upload leaves its
	// parts behind.
	info, err := c.client.PutObject(ctx, dbBackupBucket, objectName, reader, -1, minio.PutObjectOptions{
		ContentType:          "application/tar+gzip",
		ServerSideEncryption: c.encryption,
		PartSize:             64 << 20,
		SendContentMd5:       true,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to store db backup: %w", err)
	}

	return info.Size, nil
}
