package thing

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

const maxBodySize = 64 * 1024 // 64KB

type HTTPResponseValue struct {
	Status     string            `json:"status"`
	StatusCode int               `json:"status_code"`
	Body       []byte            `json:"body"`
	Headers    map[string]string `json:"headers"`
}

func NewHTTPResponseValue(v *http.Response) (HTTPResponseValue, error) {
	headers := make(map[string]string, len(v.Header))
	for k, v := range v.Header {
		headers[k] = strings.Join(v, ",")
	}

	if v.ContentLength > maxBodySize {
		return HTTPResponseValue{}, fmt.Errorf("body size exceeds max body size of %d bytes", maxBodySize)
	}

	limitedBody := io.LimitReader(v.Body, maxBodySize)
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
