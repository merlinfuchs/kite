package sharedbot

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state/store"
	"github.com/diamondburned/arikawa/v3/state/store/defaultstore"
	"github.com/diamondburned/arikawa/v3/utils/httputil/httpdriver"
)

const (
	ownGuild     = discord.GuildID(100)
	otherGuild   = discord.GuildID(200)
	ownChannel   = discord.ChannelID(101)
	otherChannel = discord.ChannelID(201)
	botApp       = discord.AppID(900)
)

func testScope(t *testing.T) *scope {
	t.Helper()

	channels := defaultstore.NewChannel()
	for _, ch := range []discord.Channel{
		{ID: ownChannel, GuildID: ownGuild},
		{ID: otherChannel, GuildID: otherGuild},
	} {
		if err := channels.ChannelSet(&ch, false); err != nil {
			t.Fatal(err)
		}
	}

	return &scope{guildID: ownGuild, botAppID: botApp, channels: newChannelGuilds(channels, nil), failures: &failures{}}
}

func TestScopeCheck(t *testing.T) {
	s := testScope(t)

	tests := []struct {
		method string
		path   string
		allow  bool
	}{
		{"GET", "/api/v10/guilds/100", true},
		{"PUT", "/api/v10/guilds/100/members/5/roles/6", true},
		{"PATCH", "/api/v10/guilds/100/members/@me", true},
		{"GET", "/api/v10/guilds/200", false},
		{"PUT", "/api/v10/guilds/200/bans/5", false},
		{"POST", "/api/v10/guilds", false},

		{"POST", "/api/v10/channels/101/messages", true},
		{"DELETE", "/api/v10/channels/101/messages/7", true},
		{"POST", "/api/v10/channels/201/messages", false},
		// Unknown channels can't be told apart from other servers'.
		{"GET", "/api/v10/channels/301", false},
		{"GET", "/api/v10/channels/abc", false},

		{"GET", "/api/v10/users/@me", true},
		{"PATCH", "/api/v10/users/@me", false},
		{"GET", "/api/v10/users/@me/guilds", false},
		{"POST", "/api/v10/users/@me/channels", false},
		{"DELETE", "/api/v10/users/@me/guilds/100", true},
		{"DELETE", "/api/v10/users/@me/guilds/200", false},
		{"GET", "/api/v10/users/5", true},
		{"PATCH", "/api/v10/users/5", false},

		{"GET", "/api/v10/applications/@me", true},
		{"PATCH", "/api/v10/applications/@me", false},
		{"PUT", "/api/v10/applications/900/guilds/100/commands", true},
		{"PUT", "/api/v10/applications/900/guilds/200/commands", false},
		{"PUT", "/api/v10/applications/900/commands", false},
		{"GET", "/api/v10/applications/900/emojis", false},
		{"PUT", "/api/v10/applications/901/guilds/100/commands", false},

		{"POST", "/api/v10/webhooks/900/token", true},
		{"PATCH", "/api/v10/webhooks/900/token/messages/@original", true},
		{"GET", "/api/v10/webhooks/5", false},
		{"POST", "/api/v10/interactions/5/token/callback", true},

		{"GET", "/api/v10/gateway/bot", true},
		{"GET", "/api/v10/invites/abc", true},
		{"DELETE", "/api/v10/invites/abc", false},
		{"GET", "/api/v10/oauth2/applications/@me", false},
		{"GET", "/", false},
	}

	for _, tt := range tests {
		err := s.check(context.Background(), tt.method, tt.path)
		if (err == nil) != tt.allow {
			t.Errorf("%s %s: allowed = %v, want %v", tt.method, tt.path, err == nil, tt.allow)
		}
	}
}

// fakeDriver records the requests that reach Discord.
type fakeDriver struct {
	httpdriver.Client
	sent []string
}

func (d *fakeDriver) Do(r httpdriver.Request) (httpdriver.Response, error) {
	hr := (*http.Request)(r.(*httpdriver.DefaultRequest))
	d.sent = append(d.sent, hr.Method+" "+hr.URL.Path)
	return refusedResponse{body: []byte("{}")}, nil
}

func doRequest(t *testing.T, s *scope, method, path, contentType string, body []byte) (int, *fakeDriver) {
	t.Helper()

	inner := &fakeDriver{Client: httpdriver.NewClient()}
	driver := scopedDriver{Client: inner, scope: s}

	req, err := driver.NewRequest(context.Background(), method, "https://discord.com"+path)
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		req.WithBody(io.NopCloser(bytes.NewReader(body)))
		req.AddHeader(http.Header{"Content-Type": {contentType}})
	}

	resp, err := driver.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp.GetStatus(), inner
}

