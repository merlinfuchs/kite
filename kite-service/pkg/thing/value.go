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

// NewHTTPResponseValue reads the response, with a body of up to maxBody
// bytes, like MaxBodySize.
func NewHTTPResponseValue(v *http.Response, maxBody int) (HTTPResponseValue, error) {
	headers := make(map[string]string, len(v.Header))
	for k, v := range v.Header {
		headers[k] = strings.Join(v, ",")
	}

	tooLarge := fmt.Errorf("the response is larger than %s", byteSize(maxBody))
	if v.ContentLength > int64(maxBody) {
		return HTTPResponseValue{}, tooLarge
	}
	body, err := io.ReadAll(io.LimitReader(v.Body, int64(maxBody)+1))
	if err != nil {
		return HTTPResponseValue{}, err
	}
	if len(body) > maxBody {
		return HTTPResponseValue{}, tooLarge
	}

	return HTTPResponseValue{
		Status:     v.Status,
		StatusCode: v.StatusCode,
		Body:       body,
		Headers:    headers,
	}, nil
}

// byteSize is n bytes in MB or KB, like 4 MB.
func byteSize(n int) string {
	if n >= 1<<20 {
		return fmt.Sprintf("%d MB", n>>20)
	}
	return fmt.Sprintf("%d KB", n>>10)
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
