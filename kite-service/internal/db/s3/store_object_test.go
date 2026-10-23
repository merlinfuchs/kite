package s3

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
)

func TestIsNoSuchKeyError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name: "no such key",
			err: minio.ErrorResponse{
				Code:       "NoSuchKey",
				StatusCode: http.StatusNotFound,
			},
			expected: true,
		},
		{
			name:     "wrapped no such key",
			err:      fmt.Errorf("read: %w", minio.ErrorResponse{Code: "NoSuchKey"}),
			expected: true,
		},
		{
			name: "no such bucket",
			err: minio.ErrorResponse{
				Code:       "NoSuchBucket",
				StatusCode: http.StatusNotFound,
			},
			expected: false,
		},
		{
			name: "access denied",
			err: minio.ErrorResponse{
				Code:       "AccessDenied",
				StatusCode: http.StatusForbidden,
			},
			expected: false,
		},
		{
			name:     "other error",
			err:      errors.New("connection reset by peer"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isNoSuchKeyError(tt.err))
		})
	}
}
