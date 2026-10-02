package flow

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
)

// transcriptGroupWindow is how long after a message another message of the
// same author is shown without repeating the author, like Discord does.
const transcriptGroupWindow = 7 * time.Minute

// transcriptInput is everything a channel transcript is rendered from.
type transcriptInput struct {
	GuildName string
	Channel   discord.Channel
	// Messages are ordered oldest first.
	Messages    []discord.Message
	Roles       map[discord.RoleID]discord.Role
	Channels    map[discord.ChannelID]string
	GeneratedAt time.Time
}

type transcriptView struct {
	Title        string
	GuildName    string
	ChannelName  string
	ChannelTopic string
	MessageCount int
	GeneratedAt  transcriptTime
	Messages     []transcriptMessageView
}

type transcriptTime struct {
	ISO  string
	Text string
}

type transcriptMessageView struct {
	ID          string
	Continued   bool
	System      string
	AuthorName  string
	AuthorTag   string
	AvatarURL   string
	Bot         bool
	Time        transcriptTime
	Edited      bool
	Reply       *transcriptReplyView
	Command     string
	Content     template.HTML
	Components  template.HTML
	Embeds      []transcriptEmbedView
	Attachments []transcriptAttachmentView
	Stickers    []string
	Reactions   []transcriptReactionView
}

type transcriptReplyView struct {
	AuthorName string
	Content    string
}

type transcriptEmbedView struct {
	Color        string
	AuthorName   string
	AuthorIcon   string
	Title        string
	URL          string
	Description  template.HTML
	Fields       []transcriptEmbedFieldView
	ImageURL     string
	ThumbnailURL string
	Footer       string
	FooterIcon   string
}

type transcriptEmbedFieldView struct {
	Name   template.HTML
	Value  template.HTML
	Inline bool
}

type transcriptAttachmentView struct {
	Name    string
	URL     string
	Size    string
	IsImage bool
}

type transcriptReactionView struct {
	EmojiURL  string
	EmojiText string
	Count     int
}

// renderTranscript renders the messages of a channel as a standalone HTML
// page that looks like the Discord client.
func renderTranscript(in transcriptInput) ([]byte, error) {
	md := &transcriptMarkdown{
		roles:    in.Roles,
		channels: in.Channels,
	}

	view := transcriptView{
		GuildName:    in.GuildName,
		ChannelName:  in.Channel.Name,
		ChannelTopic: in.Channel.Topic,
		MessageCount: len(in.Messages),
		GeneratedAt:  newTranscriptTime(in.GeneratedAt),
	}
	if view.ChannelName == "" {
		view.ChannelName = in.Channel.ID.String()
	}
	view.Title = "#" + view.ChannelName
	if view.GuildName != "" {
		view.Title += " - " + view.GuildName
	}

	var prev *discord.Message
	for i := range in.Messages {
		msg := &in.Messages[i]
		md.users = msg.Mentions

		v := transcriptMessageView{
			ID:         msg.ID.String(),
			AuthorName: msg.Author.DisplayOrUsername(),
			AuthorTag:  msg.Author.Tag(),
			AvatarURL:  msg.Author.AvatarURLWithType(discord.PNGImage),
			Bot:        msg.Author.Bot,
			Time:       newTranscriptTime(msg.Timestamp.Time()),
			Edited:     msg.EditedTimestamp.IsValid(),
			System:     transcriptSystemText(msg),
		}

		if msg.ReferencedMessage != nil {
			v.Reply = &transcriptReplyView{
				AuthorName: msg.ReferencedMessage.Author.DisplayOrUsername(),
				Content:    transcriptSnippet(msg.ReferencedMessage.Content, 100),
			}
		}
		if msg.Interaction != nil && msg.Interaction.Name != "" {
			v.Command = msg.Interaction.User.DisplayOrUsername() + " used /" + msg.Interaction.Name
		}

		if v.System == "" {
			v.Content = md.render(msg.Content)
			v.Components = md.renderComponents(msg.Components)
		}

		for _, e := range msg.Embeds {
			v.Embeds = append(v.Embeds, md.embedView(e))
		}
		for _, a := range msg.Attachments {
			v.Attachments = append(v.Attachments, transcriptAttachmentView{
				Name:    a.Filename,
				URL:     string(a.URL),
				Size:    transcriptFileSize(a.Size),
				IsImage: strings.HasPrefix(a.ContentType, "image/"),
			})
		}
		for _, s := range msg.Stickers {
			v.Stickers = append(v.Stickers, s.Name)
		}
		for _, r := range msg.Reactions {
			rv := transcriptReactionView{Count: r.Count, EmojiText: r.Emoji.Name}
			if r.Emoji.ID.IsValid() {
				rv.EmojiURL = transcriptEmojiURL(r.Emoji.ID.String(), r.Emoji.Animated)
			}
			v.Reactions = append(v.Reactions, rv)
		}

		// Consecutive messages of the same author are grouped, unless a
		// message starts something new like a reply or a command.
		v.Continued = prev != nil &&
			v.System == "" &&
			v.Reply == nil &&
			v.Command == "" &&
			transcriptSystemText(prev) == "" &&
			prev.Author.ID == msg.Author.ID &&
			msg.Timestamp.Time().Sub(prev.Timestamp.Time()) < transcriptGroupWindow

		view.Messages = append(view.Messages, v)
		prev = msg
	}

	var buf bytes.Buffer
	if err := transcriptTemplate.Execute(&buf, view); err != nil {
		return nil, fmt.Errorf("failed to render transcript: %w", err)
	}

	return buf.Bytes(), nil
}