func TestScopedDriverRefusesWithoutSending(t *testing.T) {
	s := testScope(t)

	status, inner := doRequest(t, s, "POST", "/api/v10/channels/201/messages", "application/json", []byte(`{"content":"hi"}`))
	if status != http.StatusForbidden || len(inner.sent) != 0 {
		t.Errorf("other server: status %d, sent %v, want 403 and nothing sent", status, inner.sent)
	}

	_, inner = doRequest(t, s, "POST", "/api/v10/channels/101/messages", "application/json", []byte(`{"content":"hi"}`))
	if len(inner.sent) != 1 {
		t.Errorf("own server: sent %v, want the request", inner.sent)
	}
}

func TestScopedDriverChecksMessageReferences(t *testing.T) {
	s := testScope(t)

	tests := []struct {
		name string
		body string
		sent bool
	}{
		{"forward from other server", `{"message_reference":{"type":1,"channel_id":"201","message_id":"5"}}`, false},
		{"reference with other server", `{"message_reference":{"guild_id":"200","channel_id":"101","message_id":"5"}}`, false},
		{"forward from unknown channel", `{"message_reference":{"type":1,"channel_id":"301","message_id":"5"}}`, false},
		{"reply in own server", `{"content":"hi","message_reference":{"channel_id":"101","message_id":"5"}}`, true},
		{"reply without channel", `{"content":"hi","message_reference":{"message_id":"5"}}`, true},
		{"no reference", `{"content":"hi"}`, true},
		{"forward with an escaped key", `{"message\u005freference":{"type":1,"channel_id":"201","message_id":"5"}}`, false},
		{"JSON that isn't understood", `{"message_reference":{"channel_id":201}}`, false},
	}

	for _, tt := range tests {
		_, inner := doRequest(t, s, "POST", "/api/v10/channels/101/messages", "application/json", []byte(tt.body))
		if (len(inner.sent) == 1) != tt.sent {
			t.Errorf("%s: sent = %v, want %v", tt.name, len(inner.sent) == 1, tt.sent)
		}
	}

	// Messages with files have their JSON in a part.
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, _ := w.CreateFormField("payload_json")
	part.Write([]byte(`{"message_reference":{"type":1,"channel_id":"201","message_id":"5"}}`))
	file, _ := w.CreateFormFile("files[0]", "a.txt")
	file.Write([]byte("hello"))
	w.Close()

	_, inner := doRequest(t, s, "POST", "/api/v10/channels/101/messages", w.FormDataContentType(), body.Bytes())
	if len(inner.sent) != 0 {
		t.Error("forward from other server with files was sent")
	}
}

func TestScopedCabinet(t *testing.T) {
	cabinet := defaultstore.New()
	for _, g := range []discord.Guild{{ID: ownGuild}, {ID: otherGuild}} {
		cabinet.GuildSet(&g, false)
	}
	cabinet.ChannelSet(&discord.Channel{ID: ownChannel, GuildID: ownGuild}, false)
	cabinet.ChannelSet(&discord.Channel{ID: otherChannel, GuildID: otherGuild}, false)
	cabinet.RoleSet(otherGuild, &discord.Role{ID: 7}, false)

	scoped := scopedCabinet(cabinet, ownGuild)

	if _, err := scoped.Channel(ownChannel); err != nil {
		t.Errorf("own channel: %v", err)
	}
	if _, err := scoped.Channel(otherChannel); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("other server's channel: err = %v, want not found", err)
	}
	if _, err := scoped.Guild(otherGuild); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("other server: err = %v, want not found", err)
	}
	if _, err := scoped.Role(otherGuild, 7); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("other server's role: err = %v, want not found", err)
	}

	guilds, err := scoped.Guilds()
	if err != nil || len(guilds) != 1 || guilds[0].ID != ownGuild {
		t.Errorf("guilds = %v, %v, want only the own server", guilds, err)
	}
}

// Discord answers every request of fakeDriver with a 403, so after
// maxFailures the app's requests aren't sent anymore.
func TestScopedDriverPausesAfterFailures(t *testing.T) {
	s := testScope(t)
	inner := &fakeDriver{Client: httpdriver.NewClient()}
	driver := scopedDriver{Client: inner, scope: s}

	for range maxFailures + 5 {
		req, err := driver.NewRequest(context.Background(), "GET", "https://discord.com/api/v10/guilds/100")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := driver.Do(req); err != nil {
			t.Fatal(err)
		}
	}

	if len(inner.sent) != maxFailures {
		t.Errorf("sent %d requests, want %d", len(inner.sent), maxFailures)
	}
}
