package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/minio/minio-go/v7"
)

func (c *Client) CreateBucketIfNotExists(ctx context.Context, bucket string) error {
	exists, err := c.client.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("failed to check if bucket %s exists: %w", bucket, err)
	}

	if !exists {
		err = c.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket %s: %w", bucket, err)
		}
	}

	return nil
}

func (c *Client) UploadObject(ctx context.Context, bucket string, object *model.Object) error {
	reader := bytes.NewReader(object.Content)
	length := int64(len(object.Content))

	_, err := c.client.PutObject(ctx, bucket, object.Name, reader, length, minio.PutObjectOptions{
		ContentType:          object.ContentType,
		ServerSideEncryption: c.encryption,
	})
	if err != nil {
		return fmt.Errorf("failed to upload object %s to bucket %s: %w", object.Name, bucket, err)
	}

	return nil
}

func (c *Client) UploadObjectIfNotExists(ctx context.Context, bucket string, object *model.Object) error {
	reader := bytes.NewReader(object.Content)
	length := int64(len(object.Content))

	exists, err := c.client.StatObject(ctx, bucket, object.Name, minio.StatObjectOptions{})
	// TODO: refactor to not use error string
	if err != nil && err.Error() != "The specified key does not exist." {
		return err
	}

	if exists.Size > 0 {
		return nil
	}

	_, err = c.client.PutObject(ctx, bucket, object.Name, reader, length, minio.PutObjectOptions{
		ContentType:          object.ContentType,
		ServerSideEncryption: c.encryption,
	})
	if err != nil {
		return fmt.Errorf("failed to upload object %s to bucket %s: %w", object.Name, bucket, err)
	}

	return nil
}

func (c *Client) DownloadObject(ctx context.Context, bucket string, name string) (*model.Object, error) {
	object, err := c.client.GetObject(ctx, bucket, name, minio.GetObjectOptions{
		ServerSideEncryption: c.encryption,
	})
	if err != nil {
		if err.Error() == "The specified key does not exist." {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	data, err := io.ReadAll(object)
	if err != nil {
		return nil, err
	}

	info, err := object.Stat()
	if err != nil {
		return nil, err
	}

	return &model.Object{
		Name:        name,
		ContentType: info.ContentType,
		Content:     data,
	}, nil
}

func (c *Client) DeleteObject(ctx context.Context, bucket string, name string) error {
	err := c.client.RemoveObject(ctx, bucket, name, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object %s from bucket %s: %w", name, bucket, err)
	}

	return nil
}