func newTranscriptTime(t time.Time) transcriptTime {
	t = t.UTC()
	return transcriptTime{
		ISO:  t.Format(time.RFC3339),
		Text: t.Format("2006-01-02 15:04 UTC"),
	}
}

// transcriptSystemText returns the text shown instead of the content for
// messages Discord generates itself, or "" for normal messages.
func transcriptSystemText(msg *discord.Message) string {
	switch msg.Type {
	case discord.ChannelPinnedMessage:
		return "pinned a message to this channel."
	case discord.GuildMemberJoinMessage:
		return "joined the server."
	case discord.NitroBoostMessage, discord.NitroTier1Message, discord.NitroTier2Message, discord.NitroTier3Message:
		return "boosted the server."
	case discord.ThreadCreatedMessage:
		return "started a thread: " + msg.Content
	case discord.ChannelNameChangeMessage:
		return "changed the channel name: " + msg.Content
	case discord.RecipientAddMessage:
		return "added someone to the thread."
	case discord.RecipientRemoveMessage:
		return "removed someone from the thread."
	}
	return ""
}

func transcriptSnippet(s string, max int) string {
	s = strings.Join(strings.Fields(s), " ")
	runes := []rune(s)
	if len(runes) > max {
		return string(runes[:max]) + "..."
	}
	return s
}

func transcriptFileSize(size uint64) string {
	switch {
	case size >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(size)/(1<<20))
	case size >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(size)/(1<<10))
	}
	return fmt.Sprintf("%d bytes", size)
}

func transcriptEmojiURL(id string, animated bool) string {
	ext := "webp"
	if animated {
		ext = "gif"
	}
	return "https://cdn.discordapp.com/emojis/" + id + "." + ext
}

// transcriptSafeURL returns the URL if it's a http(s) link, so a message
// can't smuggle javascript: or data: links into the transcript.
func transcriptSafeURL(u string) string {
	lower := strings.ToLower(strings.TrimSpace(u))
	if strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "http://") {
		return strings.TrimSpace(u)
	}
	return ""
}

func (m *transcriptMarkdown) embedView(e discord.Embed) transcriptEmbedView {
	v := transcriptEmbedView{
		Title:       e.Title,
		URL:         transcriptSafeURL(string(e.URL)),
		Description: m.render(e.Description),
	}
	if e.Color != 0 {
		v.Color = e.Color.String()
	}
	if e.Author != nil {
		v.AuthorName = e.Author.Name
		v.AuthorIcon = transcriptSafeURL(string(e.Author.Icon))
	}
	for _, f := range e.Fields {
		v.Fields = append(v.Fields, transcriptEmbedFieldView{
			Name:   m.render(f.Name),
			Value:  m.render(f.Value),
			Inline: f.Inline,
		})
	}
	if e.Image != nil {
		v.ImageURL = transcriptSafeURL(string(e.Image.URL))
	}
	if e.Thumbnail != nil {
		v.ThumbnailURL = transcriptSafeURL(string(e.Thumbnail.URL))
	}
	if e.Footer != nil {
		v.Footer = e.Footer.Text
		v.FooterIcon = transcriptSafeURL(string(e.Footer.Icon))
	}
	return v
}

