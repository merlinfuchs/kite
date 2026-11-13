package testrun

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/guildevent"
	"github.com/kitecloud/kite/kite-service/pkg/schedule"
	"github.com/kitecloud/kite/kite-service/pkg/webhook"
)

// interactionEvent builds an interaction like Discord sends it, from the user
// and the channel. override is merged into it.
func (s *testSession) interactionEvent(
	userID discord.UserID,
	guildID discord.GuildID,
	channelID discord.ChannelID,
	t discord.InteractionDataType,
	data map[string]any,
	msg *discord.Message,
	override json.RawMessage,
) (gateway.Event, error) {
	user, err := s.user(userID)
	if err != nil {
		return nil, testErrorf("user %s doesn't exist", userID)
	}

	appID := s.session.Ready().Application.ID
	if !appID.IsValid() {
		appID = discord.AppID(s.botID)
	}

	i := map[string]any{
		"id":             flow.TestSnowflake().String(),
		"application_id": appID.String(),
		"type":           t,
		"token":          "test-run",
		"version":        1,
		"channel_id":     channelID.String(),
		"locale":         "en-US",
		"data":           data,
		"entitlements":   []any{},
	}
	if msg != nil {
		i["message"] = msg
	}

	channel, ok := s.fakeChannel(channelID)
	if !ok {
		if c, err := s.session.Channel(channelID); err == nil {
			channel, ok = *c, true
		}
	}
	if ok {
		i["channel"] = channel
	}

	if guildID.IsValid() {
		member, err := s.member(guildID, userID)
		if err != nil {
			return nil, testErrorf("user %s isn't a member of server %s", userID, guildID)
		}
		m := *member
		m.User = *user
		i["guild_id"] = guildID.String()
		i["member"] = m
		i["context"] = discord.InteractionContextGuild
		i["authorizing_integration_owners"] = map[string]string{"0": guildID.String()}
		if g := s.testGuild(guildID); g.problem == "" {
			i["guild_locale"] = g.guild.PreferredLocale
			if ok {
				i["app_permissions"] = strconv.FormatUint(uint64(g.permissions(channel)), 10)
			}
		}
	} else {
		i["user"] = user
		i["context"] = discord.InteractionContextBotDM
		i["authorizing_integration_owners"] = map[string]string{"1": userID.String()}
	}

	var event discord.InteractionEvent
	if err := mergeJSON(i, override, &event); err != nil {
		return nil, testErrorf("invalid interaction: %v", err)
	}
	return &gateway.InteractionCreateEvent{InteractionEvent: event}, nil
}

// commandEvent builds the interaction of the command with the arguments of
// the test.
func (s *testSession) commandEvent(root *flow.CompiledFlowNode, test Test) (gateway.Event, error) {
	if root.CommandType() != discord.ChatInputCommand {
		return s.contextCommandEvent(root, test)
	}

	var args []map[string]any
	for _, arg := range root.CommandArguments() {
		b, err := json.Marshal(arg)
		if err != nil {
			return nil, err
		}
		var a map[string]any
		if err := json.Unmarshal(b, &a); err != nil {
			return nil, err
		}
		args = append(args, a)
	}

	names := make([]string, len(args))
	for i, arg := range args {
		names[i], _ = arg["name"].(string)
	}
	for name := range test.Options {
		if !slices.Contains(names, name) {
			if len(names) == 0 {
				return nil, testErrorf("the command has no arguments, got %q", name)
			}
			return nil, testErrorf("the command has no argument %q, it has: %s", name, strings.Join(names, ", "))
		}
	}

	resolved := make(map[string]any)
	options := []map[string]any{}
	for _, arg := range args {
		name, _ := arg["name"].(string)
		raw, ok := test.Options[name]
		if !ok || string(raw) == "null" {
			if required, _ := arg["required"].(bool); required {
				return nil, testErrorf("argument %q is required", name)
			}
			continue
		}

		typ, _ := arg["type"].(float64)
		value, err := s.commandOptionValue(arg, discord.CommandOptionType(typ), raw, resolved)
		if err != nil {
			return nil, testErrorf("argument %q: %v", name, err)
		}
		options = append(options, map[string]any{"name": name, "type": typ, "value": value})
	}

	// Subcommands are the words after the first in the name.
	parts := strings.Split(root.CommandName(), " ")
	var top any = options
	switch len(parts) {
	case 2:
		top = []any{map[string]any{"type": discord.SubcommandOptionType, "name": parts[1], "options": options}}
	case 3:
		top = []any{map[string]any{"type": discord.SubcommandGroupOptionType, "name": parts[1], "options": []any{
			map[string]any{"type": discord.SubcommandOptionType, "name": parts[2], "options": options},
		}}}
	}

	data := map[string]any{
		"id":       flow.TestSnowflake().String(),
		"name":     parts[0],
		"type":     discord.ChatInputCommand,
		"options":  top,
		"resolved": resolved,
	}
	if s.guildID.IsValid() {
		data["guild_id"] = s.guildID.String()
	}
	return s.interactionEvent(s.userID, s.guildID, s.channelID, discord.CommandInteractionType, data, nil, test.Event)
}

