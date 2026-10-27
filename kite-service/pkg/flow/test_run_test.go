package flow

import (
	"encoding/json"
	"testing"

	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

// Mocked results get the type of the block's result, so placeholders read
// them like real ones.
func TestTestMockResult(t *testing.T) {
	member, err := TestMockResult(FlowNodeTypeActionMemberGet, json.RawMessage(`{"user": {"id": "1", "username": "kite"}, "roles": ["2"]}`))
	if err != nil {
		t.Fatal(err)
	}
	if member.Type != thing.TypeDiscordMember || member.DiscordMember().User.Username != "kite" {
		t.Errorf("member = %+v", member)
	}

	resp, err := TestMockResult(FlowNodeTypeActionHTTPRequest, json.RawMessage(`{"status_code": 404, "body": {"error": "not found"}}`))
	if err != nil {
		t.Fatal(err)
	}
	if got := resp.HTTPResponse(); got.StatusCode != 404 || got.Status != "404 Not Found" || string(got.Body) != `{"error": "not found"}` {
		t.Errorf("response = %+v, body %s", got, got.Body)
	}

	answer, err := TestMockResult(FlowNodeTypeActionAIChatCompletion, json.RawMessage(`"Hello"`))
	if err != nil {
		t.Fatal(err)
	}
	if answer.String() != "Hello" {
		t.Errorf("answer = %+v", answer)
	}

	if _, err := TestMockResult(FlowNodeTypeActionMemberGet, json.RawMessage(`"not a member"`)); err == nil {
		t.Error("a member mock that isn't an object is accepted")
	}
}