// renderComponents renders the text of Components V2 messages, which keep
// their content in text display components instead of the content field.
func (m *transcriptMarkdown) renderComponents(components discord.TopLevelComponents) template.HTML {
	var b strings.Builder
	for _, c := range components {
		m.writeComponent(&b, c)
	}
	return template.HTML(b.String())
}

func (m *transcriptMarkdown) writeComponent(b *strings.Builder, c discord.Component) {
	switch c := c.(type) {
	case *discord.TextDisplayComponent:
		b.WriteString(`<div class="cv2-text">`)
		b.WriteString(string(m.render(c.Content)))
		b.WriteString(`</div>`)
	case *discord.SectionComponent:
		for _, child := range c.Components {
			m.writeComponent(b, child)
		}
	case *discord.ContainerComponent:
		style := ""
		if c.AccentColor != nil {
			style = ` style="border-left-color: ` + c.AccentColor.String() + `"`
		}
		b.WriteString(`<div class="cv2-container"` + style + `>`)
		for _, child := range c.Components {
			m.writeComponent(b, child)
		}
		b.WriteString(`</div>`)
	case *discord.SeparatorComponent:
		b.WriteString(`<hr class="cv2-separator">`)
	case *discord.ActionRowComponent:
		b.WriteString(`<div class="buttons">`)
		for _, child := range *c {
			if btn, ok := child.(*discord.ButtonComponent); ok && btn.Label != "" {
				b.WriteString(`<span class="button">` + html.EscapeString(btn.Label) + `</span>`)
			}
		}
		b.WriteString(`</div>`)
	}
}

// transcriptMarkdown renders the subset of Discord markdown that matters
// for reading a transcript. Text is always escaped before any HTML is added,
// and generated HTML is swapped out for placeholders so later rules can't
// touch it.
type transcriptMarkdown struct {
	users    []discord.GuildUser
	roles    map[discord.RoleID]discord.Role
	channels map[discord.ChannelID]string

	protected []string
}

var (
	mdCodeBlockRe  = regexp.MustCompile("```(?:([a-zA-Z0-9_+-]+)\n)?([\\s\\S]*?)```")
	mdInlineCodeRe = regexp.MustCompile("`([^`\n]+)`")
	mdPlaceholder  = regexp.MustCompile("\x00(\\d+)\x01")

	mdEmojiRe       = regexp.MustCompile(`&lt;(a?):(\w+):(\d+)&gt;`)
	mdUserRe        = regexp.MustCompile(`&lt;@!?(\d+)&gt;`)
	mdRoleRe        = regexp.MustCompile(`&lt;@&amp;(\d+)&gt;`)
	mdChannelRe     = regexp.MustCompile(`&lt;#(\d+)&gt;`)
	mdCommandRe     = regexp.MustCompile(`&lt;/([\w -]+):\d+&gt;`)
	mdTimestampRe   = regexp.MustCompile(`&lt;t:(-?\d+)(?::[tTdDfFR])?&gt;`)
	mdEveryoneRe    = regexp.MustCompile(`@(everyone|here)\b`)
	mdMaskedLinkRe  = regexp.MustCompile(`\[([^\[\]\n]+)\]\((https?://(?:&amp;|[^\s&<>"()])+)\)`)
	mdLinkRe        = regexp.MustCompile("https?://(?:&amp;|[^\\s&<>\"\x00\x01])+")
	mdHeadingRe     = regexp.MustCompile(`(?m)^(#{1,3}) (.+)$`)
	mdSubtextRe     = regexp.MustCompile(`(?m)^-# (.+)$`)
	mdQuoteRe       = regexp.MustCompile(`(?m)^&gt; ?(.*)$`)
	mdBoldRe        = regexp.MustCompile(`\*\*(.+?)\*\*`)
	mdUnderlineRe   = regexp.MustCompile(`__(.+?)__`)
	mdItalicStarRe  = regexp.MustCompile(`\*([^*\s](?:[^*]*[^*\s])?)\*`)
	mdItalicUnderRe = regexp.MustCompile(`\b_([^_\s](?:[^_]*[^_\s])?)_\b`)
	mdStrikeRe      = regexp.MustCompile(`~~(.+?)~~`)
	mdSpoilerRe     = regexp.MustCompile(`\|\|(.+?)\|\|`)
)

