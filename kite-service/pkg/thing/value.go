package thing

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// MaxBodySize is how much of an HTTP response body a flow can capture.
const MaxBodySize = 4 * 1024 * 1024 // 4MB

type HTTPResponseValue struct {
	Status     string            `json:"status"`
	StatusCode int               `json:"status_code"`
	Body       []byte            `json:"body"`
	Headers    map[string]string `json:"headers"`
}

// NewHTTPResponseValue reads the response. A body of more than maxBody bytes
// is cut off, or an error if its Content-Length says so.
func NewHTTPResponseValue(v *http.Response, maxBody int) (HTTPResponseValue, error) {
	headers := make(map[string]string, len(v.Header))
	for k, v := range v.Header {
		headers[k] = strings.Join(v, ",")
	}

	if v.ContentLength > int64(maxBody) {
		return HTTPResponseValue{}, fmt.Errorf("body size exceeds max body size of %d bytes", maxBody)
	}

	limitedBody := io.LimitReader(v.Body, int64(maxBody))
	body, err := io.ReadAll(limitedBody)
	if err != nil {
		return HTTPResponseValue{}, err
	}

	return HTTPResponseValue{
		Status:     v.Status,
		StatusCode: v.StatusCode,
		Body:       body,
		Headers:    headers,
	}, nil
}

type RobloxUserValue struct {
	ID                     int64  `json:"id"`
	Name                   string `json:"name"`
	DisplayName            string `json:"displayName"`
	Description            string `json:"description"`
	CreatedAt              string `json:"created"`
	IsBanned               bool   `json:"isBanned"`
	HasVerifiedBadge       bool   `json:"hasVerifiedBadge"`
	ExternalAppDisplayName string `json:"externalAppDisplayName"`
	AvatarURL              string `json:"avatarUrl"`
}
