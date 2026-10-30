// Package sound converts uploaded audio files to sounds that bots can play in
// voice channels. Converting happens once on upload, so playing a sound only
// sends its stored Opus frames.
package sound

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/diamondburned/oggreader"
)

const (
	// ContentType is the content type of the assets of sounds.
	ContentType = "audio/ogg"
	// MaxDuration is how long a sound can be.
	MaxDuration = 2 * time.Minute
	// FrameDuration is the duration of each Opus frame of a sound. It matches
	// the frames a voice connection sends.
	FrameDuration = 20 * time.Millisecond

	// convertTimeout is how long converting a file may take.
	convertTimeout = 30 * time.Second
	// maxConversions is how many files can be converted at once.
	maxConversions = 4
	// formats are the formats ffmpeg may read uploads in. Formats like
	// playlists can reference other files and URLs, which mustn't be opened.
	formats = "mov,mp4,m4a,3gp,3g2,mj2,mp3,wav,ogg,flac,aac,matroska,webm"
	// maxStderrSize is how much of ffmpeg's error output is kept.
	maxStderrSize = 1024
	// silenceFrames is how many Opus frames of silence Play sends after the
	// sound, so clients don't interpolate its last frame.
	silenceFrames = 5
)

// opusSilence is an Opus frame of silence.
var opusSilence = []byte{0xF8, 0xFF, 0xFE}

var (
	ErrFFmpegMissing = errors.New("converting sounds needs ffmpeg, which isn't installed on the server")
	// ErrInvalid is returned if the file isn't audio that can be converted.
	ErrInvalid = errors.New("the file isn't an audio file that can be read")
	ErrTooLong = fmt.Errorf("the sound can be at most %d minutes long", int(MaxDuration.Minutes()))
)

var conversions = make(chan struct{}, maxConversions)

// Convert converts an audio file in one of the allowed formats to the Ogg Opus
// file of a sound, with 20ms frames and normalized loudness.
func Convert(ctx context.Context, audio io.Reader) ([]byte, error) {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, ErrFFmpegMissing
	}

	select {
	case conversions <- struct{}{}:
		defer func() { <-conversions }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	ctx, cancel := context.WithTimeout(ctx, convertTimeout)
	defer cancel()

	// ffmpeg reads from a file instead of stdin, because formats like M4A can
	// have their index at the end, which needs seeking.
	input, err := os.CreateTemp("", "kite-sound-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(input.Name())
	defer input.Close()

	if _, err := io.Copy(input, audio); err != nil {
		return nil, fmt.Errorf("failed to write temporary file: %w", err)
	}

	cmd := exec.CommandContext(ctx, ffmpeg,
		"-hide_banner", "-loglevel", "error",
		"-protocol_whitelist", "file",
		"-format_whitelist", formats,
		"-i", input.Name(),
		// A bit more than the maximum, to tell longer files apart.
		"-t", strconv.Itoa(int((MaxDuration + time.Second).Seconds())),
		"-map", "0:a:0",
		"-filter:a", "loudnorm=I=-18:TP=-2:LRA=11",
		"-c:a", "libopus",
		"-b:a", "96k",
		"-ar", "48000",
		"-ac", "2",
		"-frame_duration", strconv.Itoa(int(FrameDuration.Milliseconds())),
		"-application", "audio",
		"-threads", "1",
		"-f", "opus",
		"pipe:1",
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	var stderr limitedBuffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("converting the sound took too long: %w", ctx.Err())
		}
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return nil, fmt.Errorf("%w: %s", ErrInvalid, msg)
		}
		return nil, fmt.Errorf("%w: %w", ErrInvalid, err)
	}

	var frames int
	err = readFrames(bytes.NewReader(stdout.Bytes()), func([]byte) error {
		frames++
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalid, err)
	}
	if frames == 0 {
		return nil, ErrInvalid
	}
	if time.Duration(frames)*FrameDuration > MaxDuration {
		return nil, ErrTooLong
	}

	return stdout.Bytes(), nil
}

// Play writes the Opus frames of the sound to w, a voice connection that
// sends one frame every 20ms, until the sound ended or ctx is done. Stopping
// early isn't an error.
func Play(ctx context.Context, w io.Writer, ogg []byte) error {
	err := readFrames(bytes.NewReader(ogg), func(frame []byte) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		_, err := w.Write(frame)
		return err
	})
	if ctx.Err() != nil {
		return nil
	}
	if err != nil {
		return err
	}

	for range silenceFrames {
		if _, err := w.Write(opusSilence); err != nil {
			return err
		}
	}
	return nil
}

// readFrames calls fn with each Opus frame of the Ogg Opus file of a sound.
func readFrames(ogg io.Reader, fn func(frame []byte) error) error {
	var header bool
	err := oggreader.DecodeBuffered(frameWriter(func(b []byte) error {
		switch {
		case bytes.HasPrefix(b, []byte("OpusHead")):
			header = true
			return nil
		case bytes.HasPrefix(b, []byte("OpusTags")), len(b) == 0:
			return nil
		case !header:
			return errors.New("not an Opus file")
		}
		return fn(b)
	}), ogg)
	if err != nil {
		return err
	}
	if !header {
		return errors.New("not an Opus file")
	}
	return nil
}

type frameWriter func(b []byte) error

func (w frameWriter) Write(b []byte) (int, error) {
	if err := w(b); err != nil {
		return 0, err
	}
	return len(b), nil
}

// limitedBuffer keeps the first maxStderrSize bytes written to it.
type limitedBuffer struct {
	buf bytes.Buffer
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if room := maxStderrSize - b.buf.Len(); room > 0 {
		b.buf.Write(p[:min(len(p), room)])
	}
	return len(p), nil
}

func (b *limitedBuffer) String() string {
	return b.buf.String()
}
