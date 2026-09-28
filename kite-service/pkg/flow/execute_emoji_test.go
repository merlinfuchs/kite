package flow

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The first bytes of a PNG file, enough for content sniffing.
var pngHeader = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")

func TestSanitizeEmojiName(t *testing.T) {
	assert.Equal(t, "party_parrot", sanitizeEmojiName("party parrot"))
	assert.Equal(t, "party_parrot", sanitizeEmojiName("  party-parrot "))
	assert.Equal(t, "cool_emoji", sanitizeEmojiName("cool_emoji!!"))
	assert.Len(t, sanitizeEmojiName("a_very_long_emoji_name_that_goes_past_the_limit"), 32)
}

func TestDecodeDataURI(t *testing.T) {
	uri := "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngHeader)
	content, err := decodeDataURI(uri)
	require.NoError(t, err)
	assert.Equal(t, pngHeader, content)

	_, err = decodeDataURI("data:image/png,notbase64")
	assert.Error(t, err)
	_, err = decodeDataURI("data:image/png;base64,!!!")
	assert.Error(t, err)
}

func TestDetectImageType(t *testing.T) {
	assert.Equal(t, "image/png", detectImageType(pngHeader))
	assert.Equal(t, "image/gif", detectImageType([]byte("GIF89a......")))
	assert.Equal(t, "application/json", detectImageType([]byte(` {"v":"5.5.2","layers":[]}`)))
	assert.Equal(t, "text/plain", detectImageType([]byte("hello")))
}

func TestEmojiMentionRe(t *testing.T) {
	m := emojiMentionRe.FindStringSubmatch("<:kite:123456789012345678>")
	require.NotNil(t, m)
	assert.Equal(t, "123456789012345678", m[1])

	m = emojiMentionRe.FindStringSubmatch("<a:dance:42>")
	require.NotNil(t, m)
	assert.Equal(t, "42", m[1])

	assert.Nil(t, emojiMentionRe.FindStringSubmatch("123"))
}
