package flow

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

// Discord's upload limits. arikawa checks emojis against 256*1000 rather than
// 256 KiB, so both limits use decimal kilobytes to stay on the safe side.
const (
	MaxEmojiImageSize   = 256 * 1000
	MaxStickerImageSize = 512 * 1000
)

var (
	emojiImageTypes   = []string{"image/png", "image/jpeg", "image/gif", "image/webp"}
	stickerImageTypes = []string{"image/png", "image/gif", "application/json"}
)

// Matches <:name:id> and <a:name:id>.
var emojiMentionRe = regexp.MustCompile(`^<a?:[A-Za-z0-9_~]+:([0-9]+)>$`)

// Characters Discord doesn't allow in emoji names.
var invalidEmojiNameRe = regexp.MustCompile(`[^A-Za-z0-9_]`)

type loadedImage struct {
	ContentType string
	Content     []byte
}

func (i loadedImage) DataURI() string {
	return "data:" + i.ContentType + ";base64," + base64.StdEncoding.EncodeToString(i.Content)
}

func (i loadedImage) Extension() string {
	switch i.ContentType {
	case "image/png":
		return "png"
	case "image/jpeg":
		return "jpg"
	case "image/gif":
		return "gif"
	case "image/webp":
		return "webp"
	case "application/json":
		return "json"
	default:
		return "bin"
	}
}

// loadImage resolves the image of an emoji or sticker block. The source is a
// data URI, which is what uploads in the editor store, or a template that
// evaluates to an http(s) URL or a data URI, e.g. the URL of an attachment
// argument.
func (n *CompiledFlowNode) loadImage(ctx *FlowContext, source string, maxSize int, allowedTypes []string) (*loadedImage, error) {
	source = strings.TrimSpace(source)
	if source == "" {
		return nil, fmt.Errorf("an image is required")
	}

	// Uploads are stored as data URIs, which can be hundreds of kilobytes, so
	// they skip template evaluation.
	if !strings.HasPrefix(source, "data:") {
		evaluated, err := ctx.EvalTemplate(source)
		if err != nil {
			return nil, err
		}
		source = strings.TrimSpace(evaluated.String())
		if source == "" {
			return nil, fmt.Errorf("the image evaluated to an empty value")
		}
	}

	var content []byte
	if strings.HasPrefix(source, "data:") {
		decoded, err := decodeDataURI(source)
		if err != nil {
			return nil, err
		}
		content = decoded
	} else {
		fetched, err := n.fetchImage(ctx, source, maxSize)
		if err != nil {
			return nil, err
		}
		content = fetched
	}

	if len(content) == 0 {
		return nil, fmt.Errorf("the image is empty")
	}
	if len(content) > maxSize {
		return nil, fmt.Errorf("the image is %.1f KB, larger than the %d KB limit", float64(len(content))/1000, maxSize/1000)
	}

	contentType := detectImageType(content)
	allowed := false
	for _, t := range allowedTypes {
		if t == contentType {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, fmt.Errorf("unsupported image type %q, expected one of %s", contentType, strings.Join(allowedTypes, ", "))
	}

	return &loadedImage{ContentType: contentType, Content: content}, nil
}

func decodeDataURI(uri string) ([]byte, error) {
	meta, data, ok := strings.Cut(strings.TrimPrefix(uri, "data:"), ",")
	if !ok || !strings.HasSuffix(meta, ";base64") {
		return nil, fmt.Errorf("the image must be a base64 data URI")
	}

	content, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("the image is not valid base64: %w", err)
	}
	return content, nil
}

