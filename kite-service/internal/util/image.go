package util

import (
	"bytes"
	"encoding/base64"
	"strings"
)

// IsAnimatedImageDataURI reports whether a data URI holds an image Discord
// shows animated, a GIF or an animated WebP. It looks at the image itself, not
// the type the URI claims. One that doesn't decode counts as animated, so it
// can't get past a check for animated images.
func IsAnimatedImageDataURI(uri string) bool {
	_, data, ok := strings.Cut(uri, ";base64,")
	if !ok {
		return false
	}

	head, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return true
	}

	if bytes.HasPrefix(head, []byte("GIF8")) {
		return true
	}

	// An animated WebP starts with a VP8X chunk that has the animation flag set.
	return len(head) >= 21 &&
		string(head[0:4]) == "RIFF" &&
		string(head[8:16]) == "WEBPVP8X" &&
		head[20]&0x02 != 0
}
