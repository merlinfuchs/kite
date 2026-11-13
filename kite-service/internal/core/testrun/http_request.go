package testrun

import (
	"context"

	"github.com/kitecloud/kite/kite-service/internal/core/engine"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
)

// SendHTTPRequest sends the request of a Send API Request block from the
// editor, with the app's real secrets, see flow.SendTestHTTPRequest. Unlike
// test runs it doesn't need the app to be online.
func (r *Runner) SendHTTPRequest(ctx context.Context, appID string, data flow.HTTPRequestData, values map[string]string) flow.TestHTTPRequestResult {
	return flow.SendTestHTTPRequest(ctx, data, values, flow.FlowProviders{
		HTTP:   engine.NewHTTPProvider(r.env.HttpClient),
		Secret: engine.NewSecretProvider(appID, r.env.AppSecretStore, r.env.TokenCrypt),
	})
}
