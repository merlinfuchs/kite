package sharedbot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/state/store"
	"github.com/diamondburned/arikawa/v3/utils/httputil/httpdriver"
	"github.com/kitecloud/kite/kite-service/internal/metrics"
	"github.com/kitecloud/kite/kite-service/internal/util"
)

// The Kite bot is in the servers of every shared app, so each app gets a view
// of it that only reaches its own server: a REST client that checks the path
// of every request, and a cache that only has the app's server. The engine
// and the API use the view like the bot of an own app.

var (
	// errOutOfScope is a request a shared app isn't allowed to send.
	errOutOfScope = errors.New("This app runs on the Kite bot and can only act in its own server.")
	// errPaused is a request of an app whose requests Discord kept refusing.
	errPaused = errors.New("Discord refused too many requests of this app, so it can't send any for a few minutes.")
	// errAllPaused is a request while Discord kept refusing the requests of
	// many apps.
	errAllPaused = errors.New("Discord refused too many requests of apps on the Kite bot, so they can't send any for a few minutes.")
)

// channelGuilds finds the server of a channel, which the paths of most
// requests only name the channel of.
type channelGuilds struct {
	// channels is the Kite bot's cache, nil in the API, which has none.
	channels store.ChannelStore
	client   *api.Client

	mu sync.Mutex
	// known are channels looked up from Discord. Channels don't move between
	// servers, so they never go stale.
	known map[discord.ChannelID]discord.GuildID
	// unknown are channels Discord didn't show the bot, until when they're
	// asked again. Asking again and again would count toward the refused
	// requests that get the IP blocked, see maxFailures. They're asked again
	// later, as the bot may be given access.
	unknown map[discord.ChannelID]time.Time
}

const (
	// maxKnownChannels bounds the channels remembered from Discord.
	maxKnownChannels  = 10_000
	unknownChannelTTL = 10 * time.Minute
)

func newChannelGuilds(channels store.ChannelStore, client *api.Client) *channelGuilds {
	return &channelGuilds{
		channels: channels,
		client:   client,
		known:    make(map[discord.ChannelID]discord.GuildID),
		unknown:  make(map[discord.ChannelID]time.Time),
	}
}

func (c *channelGuilds) guildOf(ctx context.Context, id discord.ChannelID) (discord.GuildID, error) {
	if c.channels != nil {
		if ch, err := c.channels.Channel(id); err == nil {
			return ch.GuildID, nil
		}
	}

	c.mu.Lock()
	guildID, ok := c.known[id]
	until, unknown := c.unknown[id]
	c.mu.Unlock()
	if ok {
		return guildID, nil
	}
	if c.client == nil || (unknown && time.Now().Before(until)) {
		return 0, store.ErrNotFound
	}

	ch, err := c.client.WithContext(ctx).Channel(id)
	if util.IsDiscordNotInGuild(err) {
		c.mu.Lock()
		if len(c.unknown) >= maxKnownChannels {
			clear(c.unknown)
		}
		c.unknown[id] = time.Now().Add(unknownChannelTTL)
		c.mu.Unlock()
	}
	if err != nil {
		return 0, err
	}

	c.mu.Lock()
	if len(c.known) >= maxKnownChannels {
		clear(c.known)
	}
	c.known[id] = ch.GuildID
	c.mu.Unlock()
	return ch.GuildID, nil
}

// scope is what one shared app can reach.
type scope struct {
	guildID  discord.GuildID
	botAppID discord.AppID
	channels *channelGuilds

	// failures are the app's, which outlive its scopes, and the ones of all
	// apps in the process.
	failures *failures
	all      *failures
}

const (
	// maxFailures is how many requests of one app Discord can refuse within
	// failureWindow before the app's requests are refused without sending
	// them, until failurePause is over. Discord blocks the IP of every bot on
	// the host after 10,000 refused requests in 10 minutes.
	maxFailures   = 200
	failureWindow = 10 * time.Minute
	failurePause  = 10 * time.Minute

	// Many apps could each stay just below maxFailures, so all shared apps of
	// a process are paused too once Discord refused this many of their
	// requests. The limit is per IP and the processes count on their own, so
	// together they stay below it: the Kite bot's process, which sends nearly
	// all requests, and each of the 4 API clusters of kite.onl. More clusters
	// on one host need a lower maxAPIFailures.
	maxGatewayFailures = 4_000
	maxAPIFailures     = 1_000
)

// appFailures are the failures of each app, so they add up over the scopes
// of the app, like one per API request, and of all apps of the process.
type appFailures struct {
	mu    sync.Mutex
	byApp map[string]*failures
	all   *failures
}

// newAppFailures pauses all apps after limit refused requests.
func newAppFailures(limit int) *appFailures {
	return &appFailures{byApp: make(map[string]*failures), all: &failures{max: limit}}
}

