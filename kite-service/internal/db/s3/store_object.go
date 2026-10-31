package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/minio/minio-go/v7"
)

// minio sets the NoSuchKey code itself for 404s without a body, like on StatObject.
func isNoSuchKeyError(err error) bool {
	var errResp minio.ErrorResponse
	return errors.As(err, &errResp) && errResp.Code == "NoSuchKey"
}

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

	exists, err := c.client.StatObject(ctx, bucket, object.Name, minio.StatObjectOptions{
		ServerSideEncryption: c.encryption,
	})
	if err != nil && !isNoSuchKeyError(err) {
		return err
	}

	if err == nil && exists.Size > 0 {
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
		if isNoSuchKeyError(err) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}
	defer object.Close()

	data, err := io.ReadAll(object)
	if err != nil {
		if isNoSuchKeyError(err) {
			return nil, store.ErrNotFound
		}
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

func (c *Client) ListObjects(ctx context.Context, bucket string) ([]model.ObjectInfo, error) {
	var objects []model.ObjectInfo
	for object := range c.client.ListObjects(ctx, bucket, minio.ListObjectsOptions{Recursive: true}) {
		if object.Err != nil {
			return nil, fmt.Errorf("failed to list objects of bucket %s: %w", bucket, object.Err)
		}
		objects = append(objects, model.ObjectInfo{
			Name:         object.Key,
			Size:         object.Size,
			LastModified: object.LastModified,
		})
	}
	return objects, nil
}
