package flow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSoundboardSoundContentType(t *testing.T) {
	tests := []struct {
		name   string
		body   []byte
		header string
		want   string
	}{
		{"ogg magic", []byte("OggS\x00\x02"), "", "audio/ogg"},
		{"mp3 id3 tag", []byte("ID3\x04\x00"), "application/octet-stream", "audio/mpeg"},
		{"mp3 frame sync", []byte{0xFF, 0xFB, 0x90, 0x00}, "", "audio/mpeg"},
		{"header fallback mp3", []byte("????"), "audio/mpeg", "audio/mpeg"},
		{"header fallback ogg with params", []byte("????"), "audio/ogg; codecs=opus", "audio/ogg"},
		{"wav rejected", []byte("RIFF\x00\x00\x00\x00WAVE"), "audio/wav", ""},
		{"image rejected", []byte("\x89PNG"), "image/png", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, soundboardSoundContentType(tt.body, tt.header))
		})
	}
}
