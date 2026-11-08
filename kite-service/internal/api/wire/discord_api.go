package wire

import (
	"encoding/json"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// DiscordAPIRequestRequest calls an endpoint of the Discord API Request block
// with the app's bot token. Parameters are JSON values, like the IDs in the
// path as strings.
type DiscordAPIRequestRequest struct {
	Operation  string                     `json:"operation"`
	PathParams map[string]json.RawMessage `json:"path_params" tstype:"{ [key: string]: any }"`
	Query      map[string]json.RawMessage `json:"query" tstype:"{ [key: string]: any }"`
	Body       json.RawMessage            `json:"body" tstype:"any"`
	// The reason in the server's audit log, if the endpoint takes one.
	Reason string `json:"reason"`
}

func (req DiscordAPIRequestRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Operation, validation.Required, validation.Length(1, 100)),
		validation.Field(&req.Reason, validation.Length(0, 400)),
	)
}

type DiscordAPIRequestResponse struct {
	// The JSON the endpoint returned, null if it returned nothing.
	Result json.RawMessage `json:"result" tstype:"any"`
}