// contextCommandEvent builds the interaction of a User or Message command. Its
// target is the user or message with the ID of the test's "target" option.
// By default it's the user who runs it, or a message from them in the channel
// that only exists in the test.
func (s *testSession) contextCommandEvent(root *flow.CompiledFlowNode, test Test) (gateway.Event, error) {
	var targetID string
	for name, raw := range test.Options {
		if name != "target" {
			return nil, testErrorf("only slash commands have arguments, got %q, use target for the user or message", name)
		}
		if err := json.Unmarshal(raw, &targetID); err != nil {
			return nil, testErrorf("target must be an ID")
		}
	}

	resolved := make(map[string]any)
	if root.CommandType() == discord.UserCommand {
		if targetID == "" {
			targetID = s.userID.String()
		}
		if err := s.resolveSelect(resolved, int(discord.UserSelectComponentType), []string{targetID}); err != nil {
			return nil, testErrorf("target: %v", err)
		}
	} else {
		msg, err := s.targetMessage(targetID)
		if err != nil {
			return nil, err
		}
		targetID = msg.ID.String()
		addResolved(resolved, "messages", targetID, msg)
	}

	data := map[string]any{
		"id":        flow.TestSnowflake().String(),
		"name":      root.CommandName(),
		"type":      root.CommandType(),
		"target_id": targetID,
		"resolved":  resolved,
	}
	if s.guildID.IsValid() {
		data["guild_id"] = s.guildID.String()
	}
	return s.interactionEvent(s.userID, s.guildID, s.channelID, discord.CommandInteractionType, data, nil, test.Event)
}

// targetMessage is the message a Message command is used on in a test.
func (s *testSession) targetMessage(id string) (*discord.Message, error) {
	if id != "" {
		messageID, err := discord.ParseSnowflake(id)
		if err != nil {
			return nil, testErrorf("target must be the ID of a message")
		}
		if msg, ok := s.fakeMessage(discord.MessageID(messageID)); ok {
			return &msg, nil
		}
		msg, err := s.session.Message(s.channelID, discord.MessageID(messageID))
		if err != nil {
			return nil, testErrorf("message %s isn't in channel %s or the bot can't see it", id, s.channelID)
		}
		return msg, nil
	}

	user, err := s.user(s.userID)
	if err != nil {
		return nil, testErrorf("user %s doesn't exist", s.userID)
	}
	return &discord.Message{
		ID:        discord.MessageID(flow.TestSnowflake()),
		ChannelID: s.channelID,
		GuildID:   s.guildID,
		Author:    *user,
		Content:   "Test message",
		Timestamp: discord.NewTimestamp(time.Now()),
		Type:      discord.DefaultMessage,
	}, nil
}

