package s3

import (
	"errors"
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
			name: "minio error response with NoSuchKey code",
			err: minio.ErrorResponse{
				Code:       "NoSuchKey",
				Message:    "The specified key does not exist.",
				StatusCode: http.StatusNotFound,
			},
			expected: true,
		},
		{
			name: "minio error response with 404 status and different code",
			err: minio.ErrorResponse{
				Code:       "NotFound",
				StatusCode: http.StatusNotFound,
			},
			expected: true,
		},
		{
			name: "minio error response with 404 status and empty code",
			err: minio.ErrorResponse{
				StatusCode: http.StatusNotFound,
			},
			expected: true,
		},
		{
			name:     "wrapped error with NoSuchKey text",
			err:      errors.New("s3: NoSuchKey: object does not exist"),
			expected: true,
		},
		{
			name:     "standard error string fallback",
			err:      errors.New("The specified key does not exist."),
			expected: true,
		},
		{
			name: "access denied error",
			err: minio.ErrorResponse{
				Code:       "AccessDenied",
				StatusCode: http.StatusForbidden,
			},
			expected: false,
		},
		{
			name: "internal server error",
			err: minio.ErrorResponse{
				Code:       "InternalError",
				StatusCode: http.StatusInternalServerError,
			},
			expected: false,
		},
		{
			name:     "generic unrelated error",
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
