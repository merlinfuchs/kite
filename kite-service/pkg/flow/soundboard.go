package flow

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

// maxSoundboardSoundSize is Discord's upload limit for soundboard sounds.
const maxSoundboardSoundSize = 512 * 1024

// soundboardSoundCreateData evaluates the block's inputs into the request
// body for Discord. The sound is downloaded from the evaluated URL, which
// lets users pass an attachment argument like {{arg('sound')}}.
func (n *CompiledFlowNode) soundboardSoundCreateData(ctx *FlowContext) (provider.CreateSoundboardSoundData, error) {
	var res provider.CreateSoundboardSoundData

	data := n.Data.SoundboardSoundData
	if data == nil {
		return res, fmt.Errorf("soundboard sound data is required")
	}

	name, err := ctx.EvalTemplate(data.Name)
	if err != nil {
		return res, err
	}
	res.Name = strings.TrimSpace(name.String())
	if n := len([]rune(res.Name)); n < 2 || n > 32 {
		return res, fmt.Errorf("sound name must be between 2 and 32 characters")
	}

	if data.Volume != "" {
		volume, err := ctx.EvalTemplate(data.Volume)
		if err != nil {
			return res, err
		}

		v, err := strconv.ParseFloat(strings.TrimSpace(volume.String()), 64)
		if err != nil || v < 0 || v > 1 {
			return res, fmt.Errorf("sound volume must be a number between 0 and 1")
		}
		res.Volume = &v
	}

	if emoji := n.Data.EmojiData; emoji != nil {
		if emoji.ID != "" {
			id, err := discord.ParseSnowflake(emoji.ID)
			if err != nil || !id.IsValid() {
				return res, fmt.Errorf("invalid emoji ID %q", emoji.ID)
			}
			res.EmojiID = discord.EmojiID(id)
		} else {
			res.EmojiName = emoji.Name
		}
	}

	soundURL, err := ctx.EvalTemplate(data.Sound)
	if err != nil {
		return res, err
	}

	res.Sound, err = downloadSoundboardSound(ctx, strings.TrimSpace(soundURL.String()))
	if err != nil {
		return res, err
	}

	auditLogReason, err := ctx.EvalTemplate(n.Data.AuditLogReason)
	if err != nil {
		return res, err
	}
	res.AuditLogReason = api.AuditLogReason(auditLogReason.String())

	return res, nil
}

// downloadSoundboardSound fetches an MP3 or OGG file and returns it as the
// data URI Discord expects.
func downloadSoundboardSound(ctx *FlowContext, url string) (string, error) {
	if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "http://") {
		return "", fmt.Errorf("sound must be the URL of an audio file or an attachment")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("invalid sound URL: %w", err)
	}

	resp, err := ctx.HTTP.HTTPRequest(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to download sound: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("failed to download sound: status %d", resp.StatusCode)
	}
	if resp.ContentLength > maxSoundboardSoundSize {
		return "", fmt.Errorf("sound file is larger than 512 KB")
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxSoundboardSoundSize+1))
	if err != nil {
		return "", fmt.Errorf("failed to download sound: %w", err)
	}
	if len(body) > maxSoundboardSoundSize {
		return "", fmt.Errorf("sound file is larger than 512 KB")
	}
	if len(body) == 0 {
		return "", fmt.Errorf("sound file is empty")
	}

	contentType := soundboardSoundContentType(body, resp.Header.Get("Content-Type"))
	if contentType == "" {
		return "", fmt.Errorf("sound must be an MP3 or OGG audio file")
	}

	return "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(body), nil
}

// soundboardSoundContentType returns audio/mpeg or audio/ogg, the only types
// Discord accepts, or an empty string. The file's own bytes win over the
// Content-Type header, which CDNs don't always set.
func soundboardSoundContentType(body []byte, header string) string {
	switch {
	case bytes.HasPrefix(body, []byte("OggS")):
		return "audio/ogg"
	case bytes.HasPrefix(body, []byte("ID3")):
		return "audio/mpeg"
	case len(body) > 1 && body[0] == 0xFF && body[1]&0xE0 == 0xE0:
		// MPEG audio frame sync
		return "audio/mpeg"
	}

	mediaType, _, _ := mime.ParseMediaType(header)
	switch mediaType {
	case "audio/mpeg", "audio/mp3":
		return "audio/mpeg"
	case "audio/ogg", "application/ogg":
		return "audio/ogg"
	}

	return ""
}

func newSoundboardSoundThing(sound provider.SoundboardSound) thing.Thing {
	emojiID := ""
	if sound.EmojiID.IsValid() {
		emojiID = sound.EmojiID.String()
	}

	return thing.NewObject(map[string]thing.Thing{
		"id":         thing.NewString(sound.SoundID.String()),
		"name":       thing.NewString(sound.Name),
		"volume":     thing.NewFloat(sound.Volume),
		"emoji_id":   thing.NewString(emojiID),
		"emoji_name": thing.NewString(sound.EmojiName),
		"guild_id":   thing.NewString(sound.GuildID.String()),
	})
}