func (a *appFailures) of(appID string) *failures {
	a.mu.Lock()
	defer a.mu.Unlock()

	f, ok := a.byApp[appID]
	if !ok {
		f = &failures{max: maxFailures}
		a.byApp[appID] = f
	}
	return f
}

// failures counts the requests Discord refused, and pauses after max of them
// within failureWindow.
type failures struct {
	max int

	mu          sync.Mutex
	windowStart time.Time
	count       int
	pausedUntil time.Time
}

func (f *failures) paused(now time.Time) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return now.Before(f.pausedUntil)
}

// add counts a refused request and reports whether that paused the app.
func (f *failures) add(now time.Time) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	if now.Sub(f.windowStart) > failureWindow {
		f.windowStart = now
		f.count = 0
	}
	f.count++
	if f.count < f.max {
		return false
	}

	f.pausedUntil = now.Add(failurePause)
	f.windowStart = time.Time{}
	return true
}

// check returns an error for a request outside the app's server. Paths are
// Discord's routes after the API version, like /channels/123/messages. Unknown
// routes are refused.
func (s *scope) check(ctx context.Context, method, path string) error {
	segments := util.DiscordAPISegments(path)
	if len(segments) == 0 {
		return errOutOfScope
	}

	at := func(i int) string {
		if i < len(segments) {
			return segments[i]
		}
		return ""
	}
	isOwnGuild := func(id string) bool {
		return id == s.guildID.String()
	}
	isBot := func(id string) bool {
		return id == "@me" || id == s.botAppID.String()
	}

	allowed := false
	switch at(0) {
	case "guilds":
		allowed = isOwnGuild(at(1))
	case "channels":
		id, err := discord.ParseSnowflake(at(1))
		if err != nil {
			return errOutOfScope
		}
		guildID, err := s.channels.guildOf(ctx, discord.ChannelID(id))
		if err != nil {
			// Unknown channels are refused, as Discord would.
			return errOutOfScope
		}
		allowed = guildID == s.guildID
	case "webhooks":
		// Only with the webhook's token, which authorizes it instead of the
		// bot, like interaction followups and messages with a persona.
		allowed = at(2) != ""
	case "interactions":
		allowed = at(2) != ""
	case "users":
		switch {
		case at(1) == "@me" && len(segments) == 2:
			allowed = method == http.MethodGet
		case at(1) == "@me" && at(2) == "guilds":
			// Leaving the server, which removes the app from it.
			allowed = method == http.MethodDelete && isOwnGuild(at(3)) && len(segments) == 4
		case at(1) != "@me" && len(segments) == 2:
			allowed = method == http.MethodGet
		}
	case "applications":
		switch {
		case isBot(at(1)) && len(segments) == 2:
			allowed = method == http.MethodGet
		case isBot(at(1)) && at(2) == "guilds":
			allowed = isOwnGuild(at(3))
		}
	case "stickers", "sticker-packs", "invites", "voice", "gateway":
		allowed = method == http.MethodGet
	}

	if !allowed {
		metrics.SharedBotDenied.Add(at(0), 1)
		return errOutOfScope
	}
	return nil
}

// scopedClient is client, whose rate limits it shares, with every request
// checked against the scope before it's sent.
func scopedClient(client *api.Client, s *scope) *api.Client {
	hc := client.Client.Copy()
	hc.Client = scopedDriver{Client: hc.Client, scope: s}
	return &api.Client{Client: hc, Session: client.Session, AcquireOptions: client.AcquireOptions}
}

// scopedDriver answers requests outside the scope with a 403 instead of
// sending them. Unlike an error, a response isn't retried.
type scopedDriver struct {
	httpdriver.Client
	scope *scope
}

func (d scopedDriver) Do(r httpdriver.Request) (httpdriver.Response, error) {
	req, ok := r.(*httpdriver.DefaultRequest)
	if !ok {
		return refused(errOutOfScope), nil
	}
	hr := (*http.Request)(req)

	if err := d.scope.check(hr.Context(), hr.Method, hr.URL.Path); err != nil {
		return refused(err), nil
	}
	if err := d.checkBody(hr); err != nil {
		return refused(err), nil
	}

	now := time.Now()
	if d.scope.failures.paused(now) {
		return refused(errPaused), nil
	}
	if d.scope.all.paused(now) {
		return refused(errAllPaused), nil
	}

	resp, err := d.Client.Do(r)
	if err != nil || !isRefusal(resp) {
		return resp, err
	}
	if d.scope.failures.add(now) {
		slog.Warn(
			"Paused the requests of a shared app that Discord kept refusing",
			slog.String("guild_id", d.scope.guildID.String()),
		)
	}
	if d.scope.all.add(now) {
		slog.Error("Paused the requests of all shared apps, Discord refused too many of them")
	}
	return resp, err
}

