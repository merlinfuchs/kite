package model

import "time"

type Object struct {
	Name        string
	Content     []byte
	ContentType string
}

// ObjectInfo is what listing a bucket returns about each object.
type ObjectInfo struct {
	Name         string
	Size         int64
	LastModified time.Time
}
