package eval

import (
	"context"
	"testing"

	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/pkg/webhook"
)

// testSession never connects, the placeholders in the tests don't need Discord.
func testSession() *state.State {
	return state.New("Bot test")
}

func TestWebhookEventPlaceholders(t *testing.T) {
	c := NewContextFromEvent(&webhook.Event{
		Headers: map[string]string{"x-github-event": "push"},
		Query:   map[string]string{"source": "ci"},
		Body:    `{"repository":{"name":"kite","stars":12},"commits":[{"message":"fix"}]}`,
	}, testSession())

	cases := map[string]string{
		"{{webhook.data.repository.name}}":                    "kite",
		"{{webhook.data.repository.stars}}":                   "12",
		"{{webhook.data.commits[0].message}}":                 "fix",
		"{{webhook.headers['x-github-event']}}":               "push",
		"{{webhook.query.source}}":                            "ci",
		"{{webhook.headers['x-github-event'] == 'push'}}":     "true",
		"Pushed to {{webhook.data.repository.name}} from CI!": "Pushed to kite from CI!",
	}
	for template, want := range cases {
		got, err := EvalTemplate(context.Background(), template, c)
		if err != nil {
			t.Errorf("%s: %v", template, err)
			continue
		}
		if got.String() != want {
			t.Errorf("%s = %q, want %q", template, got.String(), want)
		}
	}

	body, err := EvalTemplate(context.Background(), "{{webhook.body}}", c)
	if err != nil {
		t.Fatal(err)
	}
	if body.String() != `{"repository":{"name":"kite","stars":12},"commits":[{"message":"fix"}]}` {
		t.Errorf("webhook.body = %q", body.String())
	}
}

func TestWebhookEventPlaceholdersWithoutJSONBody(t *testing.T) {
	c := NewContextFromEvent(&webhook.Event{Body: "status=up"}, testSession())

	got, err := EvalTemplate(context.Background(), "{{webhook.body}}", c)
	if err != nil {
		t.Fatal(err)
	}
	if got.String() != "status=up" {
		t.Errorf("webhook.body = %q, want the raw body", got.String())
	}

	// A body that isn't JSON has no data, which must not fail the flow.
	got, err = EvalTemplate(context.Background(), "{{webhook.data == nil}}", c)
	if err != nil {
		t.Fatal(err)
	}
	if got.String() != "true" {
		t.Errorf("webhook.data == nil is %q, want true", got.String())
	}
}
