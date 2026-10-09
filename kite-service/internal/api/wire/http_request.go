package wire

import (
	"encoding/json"
	"errors"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
)

const (
	maxHTTPRequestTestValues      = 50
	maxHTTPRequestTestKeyLength   = 200
	maxHTTPRequestTestValueLength = 10_000
)

type FlowHTTPRequestTestRequest struct {
	Data *flow.HTTPRequestData `json:"data"`
	// Values stand in for the placeholders of the block during the test,
	// keyed by the placeholder without braces, like arg('user') or user.id.
	Values map[string]string `json:"values"`
}

func (req FlowHTTPRequestTestRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Data, validation.Required),
		validation.Field(&req.Values,
			validation.Length(0, maxHTTPRequestTestValues),
			validation.By(func(value interface{}) error {
				for k, v := range req.Values {
					if len(k) > maxHTTPRequestTestKeyLength {
						return errors.New("test value keys must be at most 200 characters")
					}
					if len(v) > maxHTTPRequestTestValueLength {
						return errors.New("test values must be at most 10000 characters")
					}
				}
				return nil
			}),
		),
	)
}

type FlowHTTPRequestTestResponse struct {
	// The request as it was sent, after placeholders were filled in. Empty
	// when the request couldn't be built.
	Method               string            `json:"method"`
	URL                  string            `json:"url"`
	RequestHeaders       map[string]string `json:"request_headers"`
	RequestBody          string            `json:"request_body"`
	RequestBodyTruncated bool              `json:"request_body_truncated"`

	// Nil when no response was received.
	Response *FlowHTTPRequestTestResponseData `json:"response"`

	// The JSON of what the transform expression returned, when the block
	// has one.
	Result json.RawMessage `json:"result"`

	Error      string `json:"error"`
	ErrorStage string `json:"error_stage"`
	DurationMS int64  `json:"duration_ms"`
}

type FlowHTTPRequestTestResponseData struct {
	Status        string            `json:"status"`
	StatusCode    int               `json:"status_code"`
	Headers       map[string]string `json:"headers"`
	Body          string            `json:"body"`
	BodySize      int               `json:"body_size"`
	BodyTruncated bool              `json:"body_truncated"`
	BodyBinary    bool              `json:"body_binary"`
}