// isRefusal is whether Discord counts the response toward its limit of
// refused requests. A 429 of a limit shared with others, like of a busy
// channel, doesn't count.
func isRefusal(resp httpdriver.Response) bool {
	switch resp.GetStatus() {
	case http.StatusUnauthorized, http.StatusForbidden:
		return true
	case http.StatusTooManyRequests:
		return resp.GetHeader().Get("X-RateLimit-Scope") != "shared"
	}
	return false
}

// checkBody refuses messages that reference a message in another server,
// like a forward, which would copy it into this one.
func (d scopedDriver) checkBody(hr *http.Request) error {
	segments := util.DiscordAPISegments(hr.URL.Path)
	if hr.Body == nil || hr.Method != http.MethodPost || len(segments) != 3 || segments[0] != "channels" || segments[2] != "messages" {
		return nil
	}

	body, err := io.ReadAll(hr.Body)
	hr.Body.Close()
	if err != nil {
		return err
	}
	hr.Body = io.NopCloser(bytes.NewReader(body))

	payload := body
	if mediaType, params, err := mime.ParseMediaType(hr.Header.Get("Content-Type")); err == nil && strings.HasPrefix(mediaType, "multipart/") {
		payload = multipartPayload(body, params["boundary"])
	}

	var msg struct {
		MessageReference *struct {
			ChannelID discord.ChannelID `json:"channel_id"`
			GuildID   discord.GuildID   `json:"guild_id"`
		} `json:"message_reference"`
	}
	// A message with only files has no JSON. JSON that isn't understood is
	// refused, Discord might read a reference from it.
	if len(payload) == 0 {
		return nil
	}
	if err := json.Unmarshal(payload, &msg); err != nil {
		return errOutOfScope
	}
	if msg.MessageReference == nil {
		return nil
	}

	ref := msg.MessageReference
	if ref.GuildID.IsValid() && ref.GuildID != d.scope.guildID {
		return errOutOfScope
	}
	if ref.ChannelID.IsValid() {
		guildID, err := d.scope.channels.guildOf(hr.Context(), ref.ChannelID)
		if err != nil || guildID != d.scope.guildID {
			return errOutOfScope
		}
	}
	return nil
}

// multipartPayload returns the JSON part of a message with files.
func multipartPayload(body []byte, boundary string) []byte {
	reader := multipart.NewReader(bytes.NewReader(body), boundary)
	for {
		part, err := reader.NextPart()
		if err != nil {
			return nil
		}
		if part.FormName() == "payload_json" {
			payload, _ := io.ReadAll(part)
			return payload
		}
	}
}

// refused is the response to a request outside the scope, which reads like
// Discord's own refusals.
func refused(err error) httpdriver.Response {
	body, _ := json.Marshal(map[string]any{"message": err.Error(), "code": 0})
	return refusedResponse{body: body}
}

type refusedResponse struct {
	body []byte
}

func (r refusedResponse) GetStatus() int { return http.StatusForbidden }
func (r refusedResponse) GetHeader() http.Header {
	return http.Header{"Content-Type": {"application/json"}}
}
func (r refusedResponse) GetBody() io.ReadCloser { return io.NopCloser(bytes.NewReader(r.body)) }

// scopedState is the Kite bot's session as the app of one server sees it. It
// shares the bot's connection, cache and rate limits.
func scopedState(shared *state.State, s *scope) *state.State {
	session := *shared.Session
	session.Client = scopedClient(shared.Session.Client, s)

	scoped := *shared
	scoped.Session = &session
	scoped.Cabinet = scopedCabinet(shared.Cabinet, s.guildID)
	return &scoped
}

func scopedCabinet(c *store.Cabinet, guildID discord.GuildID) *store.Cabinet {
	return &store.Cabinet{
		MeStore:         c.MeStore,
		ChannelStore:    scopedChannels{c.ChannelStore, guildID},
		EmojiStore:      scopedEmojis{c.EmojiStore, guildID},
		GuildStore:      scopedGuilds{c.GuildStore, guildID},
		MemberStore:     scopedMembers{c.MemberStore, guildID},
		MessageStore:    scopedMessages{c.MessageStore, c.ChannelStore, guildID},
		PresenceStore:   scopedPresences{c.PresenceStore, guildID},
		RoleStore:       scopedRoles{c.RoleStore, guildID},
		VoiceStateStore: scopedVoiceStates{c.VoiceStateStore, guildID},
	}
}

// The scoped stores only change what they return. Everything else, like
// updates from requests, goes to the bot's cache.

type scopedChannels struct {
	store.ChannelStore
	guildID discord.GuildID
}

func (s scopedChannels) Channel(id discord.ChannelID) (*discord.Channel, error) {
	ch, err := s.ChannelStore.Channel(id)
	if err != nil {
		return nil, err
	}
	if ch.GuildID != s.guildID {
		return nil, store.ErrNotFound
	}
	return ch, nil
}

