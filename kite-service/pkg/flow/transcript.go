package flow

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

// Limits of the channel transcript block. Fetching 100 messages is one
// request, so the maximum keeps the block well within the execution timeout.
const (
	transcriptMaxMessageLimit = 1000
	// Discord's upload limit for bots without boosts is 10 MB, this leaves
	// room for the rest of the request.
	transcriptMaxFileSize = 8 << 20
	// Channel mentions are resolved one by one, so only this many are.
	transcriptMaxChannelLookups = 25
)

var transcriptChannelMentionRe = regexp.MustCompile(`<#(\d+)>`)

// evalMessageLimit evaluates the message limit, which is the maximum when
// it's empty.
func (d *TranscriptData) evalMessageLimit(ctx context.Context, evalCtx eval.Context) (int, error) {
	if d == nil || d.MessageLimit == "" {
		return transcriptMaxMessageLimit, nil
	}

	raw, err := eval.EvalTemplate(ctx, d.MessageLimit, evalCtx)
	if err != nil {
		return 0, err
	}

	limit, err := strconv.Atoi(strings.TrimSpace(raw.String()))
	if err != nil {
		return 0, fmt.Errorf("message limit %q is not a whole number", raw.String())
	}
	if limit < 1 || limit > transcriptMaxMessageLimit {
		return 0, fmt.Errorf("message limit must be between 1 and %d, got %d", transcriptMaxMessageLimit, limit)
	}

	return limit, nil
}

// createTranscript fetches the most recent messages of a channel and renders
// them as an HTML file.
func createTranscript(
	ctx context.Context,
	discordProvider provider.DiscordProvider,
	channelID discord.ChannelID,
	limit int,
) (thing.FileValue, error) {
	channel, err := discordProvider.Channel(ctx, channelID)
	if err != nil {
		return thing.FileValue{}, err
	}

	messages, err := discordProvider.ChannelMessages(ctx, channelID, uint(limit))
	if err != nil {
		return thing.FileValue{}, err
	}
	// Discord returns the newest messages first.
	slices.Reverse(messages)

	in := transcriptInput{
		Channel:     *channel,
		Messages:    messages,
		Roles:       map[discord.RoleID]discord.Role{},
		Channels:    map[discord.ChannelID]string{channel.ID: channel.Name},
		GeneratedAt: time.Now(),
	}

	// Names are only used to make mentions readable, so the transcript is
	// still created if they can't be resolved.
	if channel.GuildID.IsValid() {
		if guild, err := discordProvider.Guild(ctx, channel.GuildID); err == nil && guild != nil {
			in.GuildName = guild.Name
		}
		if roles, err := discordProvider.GuildRoles(ctx, channel.GuildID); err == nil {
			for _, role := range roles {
				in.Roles[role.ID] = role
			}
		}
	}

	lookups := 0
	for _, msg := range messages {
		for _, match := range transcriptChannelMentionRe.FindAllStringSubmatch(msg.Content, -1) {
			id, err := discord.ParseSnowflake(match[1])
			if err != nil {
				continue
			}
			if _, ok := in.Channels[discord.ChannelID(id)]; ok || lookups >= transcriptMaxChannelLookups {
				continue
			}

			lookups++
			if mentioned, err := discordProvider.Channel(ctx, discord.ChannelID(id)); err == nil && mentioned != nil {
				in.Channels[mentioned.ID] = mentioned.Name
			}
		}
	}

	file, err := renderTranscript(in)
	if err != nil {
		return thing.FileValue{}, err
	}
	if len(file) > transcriptMaxFileSize {
		return thing.FileValue{}, fmt.Errorf(
			"transcript is %.1f MB, the maximum is %d MB, try a lower message limit",
			float64(len(file))/(1<<20), transcriptMaxFileSize>>20,
		)
	}

	return thing.FileValue{
		Name:        transcriptFileName(channel),
		ContentType: "text/html; charset=utf-8",
		Data:        file,
	}, nil
}

var transcriptFileNameRe = regexp.MustCompile(`[^a-z0-9_-]+`)

func transcriptFileName(channel *discord.Channel) string {
	name := transcriptFileNameRe.ReplaceAllString(strings.ToLower(channel.Name), "-")
	name = strings.Trim(name, "-")
	if name == "" {
		name = channel.ID.String()
	}
	return "transcript-" + name + ".html"
}