func (s *testSession) commandOptionValue(arg map[string]any, t discord.CommandOptionType, raw json.RawMessage, resolved map[string]any) (any, error) {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, err
	}

	switch t {
	case discord.StringOptionType:
		str, ok := v.(string)
		if !ok {
			return nil, errors.New("must be text")
		}
		if max, ok := arg["max_length"].(float64); ok && len([]rune(str)) > int(max) {
			return nil, fmt.Errorf("must be at most %d characters", int(max))
		}
		if err := checkChoice(arg, str); err != nil {
			return nil, err
		}
		return str, nil
	case discord.IntegerOptionType, discord.NumberOptionType:
		f, ok := v.(float64)
		if str, isStr := v.(string); isStr {
			parsed, err := strconv.ParseFloat(strings.TrimSpace(str), 64)
			f, ok = parsed, err == nil
		}
		if !ok {
			return nil, errors.New("must be a number")
		}
		if t == discord.IntegerOptionType && f != float64(int64(f)) {
			return nil, errors.New("must be a whole number")
		}
		if min, ok := arg["min_value"].(float64); ok && f < min {
			return nil, fmt.Errorf("must be at least %v", min)
		}
		if max, ok := arg["max_value"].(float64); ok && f > max {
			return nil, fmt.Errorf("must be at most %v", max)
		}
		if err := checkChoice(arg, strconv.FormatFloat(f, 'f', -1, 64)); err != nil {
			return nil, err
		}
		return f, nil
	case discord.BooleanOptionType:
		switch v := v.(type) {
		case bool:
			return v, nil
		case string:
			if b, err := strconv.ParseBool(v); err == nil {
				return b, nil
			}
		}
		return nil, errors.New("must be true or false")
	case discord.UserOptionType, discord.ChannelOptionType, discord.RoleOptionType, discord.MentionableOptionType:
		id, ok := v.(string)
		if !ok {
			return nil, errors.New("must be an ID")
		}
		componentType := map[discord.CommandOptionType]int{
			discord.UserOptionType:        int(discord.UserSelectComponentType),
			discord.ChannelOptionType:     int(discord.ChannelSelectComponentType),
			discord.RoleOptionType:        int(discord.RoleSelectComponentType),
			discord.MentionableOptionType: int(discord.MentionableSelectComponentType),
		}[t]
		if err := s.resolveSelect(resolved, componentType, []string{id}); err != nil {
			return nil, err
		}
		return id, nil
	case discord.AttachmentOptionType:
		var att struct {
			URL         string `json:"url"`
			Filename    string `json:"filename"`
			ContentType string `json:"content_type"`
		}
		if url, ok := v.(string); ok {
			att.URL = url
		} else if err := json.Unmarshal(raw, &att); err != nil || att.URL == "" {
			return nil, errors.New(`must be the URL of a file or {"url": ..., "filename": ...}`)
		}
		if att.Filename == "" {
			att.Filename = att.URL[strings.LastIndex(att.URL, "/")+1:]
			att.Filename, _, _ = strings.Cut(att.Filename, "?")
		}
		id := flow.TestSnowflake().String()
		addResolved(resolved, "attachments", id, map[string]any{
			"id":           id,
			"filename":     att.Filename,
			"url":          att.URL,
			"proxy_url":    att.URL,
			"content_type": att.ContentType,
			"size":         0,
		})
		return id, nil
	}
	return v, nil
}

func checkChoice(arg map[string]any, value string) error {
	choices, _ := arg["choices"].([]any)
	if len(choices) == 0 {
		return nil
	}
	var values []string
	for _, c := range choices {
		choice, _ := c.(map[string]any)
		v := fmt.Sprint(choice["value"])
		if v == value {
			return nil
		}
		values = append(values, v)
	}
	return fmt.Errorf("must be one of the choices: %s", strings.Join(values, ", "))
}

