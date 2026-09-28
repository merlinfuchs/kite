// Package voiceaudio turns an arbitrary audio file (e.g. mp3) into a sequence
// of Opus frames suitable for sending over a Discord voice connection, and
// streams them at real-time pace.
//
// Decoding is delegated to the "ffmpeg" binary (must be available on PATH),
// which avoids pulling an audio-decoding library into this Go module for
// every format ffmpeg already understands. Encoding to Opus uses libopus via
// cgo, which requires the libopus development headers at build time and the
// libopus shared library at runtime.
package voiceaudio

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os/exec"
	"time"

	"gopkg.in/hraban/opus.v2"
)

const (
	// SampleRate is the sample rate Discord voice expects.
	SampleRate = 48000
	// Channels is the channel count Discord voice expects (stereo).
	Channels = 2
	// FrameDuration is the Opus frame size used for Discord voice packets.
	FrameDuration = 20 * time.Millisecond
	// samplesPerFrame is the number of samples per channel in one frame:
	// 48000Hz * 20ms = 960 samples.
	samplesPerFrame = SampleRate * int(FrameDuration/time.Millisecond) / 1000
	// pcmFrameBytes is the size in bytes of one frame of interleaved 16-bit
	// stereo PCM: 960 samples * 2 channels * 2 bytes per sample.
	pcmFrameBytes = samplesPerFrame * Channels * 2
	// maxOpusFrameBytes is a safe upper bound for a single encoded Opus
	// packet at the bitrates used here.
	maxOpusFrameBytes = 4000
)

// DecodeToOpusFrames decodes an arbitrary audio file (anything ffmpeg can
// read, e.g. mp3) into a sequence of 20ms, 48kHz stereo Opus frames ready to
// be sent over a Discord voice connection.
func DecodeToOpusFrames(ctx context.Context, input []byte, volumePercent int) ([][]byte, error) {
	pcm, err := decodeToPCM(ctx, input)
	if err != nil {
		return nil, err
	}
	if len(pcm) == 0 {
		return nil, fmt.Errorf("decoded audio is empty")
	}

	if volumePercent < 0 {
		volumePercent = 0
	}
	if volumePercent > 200 {
		volumePercent = 200
	}
	volume := float64(volumePercent) / 100.0

	enc, err := opus.NewEncoder(SampleRate, Channels, opus.AppAudio)
	if err != nil {
		return nil, fmt.Errorf("failed to create opus encoder: %w", err)
	}

	var frames [][]byte
	pcmBuf := make([]int16, samplesPerFrame*Channels)
	opusBuf := make([]byte, maxOpusFrameBytes)

	r := bytes.NewReader(pcm)
	for {
		raw := make([]byte, pcmFrameBytes)
		n, readErr := io.ReadFull(r, raw)
		if n == 0 {
			if readErr == io.EOF {
				break
			}
			if readErr != nil {
				return nil, fmt.Errorf("failed to read pcm: %w", readErr)
			}
		}

		// raw is zero-initialized, so a short final read (ErrUnexpectedEOF)
		// is naturally padded with silence rather than encoding garbage.
		for i := range pcmBuf {
			sample := int16(binary.LittleEndian.Uint16(raw[i*2 : i*2+2]))
			if volume != 1 {
				boosted := math.Round(float64(sample) * volume)
				if boosted > 32767 {
					boosted = 32767
				} else if boosted < -32768 {
					boosted = -32768
				}
				sample = int16(boosted)
			}
			pcmBuf[i] = sample
		}

		encodedLen, err := enc.Encode(pcmBuf, opusBuf)
		if err != nil {
			return nil, fmt.Errorf("failed to encode opus frame: %w", err)
		}

		frame := make([]byte, encodedLen)
		copy(frame, opusBuf[:encodedLen])
		frames = append(frames, frame)

		if readErr == io.ErrUnexpectedEOF || readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("failed to read pcm: %w", readErr)
		}
	}

	return frames, nil
}

// decodeToPCM shells out to ffmpeg to turn an arbitrary audio file into raw
// signed 16-bit little-endian stereo PCM at 48kHz.
func decodeToPCM(ctx context.Context, input []byte) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-hide_banner", "-loglevel", "error", "-nostdin",
		"-i", "pipe:0",
		"-f", "s16le",
		"-ar", "48000",
		"-ac", "2",
		"pipe:1",
	)
	cmd.Stdin = bytes.NewReader(input)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := stderr.String()
		if len(msg) > 500 {
			msg = msg[:500]
		}
		return nil, fmt.Errorf("ffmpeg failed to decode audio (is ffmpeg installed?): %w: %s", err, msg)
	}

	return stdout.Bytes(), nil
}

// Stream writes each Opus frame to w, pacing writes at the frame duration so
// playback happens in roughly real time. It stops early and returns ctx's
// error if ctx is canceled before all frames are sent.
func Stream(ctx context.Context, w io.Writer, frames [][]byte) error {
	ticker := time.NewTicker(FrameDuration)
	defer ticker.Stop()

	for _, frame := range frames {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}

		if _, err := w.Write(frame); err != nil {
			return fmt.Errorf("failed to write opus frame: %w", err)
		}
	}

	return nil
}
