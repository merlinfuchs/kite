package flow

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranscriptMarkdown(t *testing.T) {
	md := &transcriptMarkdown{
		users: []discord.GuildUser{{User: discord.User{ID: 2, Username: "jack"}}},
		roles: map[discord.RoleID]discord.Role{3: {ID: 3, Name: "Staff", Color: 0x00ff00}},
	}

	cases := map[string]string{
		"**bold** *italic* __underline__ ~~strike~~": "<strong>bold</strong> <em>italic</em> <u>underline</u> <s>strike</s>",
		"||secret||":               `<span class="spoiler">secret</span>`,
		"snake_case_name":          "snake_case_name",
		"`**not bold**`":           `<code class="inline-code">**not bold**</code>`,
		"<@2> <@!2>":               `<span class="mention">@jack</span> <span class="mention">@jack</span>`,
		"<@&3>":                    `<span class="mention" style="color: #00FF00">@Staff</span>`,
		"<#4>":                     `<span class="mention">#unknown-channel</span>`,
		"<:kite:123>":              `<img class="emoji" src="https://cdn.discordapp.com/emojis/123.webp" alt=":kite:" title=":kite:">`,
		"https://a.com/x_y":        `<a href="https://a.com/x_y" target="_blank" rel="noopener noreferrer">https://a.com/x_y</a>`,
		"[kite](https://kite.onl)": `<a href="https://kite.onl" target="_blank" rel="noopener noreferrer">kite</a>`,
	}
	for in, want := range cases {
		assert.Equal(t, want, string(md.render(in)), in)
	}
}

func TestTranscriptMarkdownEscapes(t *testing.T) {
	md := &transcriptMarkdown{}

	cases := []string{
		`<script>alert(1)</script>`,
		`<img src=x onerror=alert(1)>`,
		"```\n<script>alert(1)</script>\n```",
		"`<script>`",
		`[x](javascript:alert(1))`,
		"\x00" + "0\x01",
	}
	for _, in := range cases {
		res := string(md.render(in))
		assert.NotContains(t, res, "<script", in)
		assert.NotContains(t, res, "<img", in)
		assert.NotContains(t, res, `href="javascript`, in)
		assert.NotContains(t, res, "\x00", in)
	}
}

func TestRenderTranscript(t *testing.T) {
	start := time.Date(2026, 9, 29, 10, 0, 0, 0, time.UTC)
	author := discord.User{ID: 2, Username: "jack", DisplayName: "Jack"}

	res, err := renderTranscript(transcriptInput{
		GuildName: "Server <b>",
		Channel:   discord.Channel{ID: 4, Name: "ticket-1"},
		Messages: []discord.Message{
			{ID: 10, Author: author, Content: "first", Timestamp: discord.NewTimestamp(start)},
			{ID: 11, Author: author, Content: "second", Timestamp: discord.NewTimestamp(start.Add(time.Minute))},
			{
				ID: 12, Author: author, Timestamp: discord.NewTimestamp(start.Add(2 * time.Minute)),
				Embeds: []discord.Embed{{Title: "Embed", URL: "javascript:alert(1)"}},
			},
			{ID: 13, Author: author, Type: discord.ChannelPinnedMessage, Timestamp: discord.NewTimestamp(start.Add(3 * time.Minute))},
		},
		GeneratedAt: start,
	})
	require.NoError(t, err)

	html := string(res)
	assert.Contains(t, html, "<title>#ticket-1 - Server &lt;b&gt;</title>")
	assert.Contains(t, html, `<div class="message continued" id="m11">`)
	assert.Contains(t, html, "pinned a message to this channel.")
	assert.NotContains(t, html, "javascript:")
}