// resolveSelect adds the resolved data Discord sends for the users, roles or
// channels picked in a select menu or command argument of the type to dst.
func (s *testSession) resolveSelect(dst map[string]any, componentType int, values []string) error {
	addUser := func(id discord.Snowflake) bool {
		user, err := s.user(discord.UserID(id))
		if err != nil {
			return false
		}
		addResolved(dst, "users", id.String(), user)
		if s.guildID.IsValid() {
			if member, err := s.member(s.guildID, discord.UserID(id)); err == nil {
				addResolved(dst, "members", id.String(), member)
			}
		}
		return true
	}
	addRole := func(id discord.Snowflake) bool {
		role, err := s.session.Role(s.guildID, discord.RoleID(id))
		if err != nil {
			return false
		}
		addResolved(dst, "roles", id.String(), role)
		return true
	}

	for _, value := range values {
		id, err := discord.ParseSnowflake(value)
		if err != nil {
			return fmt.Errorf("%q isn't an ID", value)
		}

		switch discord.ComponentType(componentType) {
		case discord.UserSelectComponentType:
			if !addUser(id) {
				return fmt.Errorf("user %s doesn't exist", value)
			}
		case discord.RoleSelectComponentType:
			if !addRole(id) {
				return fmt.Errorf("role %s doesn't exist in the server", value)
			}
		case discord.MentionableSelectComponentType:
			if !addRole(id) && !addUser(id) {
				return fmt.Errorf("%s is neither a user nor a role of the server", value)
			}
		case discord.ChannelSelectComponentType:
			channel, err := s.session.Channel(discord.ChannelID(id))
			if err != nil || channel.GuildID != s.guildID {
				return fmt.Errorf("channel %s isn't in the server", value)
			}
			addResolved(dst, "channels", value, channel)
		}
	}
	return nil
}

// addResolved adds v to the resolved data of the kind, like "users".
func addResolved(dst map[string]any, kind string, id string, v any) {
	m, _ := dst[kind].(map[string]any)
	if m == nil {
		m = map[string]any{}
		dst[kind] = m
	}
	m[id] = v
}

