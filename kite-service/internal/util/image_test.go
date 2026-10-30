package util

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func dataURI(mime string, data string) string {
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString([]byte(data))
}

func TestIsAnimatedImageDataURI(t *testing.T) {
	webp := func(flags byte) string {
		return "RIFF\x00\x00\x00\x00WEBPVP8X\x0a\x00\x00\x00" + string([]byte{flags}) + "\x00\x00\x00\x00\x00\x00\x00\x00\x00"
	}

	assert.True(t, IsAnimatedImageDataURI(dataURI("image/gif", "GIF89a\x01\x00\x01\x00")))
	// The image counts, not the type the URI claims.
	assert.True(t, IsAnimatedImageDataURI(dataURI("image/png", "GIF89a\x01\x00\x01\x00")))
	assert.True(t, IsAnimatedImageDataURI(dataURI("image/webp", webp(0x02))))
	assert.False(t, IsAnimatedImageDataURI(dataURI("image/webp", webp(0x10))))
	assert.False(t, IsAnimatedImageDataURI(dataURI("image/png", "\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")))
	assert.False(t, IsAnimatedImageDataURI(""))
}