func (s scopedChannels) CreatePrivateChannel(discord.UserID) (*discord.Channel, error) {
	return nil, store.ErrNotFound
}

func (s scopedChannels) Channels(guildID discord.GuildID) ([]discord.Channel, error) {
	if guildID != s.guildID {
		return nil, store.ErrNotFound
	}
	return s.ChannelStore.Channels(guildID)
}

func (s scopedChannels) PrivateChannels() ([]discord.Channel, error) {
	return nil, store.ErrNotFound
}

type scopedEmojis struct {
	store.EmojiStore
	guildID discord.GuildID
}

func (s scopedEmojis) Emoji(guildID discord.GuildID, id discord.EmojiID) (*discord.Emoji, error) {
	if guildID != s.guildID {
		return nil, store.ErrNotFound
	}
	return s.EmojiStore.Emoji(guildID, id)
}

func (s scopedEmojis) Emojis(guildID discord.GuildID) ([]discord.Emoji, error) {
	if guildID != s.guildID {
		return nil, store.ErrNotFound
	}
	return s.EmojiStore.Emojis(guildID)
}

type scopedGuilds struct {
	store.GuildStore
	guildID discord.GuildID
}

func (s scopedGuilds) Guild(id discord.GuildID) (*discord.Guild, error) {
	if id != s.guildID {
		return nil, store.ErrNotFound
	}
	return s.GuildStore.Guild(id)
}

func (s scopedGuilds) Guilds() ([]discord.Guild, error) {
	guild, err := s.GuildStore.Guild(s.guildID)
	if err != nil {
		return nil, err
	}
	return []discord.Guild{*guild}, nil
}

type scopedMembers struct {
	store.MemberStore
	guildID discord.GuildID
}

func (s scopedMembers) Member(guildID discord.GuildID, userID discord.UserID) (*discord.Member, error) {
	if guildID != s.guildID {
		return nil, store.ErrNotFound
	}
	return s.MemberStore.Member(guildID, userID)
}

func (s scopedMembers) Members(guildID discord.GuildID) ([]discord.Member, error) {
	if guildID != s.guildID {
		return nil, store.ErrNotFound
	}
	return s.MemberStore.Members(guildID)
}

type scopedMessages struct {
	store.MessageStore
	channels store.ChannelStore
	guildID  discord.GuildID
}

func (s scopedMessages) inGuild(channelID discord.ChannelID) bool {
	ch, err := s.channels.Channel(channelID)
	return err == nil && ch.GuildID == s.guildID
}

func (s scopedMessages) Message(channelID discord.ChannelID, id discord.MessageID) (*discord.Message, error) {
	if !s.inGuild(channelID) {
		return nil, store.ErrNotFound
	}
	return s.MessageStore.Message(channelID, id)
}

func (s scopedMessages) Messages(channelID discord.ChannelID) ([]discord.Message, error) {
	if !s.inGuild(channelID) {
		return nil, store.ErrNotFound
	}
	return s.MessageStore.Messages(channelID)
}

type scopedPresences struct {
	store.PresenceStore
	guildID discord.GuildID
}

func (s scopedPresences) Presence(guildID discord.GuildID, userID discord.UserID) (*discord.Presence, error) {
	if guildID != s.guildID {
		return nil, store.ErrNotFound
	}
	return s.PresenceStore.Presence(guildID, userID)
}

func (s scopedPresences) Presences(guildID discord.GuildID) ([]discord.Presence, error) {
	if guildID != s.guildID {
		return nil, store.ErrNotFound
	}
	return s.PresenceStore.Presences(guildID)
}

type scopedRoles struct {
	store.RoleStore
	guildID discord.GuildID
}

func (s scopedRoles) Role(guildID discord.GuildID, id discord.RoleID) (*discord.Role, error) {
	if guildID != s.guildID {
		return nil, store.ErrNotFound
	}
	return s.RoleStore.Role(guildID, id)
}

func (s scopedRoles) Roles(guildID discord.GuildID) ([]discord.Role, error) {
	if guildID != s.guildID {
		return nil, store.ErrNotFound
	}
	return s.RoleStore.Roles(guildID)
}

type scopedVoiceStates struct {
	store.VoiceStateStore
	guildID discord.GuildID
}

func (s scopedVoiceStates) VoiceState(guildID discord.GuildID, userID discord.UserID) (*discord.VoiceState, error) {
	if guildID != s.guildID {
		return nil, store.ErrNotFound
	}
	return s.VoiceStateStore.VoiceState(guildID, userID)
}

func (s scopedVoiceStates) VoiceStates(guildID discord.GuildID) ([]discord.VoiceState, error) {
	if guildID != s.guildID {
		return nil, store.ErrNotFound
	}
	return s.VoiceStateStore.VoiceStates(guildID)
}
