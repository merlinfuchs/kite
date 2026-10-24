package eval

import (
	"context"
	"testing"

	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/pkg/webhook"
)

// emptyTestSession never connects, the placeholders in the tests don't need Discord.
func emptyTestSession() *state.State {
	return state.New("Bot test")
}

func TestWebhookEventPlaceholders(t *testing.T) {
	c := NewContextFromEvent(&webhook.Event{
		Headers: map[string]string{"x-github-event": "push"},
		Query:   map[string]string{"source": "ci"},
		Body:    `{"repository":{"name":"kite","stars":12},"commits":[{"message":"fix"}]}`,
	}, emptyTestSession())

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

func TestWebhookEventPlaceholdersKeepLargeIntegers(t *testing.T) {
	c := NewContextFromEvent(&webhook.Event{
		Body: `{"user_id":1497746534387941386,"ids":[1497746534387941386],"count":3,"ratio":0.5}`,
	}, emptyTestSession())

	cases := map[string]string{
		"{{webhook.data.user_id}}":                        "1497746534387941386",
		"{{webhook.data.ids[0]}}":                         "1497746534387941386",
		"{{webhook.data.user_id == 1497746534387941386}}": "true",
		"{{webhook.data.count + 1}}":                      "4",
		"{{webhook.data.count > 2}}":                      "true",
		"{{webhook.data.ratio * 2}}":                      "1",
		"{{webhook.data.ratio}}":                          "0.5",
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
}

func TestWebhookEventPlaceholdersWithoutJSONBody(t *testing.T) {
	c := NewContextFromEvent(&webhook.Event{Body: "status=up"}, emptyTestSession())

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