func (m *transcriptMarkdown) protect(s string) string {
	m.protected = append(m.protected, s)
	return "\x00" + strconv.Itoa(len(m.protected)-1) + "\x01"
}

func (m *transcriptMarkdown) render(s string) template.HTML {
	if s == "" {
		return ""
	}
	m.protected = m.protected[:0]

	// Placeholders use these characters, so they must not come from the text.
	s = strings.NewReplacer("\x00", "", "\x01", "").Replace(s)

	s = mdCodeBlockRe.ReplaceAllStringFunc(s, func(match string) string {
		parts := mdCodeBlockRe.FindStringSubmatch(match)
		code := strings.TrimPrefix(parts[2], "\n")
		return m.protect(`<pre class="code-block"><code>` + html.EscapeString(code) + `</code></pre>`)
	})
	s = mdInlineCodeRe.ReplaceAllStringFunc(s, func(match string) string {
		code := mdInlineCodeRe.FindStringSubmatch(match)[1]
		return m.protect(`<code class="inline-code">` + html.EscapeString(code) + `</code>`)
	})

	s = html.EscapeString(s)

	s = mdEmojiRe.ReplaceAllStringFunc(s, func(match string) string {
		parts := mdEmojiRe.FindStringSubmatch(match)
		return m.protect(`<img class="emoji" src="` + transcriptEmojiURL(parts[3], parts[1] == "a") + `" alt=":` + parts[2] + `:" title=":` + parts[2] + `:">`)
	})
	s = mdRoleRe.ReplaceAllStringFunc(s, func(match string) string {
		id, _ := discord.ParseSnowflake(mdRoleRe.FindStringSubmatch(match)[1])
		name := "unknown-role"
		style := ""
		if role, ok := m.roles[discord.RoleID(id)]; ok {
			name = role.Name
			if role.Color != 0 {
				style = ` style="color: ` + role.Color.String() + `"`
			}
		}
		return m.protect(`<span class="mention"` + style + `>@` + html.EscapeString(name) + `</span>`)
	})
	s = mdUserRe.ReplaceAllStringFunc(s, func(match string) string {
		id, _ := discord.ParseSnowflake(mdUserRe.FindStringSubmatch(match)[1])
		name := "unknown-user"
		for _, u := range m.users {
			if u.ID == discord.UserID(id) {
				name = u.DisplayOrUsername()
				if u.Member != nil && u.Member.Nick != "" {
					name = u.Member.Nick
				}
				break
			}
		}
		return m.protect(`<span class="mention">@` + html.EscapeString(name) + `</span>`)
	})
	s = mdChannelRe.ReplaceAllStringFunc(s, func(match string) string {
		id, _ := discord.ParseSnowflake(mdChannelRe.FindStringSubmatch(match)[1])
		name, ok := m.channels[discord.ChannelID(id)]
		if !ok {
			name = "unknown-channel"
		}
		return m.protect(`<span class="mention">#` + html.EscapeString(name) + `</span>`)
	})
	s = mdCommandRe.ReplaceAllStringFunc(s, func(match string) string {
		// The name is already escaped.
		return m.protect(`<span class="mention">/` + mdCommandRe.FindStringSubmatch(match)[1] + `</span>`)
	})
	s = mdTimestampRe.ReplaceAllStringFunc(s, func(match string) string {
		unix, err := strconv.ParseInt(mdTimestampRe.FindStringSubmatch(match)[1], 10, 64)
		if err != nil {
			return match
		}
		t := newTranscriptTime(time.Unix(unix, 0))
		return m.protect(`<time class="timestamp" datetime="` + t.ISO + `">` + t.Text + `</time>`)
	})
	s = mdEveryoneRe.ReplaceAllStringFunc(s, func(match string) string {
		return m.protect(`<span class="mention">` + match + `</span>`)
	})

	// The URLs are already escaped, which also makes them safe to use as
	// attribute values.
	s = mdMaskedLinkRe.ReplaceAllStringFunc(s, func(match string) string {
		parts := mdMaskedLinkRe.FindStringSubmatch(match)
		return m.protect(`<a href="`+parts[2]+`" target="_blank" rel="noopener noreferrer">`) +
			parts[1] + m.protect(`</a>`)
	})
	s = mdLinkRe.ReplaceAllStringFunc(s, func(match string) string {
		return m.protect(`<a href="` + match + `" target="_blank" rel="noopener noreferrer">` + match + `</a>`)
	})

	s = mdHeadingRe.ReplaceAllStringFunc(s, func(match string) string {
		parts := mdHeadingRe.FindStringSubmatch(match)
		level := strconv.Itoa(len(parts[1]))
		return m.protect(`<span class="h`+level+`">`) + parts[2] + m.protect(`</span>`)
	})
	s = mdSubtextRe.ReplaceAllString(s, m.protect(`<span class="subtext">`)+"$1"+m.protect(`</span>`))
	s = mdQuoteRe.ReplaceAllString(s, m.protect(`<span class="quote">`)+"$1"+m.protect(`</span>`))

	s = mdBoldRe.ReplaceAllString(s, "<strong>$1</strong>")
	s = mdUnderlineRe.ReplaceAllString(s, "<u>$1</u>")
	s = mdItalicStarRe.ReplaceAllString(s, "<em>$1</em>")
	s = mdItalicUnderRe.ReplaceAllString(s, "<em>$1</em>")
	s = mdStrikeRe.ReplaceAllString(s, "<s>$1</s>")
	s = mdSpoilerRe.ReplaceAllString(s, `<span class="spoiler">$1</span>`)

	// Protected HTML can contain placeholders itself, e.g. an emoji in the
	// text of a masked link, so this runs until none are left.
	for range 4 {
		if !mdPlaceholder.MatchString(s) {
			break
		}
		s = mdPlaceholder.ReplaceAllStringFunc(s, func(match string) string {
			i, _ := strconv.Atoi(mdPlaceholder.FindStringSubmatch(match)[1])
			if i < len(m.protected) {
				return m.protected[i]
			}
			return ""
		})
	}

	return template.HTML(s)
}

