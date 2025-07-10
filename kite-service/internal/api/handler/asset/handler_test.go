package asset

import (
	"testing"
)

func TestValidateAndSanitizeContentType(t *testing.T) {
	tests := []struct {
		name            string
		userContentType string
		expectedType    string
	}{
		{
			name:            "valid image",
			userContentType: "image/jpeg",
			expectedType:    "image/jpeg",
		},
		{
			name:            "valid audio",
			userContentType: "audio/mpeg",
			expectedType:    "audio/mpeg",
		},
		{
			name:            "valid video",
			userContentType: "video/mp4",
			expectedType:    "video/mp4",
		},
		{
			name:            "valid pdf",
			userContentType: "application/pdf",
			expectedType:    "application/pdf",
		},
		{
			name:            "valid text",
			userContentType: "text/plain",
			expectedType:    "text/plain",
		},
		{
			name:            "valid json",
			userContentType: "application/json",
			expectedType:    "application/json",
		},
		{
			name:            "valid xml",
			userContentType: "application/xml",
			expectedType:    "application/xml",
		},
		{
			name:            "invalid html",
			userContentType: "text/html",
			expectedType:    "application/octet-stream",
		},
		{
			name:            "invalid js",
			userContentType: "application/javascript",
			expectedType:    "application/octet-stream",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeContentType(tt.userContentType)

			if result != tt.expectedType {
				t.Errorf("expected content type %s, got %s", tt.expectedType, result)
			}
		})
	}
}
