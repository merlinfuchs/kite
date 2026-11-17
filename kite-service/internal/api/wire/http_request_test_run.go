package wire

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
)

// HTTPRequestTestRequest sends the request of a Send API Request block, to
// try it out in the editor.
type HTTPRequestTestRequest struct {
	Data flow.HTTPRequestData `json:"data"`
	// The text of each placeholder by its expression, like "user.id", as
	// there's no event to fill them in. Placeholders of the app's secrets
	// don't need one.
	Values map[string]string `json:"values,omitempty"`
}

func (req HTTPRequestTestRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Values, validation.Length(0, 100)),
	)
}

type HTTPRequestTestResponse struct {
	Method string `json:"method"`
	URL    string `json:"url"`
	// Null if no response was received, see Error.
	Response   *HTTPRequestTestResponseData `json:"response"`
	DurationMS int64                        `json:"duration_ms"`
	// Why the request couldn't be built or sent.
	Error string `json:"error,omitempty"`
}

type HTTPRequestTestResponseData struct {
	Status     string            `json:"status"`
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	// Cut off after 64 KB.
	Body          string `json:"body"`
	BodyTruncated bool   `json:"body_truncated"`
}