var transcriptTemplate = template.Must(template.New("transcript").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}}</title>
<style>
:root {
  --bg: #313338; --bg-alt: #2b2d31; --bg-dark: #1e1f22; --text: #dbdee1;
  --muted: #949ba4; --link: #00a8fc; --mention-bg: rgba(88, 101, 242, 0.3);
  --mention: #c9cdfb; --embed-border: #1e1f22; --hover: #2e3035;
}
* { box-sizing: border-box; }
body { margin: 0; background: var(--bg); color: var(--text); font: 16px/1.375 "gg sans", "Noto Sans", "Helvetica Neue", Helvetica, Arial, sans-serif; }
header { position: sticky; top: 0; z-index: 1; background: var(--bg); border-bottom: 1px solid var(--bg-dark); padding: 12px 16px; display: flex; flex-wrap: wrap; align-items: baseline; gap: 4px 12px; }
header .channel { font-weight: 600; font-size: 16px; color: #f2f3f5; }
header .channel::before { content: "#"; color: var(--muted); margin-right: 4px; }
header .guild, header .meta { color: var(--muted); font-size: 14px; }
header .topic { color: var(--muted); font-size: 14px; flex-basis: 100%; }
main { padding: 16px 0 32px; }
.message { position: relative; padding: 2px 16px 2px 72px; min-height: 44px; margin-top: 17px; }
.message.continued { margin-top: 0; min-height: 0; }
.message:hover { background: var(--hover); }
.avatar { position: absolute; left: 16px; top: 4px; width: 40px; height: 40px; border-radius: 50%; }
.reply, .command { font-size: 14px; color: var(--muted); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.reply strong, .command strong { color: #f2f3f5; font-weight: 500; }
.header { display: flex; align-items: baseline; gap: 8px; }
.author { font-weight: 500; color: #f2f3f5; }
.bot { background: #5865f2; color: #fff; font-size: 10px; font-weight: 600; padding: 1px 4px; border-radius: 3px; text-transform: uppercase; }
.time { font-size: 12px; color: var(--muted); }
.side-time { display: none; position: absolute; left: 0; width: 64px; text-align: right; font-size: 11px; color: var(--muted); top: 6px; }
.message.continued:hover .side-time { display: block; }
.content { white-space: pre-wrap; word-wrap: break-word; }
.system { color: var(--muted); font-style: italic; }
.edited { font-size: 10px; color: var(--muted); margin-left: 4px; }
a { color: var(--link); text-decoration: none; }
a:hover { text-decoration: underline; }
.mention { background: var(--mention-bg); color: var(--mention); border-radius: 3px; padding: 0 2px; font-weight: 500; }
.emoji { width: 22px; height: 22px; vertical-align: bottom; object-fit: contain; }
.inline-code { background: var(--bg-dark); border-radius: 4px; padding: 0 3px; font-family: Consolas, "Courier New", monospace; font-size: 85%; }
.code-block { background: var(--bg-alt); border: 1px solid var(--bg-dark); border-radius: 4px; padding: 8px; margin: 4px 0; white-space: pre-wrap; font-family: Consolas, "Courier New", monospace; font-size: 14px; }
.spoiler { background: var(--bg-dark); color: transparent; border-radius: 3px; cursor: pointer; }
.spoiler:hover, .spoiler.revealed { color: inherit; background: rgba(255, 255, 255, 0.1); }
.quote { display: inline-block; border-left: 4px solid #4e5058; padding-left: 12px; }
.h1, .h2, .h3 { display: inline-block; font-weight: 700; color: #f2f3f5; }
.h1 { font-size: 24px; } .h2 { font-size: 20px; } .h3 { font-size: 16px; }
.subtext { font-size: 12px; color: var(--muted); }
.embed { display: flex; max-width: 520px; margin-top: 4px; background: var(--bg-alt); border-left: 4px solid var(--embed-border); border-radius: 4px; padding: 8px 16px 16px 12px; gap: 16px; }
.embed-body { flex: 1; min-width: 0; }
.embed-author { display: flex; align-items: center; gap: 8px; font-size: 14px; font-weight: 600; margin-top: 8px; }
.embed-author img { width: 24px; height: 24px; border-radius: 50%; }
.embed-title { font-weight: 600; margin-top: 8px; color: #f2f3f5; }
.embed-description { font-size: 14px; margin-top: 8px; white-space: pre-wrap; }
.embed-fields { display: flex; flex-wrap: wrap; gap: 8px 16px; margin-top: 8px; }
.embed-field { flex-basis: 100%; font-size: 14px; }
.embed-field.inline { flex: 1 1 150px; }
.embed-field-name { font-weight: 600; color: #f2f3f5; }
.embed-field-value { white-space: pre-wrap; }
.embed-image { max-width: 100%; border-radius: 4px; margin-top: 16px; }
.embed-thumbnail { width: 80px; height: 80px; object-fit: contain; border-radius: 4px; margin-top: 8px; }
.embed-footer { display: flex; align-items: center; gap: 8px; font-size: 12px; color: var(--muted); margin-top: 8px; }
.embed-footer img { width: 20px; height: 20px; border-radius: 50%; }
.cv2-container { max-width: 520px; margin-top: 4px; background: var(--bg-alt); border-left: 4px solid var(--embed-border); border-radius: 8px; padding: 12px 16px; }
.cv2-text { white-space: pre-wrap; margin: 2px 0; }
.cv2-separator { border: none; border-top: 1px solid #3f4147; margin: 8px 0; }
.buttons { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 4px; }
.button { background: #4e5058; color: #fff; border-radius: 3px; padding: 2px 16px; font-size: 14px; line-height: 28px; }
.attachment-image { display: block; max-width: 400px; max-height: 300px; border-radius: 8px; margin-top: 4px; }
.attachment { display: inline-flex; flex-direction: column; margin-top: 4px; background: var(--bg-alt); border: 1px solid var(--bg-dark); border-radius: 8px; padding: 10px 12px; max-width: 432px; }
.attachment-size { font-size: 12px; color: var(--muted); }
.sticker { display: inline-block; margin-top: 4px; color: var(--muted); font-size: 14px; }
.reactions { display: flex; flex-wrap: wrap; gap: 4px; margin-top: 4px; }
.reaction { display: inline-flex; align-items: center; gap: 6px; background: var(--bg-alt); border-radius: 8px; padding: 2px 6px; font-size: 14px; }
.reaction img { width: 16px; height: 16px; }
footer { color: var(--muted); font-size: 12px; text-align: center; padding: 16px; border-top: 1px solid var(--bg-dark); }
</style>
</head>
<body>
<header>
  <span class="channel">{{.ChannelName}}</span>
  {{if .GuildName}}<span class="guild">{{.GuildName}}</span>{{end}}
  <span class="meta">{{.MessageCount}} message{{if ne .MessageCount 1}}s{{end}}</span>
  {{if .ChannelTopic}}<span class="topic">{{.ChannelTopic}}</span>{{end}}
</header>
<main>
{{range .Messages}}
<div class="message{{if .Continued}} continued{{end}}" id="m{{.ID}}">
  {{if .Reply}}<div class="reply">&#8627; <strong>{{.Reply.AuthorName}}</strong> {{.Reply.Content}}</div>{{end}}
  {{if .Command}}<div class="command">&#8627; {{.Command}}</div>{{end}}
  {{if .Continued}}
  <time class="side-time" datetime="{{.Time.ISO}}" data-format="time">{{.Time.Text}}</time>
  {{else}}
  <img class="avatar" src="{{.AvatarURL}}" alt="" loading="lazy">
  <div class="header">
    <span class="author" title="{{.AuthorTag}}">{{.AuthorName}}</span>
    {{if .Bot}}<span class="bot">App</span>{{end}}
    <time class="time" datetime="{{.Time.ISO}}">{{.Time.Text}}</time>
  </div>
  {{end}}
  {{if .System}}<div class="content system">{{.System}}</div>{{end}}
  {{if .Content}}<div class="content">{{.Content}}{{if .Edited}}<span class="edited">(edited)</span>{{end}}</div>{{end}}
  {{.Components}}
  {{range .Embeds}}
  <div class="embed"{{if .Color}} style="border-left-color: {{.Color}}"{{end}}>
    <div class="embed-body">
      {{if .AuthorName}}<div class="embed-author">{{if .AuthorIcon}}<img src="{{.AuthorIcon}}" alt="">{{end}}{{.AuthorName}}</div>{{end}}
      {{if .Title}}<div class="embed-title">{{if .URL}}<a href="{{.URL}}" target="_blank" rel="noopener noreferrer">{{.Title}}</a>{{else}}{{.Title}}{{end}}</div>{{end}}
      {{if .Description}}<div class="embed-description">{{.Description}}</div>{{end}}
      {{if .Fields}}<div class="embed-fields">{{range .Fields}}<div class="embed-field{{if .Inline}} inline{{end}}"><div class="embed-field-name">{{.Name}}</div><div class="embed-field-value">{{.Value}}</div></div>{{end}}</div>{{end}}
      {{if .ImageURL}}<img class="embed-image" src="{{.ImageURL}}" alt="" loading="lazy">{{end}}
      {{if .Footer}}<div class="embed-footer">{{if .FooterIcon}}<img src="{{.FooterIcon}}" alt="">{{end}}{{.Footer}}</div>{{end}}
    </div>
    {{if .ThumbnailURL}}<img class="embed-thumbnail" src="{{.ThumbnailURL}}" alt="" loading="lazy">{{end}}
  </div>
  {{end}}
  {{range .Attachments}}
  {{if .IsImage}}<a href="{{.URL}}" target="_blank" rel="noopener noreferrer"><img class="attachment-image" src="{{.URL}}" alt="{{.Name}}" loading="lazy"></a>
  {{else}}<div class="attachment"><a href="{{.URL}}" target="_blank" rel="noopener noreferrer">{{.Name}}</a><span class="attachment-size">{{.Size}}</span></div>{{end}}
  {{end}}
  {{range .Stickers}}<div class="sticker">Sticker: {{.}}</div>{{end}}
  {{if .Reactions}}<div class="reactions">{{range .Reactions}}<span class="reaction">{{if .EmojiURL}}<img src="{{.EmojiURL}}" alt="{{.EmojiText}}">{{else}}{{.EmojiText}}{{end}} {{.Count}}</span>{{end}}</div>{{end}}
</div>
{{end}}
</main>
<footer>Transcript of #{{.ChannelName}} &middot; generated <time datetime="{{.GeneratedAt.ISO}}">{{.GeneratedAt.Text}}</time></footer>
<script>
document.querySelectorAll("time[datetime]").forEach(function (el) {
  var d = new Date(el.getAttribute("datetime"));
  if (isNaN(d)) return;
  el.textContent = el.dataset.format === "time"
    ? d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })
    : d.toLocaleString([], { dateStyle: "medium", timeStyle: "short" });
});
document.querySelectorAll(".spoiler").forEach(function (el) {
  el.addEventListener("click", function () { el.classList.add("revealed"); });
});
</script>
</body>
</html>
`))