// listenerEvent builds the event of the listener's type, with the user and
// channel of the test, and test.Event merged into it.
func (s *testSession) listenerEvent(root *flow.CompiledFlowNode, test Test) (gateway.Event, error) {
	eventType := root.EventListenerType()
	now := time.Now().UTC()

	switch eventType {
	case flow.EventTypeScheduleCron:
		return &schedule.Event{Time: now}, nil
	case flow.EventTypeWebhook:
		w := test.Webhook
		if w == nil {
			w = &Webhook{}
		}
		if len(w.Body) > webhook.MaxBodySize {
			return nil, testErrorf("the body can be at most %d bytes", webhook.MaxBodySize)
		}
		headers := make(map[string]string, len(w.Headers))
		for k, v := range w.Headers {
			headers[strings.ToLower(k)] = v
		}
		query := w.Query
		if query == nil {
			query = map[string]string{}
		}
		return &webhook.Event{
			Headers:    headers,
			Query:      query,
			Body:       w.Body,
			ReceivedAt: now,
			Responder:  webhook.NewResponder(),
		}, nil
	}

	user, err := s.user(s.userID)
	if err != nil {
		return nil, testErrorf("user %s doesn't exist", s.userID)
	}
	var member *discord.Member
	if s.guildID.IsValid() {
		member, _ = s.member(s.guildID, s.userID)
	}

	msg := map[string]any{
		"id":         flow.TestSnowflake().String(),
		"channel_id": s.channelID.String(),
		"author":     user,
		"content":    "",
		"timestamp":  now.Format(time.RFC3339),
		"type":       discord.DefaultMessage,
	}
	if s.guildID.IsValid() {
		msg["guild_id"] = s.guildID.String()
		if member != nil {
			msg["member"] = member
		}
	}

	reaction := func() map[string]any {
		return map[string]any{
			"user_id":    s.userID.String(),
			"channel_id": s.channelID.String(),
			"message_id": flow.TestSnowflake().String(),
			"guild_id":   s.guildID.String(),
			"emoji":      map[string]any{"name": "👍"},
		}
	}

	switch model.EventListenerType(eventType) {
	case model.EventListenerTypeDiscordMessageCreate, model.EventListenerTypeDiscordDirectMessageCreate:
		return decodeEvent[gateway.MessageCreateEvent](msg, test.Event)
	case model.EventListenerTypeDiscordGuildBoost:
		msg["type"] = discord.NitroBoostMessage
		return decodeEvent[gateway.MessageCreateEvent](msg, test.Event)
	case model.EventListenerTypeDiscordMessageUpdate:
		msg["edited_timestamp"] = now.Format(time.RFC3339)
		return decodeEvent[gateway.MessageUpdateEvent](msg, test.Event)
	case model.EventListenerTypeDiscordMessageDelete:
		deleted, err := decodeEvent[discord.Message](msg, test.Event)
		if err != nil {
			return nil, err
		}
		return &guildevent.MessageDelete{
			MessageDeleteEvent: gateway.MessageDeleteEvent{ID: deleted.ID, ChannelID: deleted.ChannelID, GuildID: deleted.GuildID},
			Message:            deleted,
		}, nil
	case model.EventListenerTypeDiscordGuildMemberAdd:
		base := map[string]any{"guild_id": s.guildID.String(), "user": user, "joined_at": now.Format(time.RFC3339), "roles": []string{}}
		return decodeEvent[gateway.GuildMemberAddEvent](base, test.Event)
	case model.EventListenerTypeDiscordGuildMemberRemove:
		base := map[string]any{"guild_id": s.guildID.String(), "user": user}
		return decodeEvent[gateway.GuildMemberRemoveEvent](base, test.Event)
	case model.EventListenerTypeDiscordMessageReactionAdd:
		base := reaction()
		if member != nil {
			base["member"] = member
		}
		return decodeEvent[gateway.MessageReactionAddEvent](base, test.Event)
	case model.EventListenerTypeDiscordMessageReactionRemove:
		return decodeEvent[gateway.MessageReactionRemoveEvent](reaction(), test.Event)
	case model.EventListenerTypeDiscordGuildCreate:
		guild, err := s.session.Guild(s.guildID)
		if err != nil {
			return nil, testErrorf("the bot isn't in server %s", s.guildID)
		}
		e, err := decodeEvent[gateway.GuildCreateEvent](map[string]any{"joined_at": now.Format(time.RFC3339)}, test.Event)
		if err != nil {
			return nil, err
		}
		if !e.ID.IsValid() {
			e.Guild = *guild
		}
		return &state.GuildJoinEvent{GuildCreateEvent: e}, nil
	case model.EventListenerTypeDiscordGuildDelete:
		e, err := decodeEvent[gateway.GuildDeleteEvent](map[string]any{"id": s.guildID.String()}, test.Event)
		if err != nil {
			return nil, err
		}
		return &state.GuildLeaveEvent{GuildDeleteEvent: e}, nil
	case model.EventListenerTypeDiscordGuildAuditLogEntryCreate:
		base := map[string]any{
			"id":          flow.TestSnowflake().String(),
			"guild_id":    s.guildID.String(),
			"user_id":     s.userID.String(),
			"action_type": discord.GuildUpdate,
		}
		return decodeEvent[gateway.GuildAuditLogEntryCreateEvent](base, test.Event)
	}
	return nil, testErrorf("test runs don't support %s event listeners yet", eventType)
}

// decodeEvent decodes base with override merged into it as a T.
func decodeEvent[T any](base map[string]any, override json.RawMessage) (*T, error) {
	var e T
	if err := mergeJSON(base, override, &e); err != nil {
		return nil, testErrorf("invalid event: %v", err)
	}
	return &e, nil
}

// mergeJSON merges override into base, both JSON objects, and decodes the
// result into target. Objects are merged, other values replaced.
func mergeJSON(base map[string]any, override json.RawMessage, target any) error {
	b, err := json.Marshal(base)
	if err != nil {
		return err
	}

	if len(override) > 0 && string(override) != "null" {
		var merged map[string]any
		if err := json.Unmarshal(b, &merged); err != nil {
			return err
		}
		var o map[string]any
		if err := json.Unmarshal(override, &o); err != nil {
			return errors.New("must be a JSON object")
		}
		mergeMaps(merged, o)
		if b, err = json.Marshal(merged); err != nil {
			return err
		}
	}

	return json.Unmarshal(b, target)
}

func mergeMaps(dst map[string]any, src map[string]any) {
	for k, v := range src {
		if vm, ok := v.(map[string]any); ok {
			if dm, ok := dst[k].(map[string]any); ok {
				mergeMaps(dm, vm)
				continue
			}
		}
		dst[k] = v
	}
}