func TestTranscriptDataEvalMessageLimit(t *testing.T) {
	evalCtx := eval.NewContext(eval.Env{})

	limit, err := (&TranscriptData{}).evalMessageLimit(context.Background(), evalCtx)
	require.NoError(t, err)
	assert.Equal(t, transcriptMaxMessageLimit, limit)

	limit, err = (*TranscriptData)(nil).evalMessageLimit(context.Background(), evalCtx)
	require.NoError(t, err)
	assert.Equal(t, transcriptMaxMessageLimit, limit)

	limit, err = (&TranscriptData{MessageLimit: "250"}).evalMessageLimit(context.Background(), evalCtx)
	require.NoError(t, err)
	assert.Equal(t, 250, limit)

	for _, invalid := range []string{"0", "-5", "1001", "abc", "1.5"} {
		_, err := (&TranscriptData{MessageLimit: invalid}).evalMessageLimit(context.Background(), evalCtx)
		assert.Error(t, err, invalid)
	}
}

func TestTranscriptFileName(t *testing.T) {
	assert.Equal(t, "transcript-ticket-0001.html", transcriptFileName(&discord.Channel{Name: "ticket-0001"}))
	assert.Equal(t, "transcript-general-chat.html", transcriptFileName(&discord.Channel{Name: "General Chat!"}))
	assert.Equal(t, "transcript-42.html", transcriptFileName(&discord.Channel{ID: 42, Name: "🎫"}))
}

type transcriptTestDiscordProvider struct {
	provider.MockDiscordProvider

	messages []discord.Message
	limit    uint
}

func (p *transcriptTestDiscordProvider) Channel(ctx context.Context, channelID discord.ChannelID) (*discord.Channel, error) {
	return &discord.Channel{ID: channelID, Name: "ticket"}, nil
}

func (p *transcriptTestDiscordProvider) ChannelMessages(ctx context.Context, channelID discord.ChannelID, limit uint) ([]discord.Message, error) {
	p.limit = limit
	return p.messages, nil
}

func TestCreateTranscript(t *testing.T) {
	start := time.Date(2026, 9, 29, 10, 0, 0, 0, time.UTC)
	p := &transcriptTestDiscordProvider{
		// Newest first, like Discord returns them.
		messages: []discord.Message{
			{ID: 2, Content: "newer", Timestamp: discord.NewTimestamp(start.Add(time.Minute))},
			{ID: 1, Content: "older", Timestamp: discord.NewTimestamp(start)},
		},
	}

	file, err := createTranscript(context.Background(), p, 4, 50)
	require.NoError(t, err)
	assert.Equal(t, uint(50), p.limit)
	assert.Equal(t, "transcript-ticket.html", file.Name)

	html := string(file.Data)
	assert.Less(t, strings.Index(html, "older"), strings.Index(html, "newer"))
}

func TestEvalTemplateFiles(t *testing.T) {
	file := thing.FileValue{Name: "transcript-ticket.html", Data: []byte("<html></html>")}
	evalCtx := eval.NewContext(eval.Env{
		"result": func(id string) (any, error) {
			if id == "transcript" {
				return eval.NewThingEnv(thing.NewFile(file)), nil
			}
			return eval.NewThingEnv(thing.NewString("text")), nil
		},
	})

	template, files := evalTemplateFiles(context.Background(), "Closed {{result('transcript')}} by {{result('other')}}", evalCtx)
	assert.Equal(t, "Closed  by {{result('other')}}", template)
	require.Len(t, files, 1)
	assert.Equal(t, file, files[0])

	template, files = evalTemplateFiles(context.Background(), "{{result('transcript').name}}", evalCtx)
	assert.Equal(t, "{{result('transcript').name}}", template)
	assert.Empty(t, files)

	res, err := eval.EvalTemplateToString(context.Background(), "{{result('transcript').name}}", evalCtx)
	require.NoError(t, err)
	assert.Equal(t, "transcript-ticket.html", res)
}

func TestToSendFiles(t *testing.T) {
	files, err := toSendFiles(nil)
	require.NoError(t, err)
	assert.Nil(t, files)

	_, err = toSendFiles([]thing.FileValue{{Name: "empty.html"}})
	assert.Error(t, err)

	tooMany := make([]thing.FileValue, maxMessageFiles+1)
	for i := range tooMany {
		tooMany[i] = thing.FileValue{Name: "a.html", Data: []byte("a")}
	}
	_, err = toSendFiles(tooMany)
	assert.Error(t, err)
}