func (n *CompiledFlowNode) fetchImage(ctx *FlowContext, rawURL string, maxSize int) ([]byte, error) {
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, fmt.Errorf("the image must be an http(s) URL or an uploaded image")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := ctx.HTTP.HTTPRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	if resp == nil {
		return nil, fmt.Errorf("failed to download image: no response")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("failed to download image: %s", resp.Status)
	}

	// Read one byte past the limit to tell a file at the limit from a larger one
	// without buffering all of it.
	content, err := io.ReadAll(io.LimitReader(resp.Body, int64(maxSize)+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read image: %w", err)
	}
	if len(content) > maxSize {
		return nil, fmt.Errorf("the image is larger than the %d KB limit", maxSize/1000)
	}

	return content, nil
}

// detectImageType sniffs the content instead of trusting a declared type, as
// URLs and data URIs often carry a wrong or generic one.
func detectImageType(content []byte) string {
	contentType, _, _ := strings.Cut(http.DetectContentType(content), ";")

	// Lottie stickers are JSON, which sniffs as text.
	if strings.HasPrefix(contentType, "text/") {
		trimmed := strings.TrimSpace(string(content[:min(len(content), 64)]))
		if strings.HasPrefix(trimmed, "{") {
			return "application/json"
		}
	}
	return contentType
}

// sanitizeEmojiName makes a name Discord accepts: letters, numbers and
// underscores only, the same cleanup the Discord client does.
func sanitizeEmojiName(name string) string {
	name = strings.Join(strings.Fields(name), "_")
	name = strings.ReplaceAll(name, "-", "_")
	name = invalidEmojiNameRe.ReplaceAllString(name, "")
	if len(name) > 32 {
		name = name[:32]
	}
	return name
}

// targetSnowflake resolves an ID template. Besides plain IDs it accepts emoji
// mentions like <:name:123> and the results of the create blocks.
func targetSnowflake(ctx *FlowContext, template string) (discord.Snowflake, error) {
	value, err := ctx.EvalTemplate(template)
	if err != nil {
		return 0, err
	}

	if id := value.Snowflake(); id.IsValid() {
		return id, nil
	}

	s := strings.TrimSpace(value.String())
	if m := emojiMentionRe.FindStringSubmatch(s); m != nil {
		s = m[1]
	}

	id, err := strconv.ParseUint(s, 10, 64)
	if err != nil || !discord.Snowflake(id).IsValid() {
		return 0, fmt.Errorf("%q is not a valid ID", s)
	}
	return discord.Snowflake(id), nil
}

func (n *CompiledFlowNode) auditLogReason(ctx *FlowContext) (api.AuditLogReason, error) {
	reason, err := ctx.EvalTemplate(n.Data.AuditLogReason)
	if err != nil {
		return "", err
	}
	return api.AuditLogReason(reason.String()), nil
}

func emojiResult(guildID discord.GuildID, e discord.Emoji) thing.Thing {
	mention := "<:" + e.Name + ":" + e.ID.String() + ">"
	if e.Animated {
		mention = "<a:" + e.Name + ":" + e.ID.String() + ">"
	}

	return thing.NewObject(map[string]thing.Thing{
		"id":       thing.NewString(e.ID.String()),
		"name":     thing.NewString(e.Name),
		"animated": thing.NewBool(e.Animated),
		"mention":  thing.NewString(mention),
		"url":      thing.NewString(e.EmojiURL()),
		"guild_id": thing.NewString(guildID.String()),
	})
}

func stickerResult(s discord.Sticker) thing.Thing {
	ext := "png"
	switch s.FormatType {
	case discord.StickerFormatLottie:
		ext = "json"
	case 4: // GIF, which arikawa doesn't define.
		ext = "gif"
	}

	return thing.NewObject(map[string]thing.Thing{
		"id":          thing.NewString(s.ID.String()),
		"name":        thing.NewString(s.Name),
		"description": thing.NewString(s.Description),
		"tags":        thing.NewString(s.Tags),
		"url":         thing.NewString("https://media.discordapp.net/stickers/" + s.ID.String() + "." + ext),
		"guild_id":    thing.NewString(s.GuildID.String()),
	})
}

func (n *CompiledFlowNode) executeMemberTimeoutRemove(ctx *FlowContext) error {
	guildID, err := n.targetGuildID(ctx)
	if err != nil {
		return traceError(n, err)
	}

	userID, err := ctx.EvalTemplate(n.Data.UserTarget)
	if err != nil {
		return traceError(n, err)
	}

	reason, err := n.auditLogReason(ctx)
	if err != nil {
		return traceError(n, err)
	}

	// A zero timestamp marshals to null, which is how Discord lifts a timeout.
	var until discord.Timestamp
	err = ctx.Discord.EditMember(
		ctx,
		guildID,
		discord.UserID(userID.Snowflake()),
		api.ModifyMemberData{
			CommunicationDisabledUntil: &until,
			AuditLogReason:             reason,
		},
	)
	if err != nil {
		return traceError(n, err)
	}

	return n.ExecuteChildren(ctx)
}

func (n *CompiledFlowNode) executeEmojiCreate(ctx *FlowContext) error {
	data := n.Data.CustomEmojiData
	if data == nil {
		return traceError(n, fmt.Errorf("emoji data is required"))
	}

	guildID, err := n.targetGuildID(ctx)
	if err != nil {
		return traceError(n, err)
	}

	name, err := ctx.EvalTemplate(data.Name)
	if err != nil {
		return traceError(n, err)
	}
	emojiName := sanitizeEmojiName(name.String())
	if len(emojiName) < 2 {
		return traceError(n, fmt.Errorf("emoji name must be at least 2 letters, numbers or underscores"))
	}

	image, err := n.loadImage(ctx, data.Image, MaxEmojiImageSize, emojiImageTypes)
	if err != nil {
		return traceError(n, err)
	}

	reason, err := n.auditLogReason(ctx)
	if err != nil {
		return traceError(n, err)
	}

	emoji, err := ctx.Discord.CreateEmoji(ctx, guildID, provider.CreateEmojiData{
		Name:           emojiName,
		Image:          image.DataURI(),
		AuditLogReason: reason,
	})
	if err != nil {
		return traceError(n, err)
	}

	if emoji != nil {
		ctx.StoreNodeResult(n, emojiResult(guildID, *emoji))
	}
	return n.ExecuteChildren(ctx)
}

func (n *CompiledFlowNode) executeEmojiEdit(ctx *FlowContext) error {
	data := n.Data.CustomEmojiData
	if data == nil {
		return traceError(n, fmt.Errorf("emoji data is required"))
	}

	guildID, err := n.targetGuildID(ctx)
	if err != nil {
		return traceError(n, err)
	}

	emojiID, err := targetSnowflake(ctx, n.Data.EmojiTarget)
	if err != nil {
		return traceError(n, err)
	}

	name, err := ctx.EvalTemplate(data.Name)
	if err != nil {
		return traceError(n, err)
	}
	emojiName := sanitizeEmojiName(name.String())
	if len(emojiName) < 2 {
		return traceError(n, fmt.Errorf("emoji name must be at least 2 letters, numbers or underscores"))
	}

	reason, err := n.auditLogReason(ctx)
	if err != nil {
		return traceError(n, err)
	}

	emoji, err := ctx.Discord.EditEmoji(ctx, guildID, discord.EmojiID(emojiID), provider.EditEmojiData{
		Name:           emojiName,
		AuditLogReason: reason,
	})
	if err != nil {
		return traceError(n, err)
	}

	if emoji != nil {
		ctx.StoreNodeResult(n, emojiResult(guildID, *emoji))
	}
	return n.ExecuteChildren(ctx)
}

func (n *CompiledFlowNode) executeEmojiDelete(ctx *FlowContext) error {
	guildID, err := n.targetGuildID(ctx)
	if err != nil {
		return traceError(n, err)
	}

	emojiID, err := targetSnowflake(ctx, n.Data.EmojiTarget)
	if err != nil {
		return traceError(n, err)
	}

	reason, err := n.auditLogReason(ctx)
	if err != nil {
		return traceError(n, err)
	}

	err = ctx.Discord.DeleteEmoji(ctx, guildID, discord.EmojiID(emojiID), reason)
	if err != nil {
		return traceError(n, err)
	}

	return n.ExecuteChildren(ctx)
}

// stickerFields evaluates the text fields of a sticker block. Empty values
// mean the field was left blank.
func (n *CompiledFlowNode) stickerFields(ctx *FlowContext, data *GuildStickerData) (name, description, tags string, err error) {
	nameValue, err := ctx.EvalTemplate(data.Name)
	if err != nil {
		return
	}
	descriptionValue, err := ctx.EvalTemplate(data.Description)
	if err != nil {
		return
	}
	tagsValue, err := ctx.EvalTemplate(data.Tags)
	if err != nil {
		return
	}

	name = strings.TrimSpace(nameValue.String())
	description = strings.TrimSpace(descriptionValue.String())
	tags = strings.TrimSpace(tagsValue.String())

	if name != "" && (len([]rune(name)) < 2 || len([]rune(name)) > 30) {
		err = fmt.Errorf("sticker name must be 2 to 30 characters")
		return
	}
	if description != "" && (len([]rune(description)) < 2 || len([]rune(description)) > 100) {
		err = fmt.Errorf("sticker description must be empty or 2 to 100 characters")
		return
	}
	if len([]rune(tags)) > 200 {
		err = fmt.Errorf("sticker emoji must be at most 200 characters")
		return
	}
	return
}

func (n *CompiledFlowNode) executeStickerCreate(ctx *FlowContext) error {
	data := n.Data.GuildStickerData
	if data == nil {
		return traceError(n, fmt.Errorf("sticker data is required"))
	}

	guildID, err := n.targetGuildID(ctx)
	if err != nil {
		return traceError(n, err)
	}

	name, description, tags, err := n.stickerFields(ctx, data)
	if err != nil {
		return traceError(n, err)
	}
	if name == "" {
		return traceError(n, fmt.Errorf("sticker name is required"))
	}
	if tags == "" {
		return traceError(n, fmt.Errorf("sticker emoji is required"))
	}

	image, err := n.loadImage(ctx, data.Image, MaxStickerImageSize, stickerImageTypes)
	if err != nil {
		return traceError(n, err)
	}

	reason, err := n.auditLogReason(ctx)
	if err != nil {
		return traceError(n, err)
	}

	sticker, err := ctx.Discord.CreateSticker(ctx, guildID, provider.CreateStickerData{
		Name:           name,
		Description:    description,
		Tags:           tags,
		FileName:       "sticker." + image.Extension(),
		ContentType:    image.ContentType,
		File:           image.Content,
		AuditLogReason: reason,
	})
	if err != nil {
		return traceError(n, err)
	}

	if sticker != nil {
		ctx.StoreNodeResult(n, stickerResult(*sticker))
	}
	return n.ExecuteChildren(ctx)
}

func (n *CompiledFlowNode) executeStickerEdit(ctx *FlowContext) error {
	data := n.Data.GuildStickerData
	if data == nil {
		data = &GuildStickerData{}
	}

	guildID, err := n.targetGuildID(ctx)
	if err != nil {
		return traceError(n, err)
	}

	stickerID, err := targetSnowflake(ctx, n.Data.StickerTarget)
	if err != nil {
		return traceError(n, err)
	}

	name, description, tags, err := n.stickerFields(ctx, data)
	if err != nil {
		return traceError(n, err)
	}

	reason, err := n.auditLogReason(ctx)
	if err != nil {
		return traceError(n, err)
	}

	edit := provider.EditStickerData{AuditLogReason: reason}
	if name != "" {
		edit.Name = &name
	}
	if description != "" {
		edit.Description = &description
	}
	if tags != "" {
		edit.Tags = &tags
	}
	if edit.Name == nil && edit.Description == nil && edit.Tags == nil {
		return traceError(n, fmt.Errorf("set a name, description or emoji to edit"))
	}

	sticker, err := ctx.Discord.EditSticker(ctx, guildID, discord.StickerID(stickerID), edit)
	if err != nil {
		return traceError(n, err)
	}

	if sticker != nil {
		ctx.StoreNodeResult(n, stickerResult(*sticker))
	}
	return n.ExecuteChildren(ctx)
}

func (n *CompiledFlowNode) executeStickerDelete(ctx *FlowContext) error {
	guildID, err := n.targetGuildID(ctx)
	if err != nil {
		return traceError(n, err)
	}

	stickerID, err := targetSnowflake(ctx, n.Data.StickerTarget)
	if err != nil {
		return traceError(n, err)
	}

	reason, err := n.auditLogReason(ctx)
	if err != nil {
		return traceError(n, err)
	}

	err = ctx.Discord.DeleteSticker(ctx, guildID, discord.StickerID(stickerID), reason)
	if err != nil {
		return traceError(n, err)
	}

	return n.ExecuteChildren(ctx)
}
