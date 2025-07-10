package asset

import (
	"strings"

	"github.com/gobwas/glob"
)

var allowedContentTypes = []string{
	"image/*",
	"audio/*",
	"video/*",
	"application/pdf",
	"text/plain",
	"application/json",
	"application/xml",
}

var allowedContentTypePatterns []glob.Glob

func init() {
	allowedContentTypePatterns = make([]glob.Glob, len(allowedContentTypes))
	for i, pattern := range allowedContentTypes {
		allowedContentTypePatterns[i] = glob.MustCompile(pattern)
	}
}

// sanitizeContentType sanitizes the content type to prevent XSS attacks
func sanitizeContentType(userContentType string) string {
	userContentType = strings.TrimSpace(strings.ToLower(userContentType))
	for _, pattern := range allowedContentTypePatterns {
		if pattern.Match(userContentType) {
			return userContentType
		}
	}

	// Default to application/octet-stream for unknown types
	return "application/octet-stream"
}
