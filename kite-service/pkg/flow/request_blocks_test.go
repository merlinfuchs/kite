package flow

import (
	"context"
	"encoding/json"
	"io"
	"regexp"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/stretchr/testify/require"
)

// sentRequest is what a request block sent, for the golden file.
type sentRequest struct {
	Method string `json:"method,omitempty"`
	// Path for Discord, the whole URL for other integrations.
	URL    string `json:"url,omitempty"`
	Body   string `json:"body,omitempty"`
	Reason string `json:"reason,omitempty"`
	Error  string `json:"error,omitempty"`
}

// sampleSetting is a value for a field of a request block that it accepts.
func sampleSetting(f blockField) any {
	switch f.Type {
	case "snowflake":
		return "11"
	case "snowflake_list":
		return "11, 12"
	case "integer", "seconds", "seconds_until":
		n := int64(60)
		if f.Max != nil && n > *f.Max {
			n = *f.Max
		}
		if f.Min != nil && n < *f.Min {
			n = *f.Min
		}
		return strconv.FormatInt(n, 10)
	case "color":
		return "#ff0000"
	case "boolean":
		return true
	case "emoji":
		return map[string]any{"name": "👍"}
	case "json_object":
		return `{"note":"text"}`
	}
	if len(f.Options) > 0 {
		return f.Options[0]
	}
	if f.MaxLength != nil && *f.MaxLength < 4 {
		return "t"
	}
	return "text"
}

var isoTimeRe = regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})`)

// Every request block with every setting filled in sends the request in
// testdata/request_blocks.json. Run with -update after a change to a block
// that is meant to change what it sends.
func TestRequestBlocksSendSameRequests(t *testing.T) {
	var types []FlowNodeType
	for nodeType, block := range blockDefinitions {
		if block.Run.Kind == "request" {
			types = append(types, nodeType)
		}
	}
	slices.Sort(types)

	credentials := map[string]string{}
	for id, integration := range integrations {
		if integration.NeedsCredential() {
			credentials[id] = "cr3dential"
		}
	}

	got := map[FlowNodeType]sentRequest{}
	for _, nodeType := range types {
		block := blockDefinitions[nodeType]
		settings := map[string]any{}
		for _, f := range block.Fields {
			settings[f.Name] = sampleSetting(f)
		}
		if block.AuditLogReason {
			settings["audit_log_reason"] = "reason"
		}
		raw, err := json.Marshal(settings)
		require.NoError(t, err)
		var data FlowNodeData
		require.NoError(t, json.Unmarshal(raw, &data), nodeType)

		response := "{}"
		if block.Result != nil && block.Result.List {
			response = "[]"
		}
		discordProvider := &blockTestProvider{response: response}
		httpProvider := &httpTestProvider{}
		choices := map[string]bool{}
		for id := range integrations {
			choices[id] = true
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		c := NewContext(ctx, 5*time.Second, &blockTestContextData{}, FlowProviders{
			Discord:     discordProvider,
			HTTP:        httpProvider,
			Log:         &provider.MockLogProvider{},
			Integration: &integrationTestProvider{credentials: credentials, choices: choices},
		}, FlowContextLimits{MaxStackDepth: 10, MaxOperations: 1000, MaxCredits: 1000}, eval.NewContext(eval.Env{}), nil)
		node := &CompiledFlowNode{
			ID:   "0",
			Type: FlowNodeTypeEntryEvent,
			Children: ConnectedFlowNodes{Default: []*CompiledFlowNode{
				{ID: "1", Type: nodeType, Data: data},
			}},
		}
		err = node.Execute(c)
		c.Cancel()
		cancel()

		var sent sentRequest
		switch {
		case discordProvider.req.Method != "":
			sent = sentRequest{
				Method: discordProvider.req.Method,
				URL:    discordProvider.req.Path,
				Body:   string(discordProvider.req.Body),
				Reason: string(discordProvider.req.Reason),
			}
		case httpProvider.req != nil:
			sent = sentRequest{Method: httpProvider.req.Method, URL: httpProvider.req.URL.String()}
			if httpProvider.req.Body != nil {
				body, _ := io.ReadAll(httpProvider.req.Body)
				sent.Body = string(body)
			}
		}
		if err != nil {
			sent.Error = err.Error()
		}
		sent.Body = isoTimeRe.ReplaceAllString(sent.Body, "<time>")
		got[nodeType] = sent
	}

	b, err := json.MarshalIndent(got, "", "  ")
	require.NoError(t, err)
	checkGeneratedFile(t, "testdata/request_blocks.json", append(b, '\n'))
}
