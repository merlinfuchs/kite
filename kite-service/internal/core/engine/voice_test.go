package engine

import (
	"bytes"
	"context"
	"encoding/binary"
	"math"
	"os/exec"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testWAV is a mono 16-bit PCM WAV file with a sine tone of the duration.
func testWAV(duration time.Duration) []byte {
	const sampleRate = 48000
	samples := int(duration.Seconds() * sampleRate)

	data := make([]byte, samples*2)
	for i := range samples {
		v := int16(math.Sin(2*math.Pi*440*float64(i)/sampleRate) * 8000)
		binary.LittleEndian.PutUint16(data[i*2:], uint16(v))
	}

	var b bytes.Buffer
	b.WriteString("RIFF")
	binary.Write(&b, binary.LittleEndian, uint32(36+len(data)))
	b.WriteString("WAVEfmt ")
	binary.Write(&b, binary.LittleEndian, uint32(16))
	binary.Write(&b, binary.LittleEndian, uint16(1)) // PCM
	binary.Write(&b, binary.LittleEndian, uint16(1)) // mono
	binary.Write(&b, binary.LittleEndian, uint32(sampleRate))
	binary.Write(&b, binary.LittleEndian, uint32(sampleRate*2))
	binary.Write(&b, binary.LittleEndian, uint16(2))
	binary.Write(&b, binary.LittleEndian, uint16(16))
	b.WriteString("data")
	binary.Write(&b, binary.LittleEndian, uint32(len(data)))
	b.Write(data)
	return b.Bytes()
}

type frameRecorder struct {
	frames [][]byte
}

func (r *frameRecorder) Write(b []byte) (int, error) {
	r.frames = append(r.frames, bytes.Clone(b))
	return len(b), nil
}

func lookFFmpeg(t *testing.T) string {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg isn't installed")
	}
	return ffmpeg
}

func TestStreamAudio(t *testing.T) {
	ffmpeg := lookFFmpeg(t)

	var rec frameRecorder
	err := streamAudio(context.Background(), &rec, ffmpeg, provider.PlayAudioOpts{
		Audio:  testWAV(time.Second),
		Volume: 1,
	})
	require.NoError(t, err)

	// One second of 20ms frames, then the silence.
	require.GreaterOrEqual(t, len(rec.frames), 50+silenceFrames)
	for _, frame := range rec.frames {
		assert.False(t, bytes.HasPrefix(frame, []byte("Opus")), "header packets are skipped")
	}
	assert.Equal(t, opusSilence, rec.frames[len(rec.frames)-1])
}

func TestStreamAudioInvalid(t *testing.T) {
	ffmpeg := lookFFmpeg(t)

	var rec frameRecorder
	err := streamAudio(context.Background(), &rec, ffmpeg, provider.PlayAudioOpts{
		Audio:  []byte("this is not audio"),
		Volume: 1,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ffmpeg failed")
	assert.Empty(t, rec.frames)
}

// cancelingRecorder cancels the playback after a few frames.
type cancelingRecorder struct {
	frameRecorder
	cancel context.CancelFunc
}

func (r *cancelingRecorder) Write(b []byte) (int, error) {
	if len(r.frames) == 5 {
		r.cancel()
	}
	return r.frameRecorder.Write(b)
}

func TestStreamAudioCancel(t *testing.T) {
	ffmpeg := lookFFmpeg(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rec := cancelingRecorder{cancel: cancel}
	err := streamAudio(ctx, &rec, ffmpeg, provider.PlayAudioOpts{
		Audio:  testWAV(10 * time.Second),
		Volume: 1,
	})
	assert.NoError(t, err, "a stopped playback isn't an error")
	assert.Less(t, len(rec.frames), 500, "the playback stopped early")
}

func TestVoiceManagerStopWithoutPlayer(t *testing.T) {
	m := NewVoiceManager()
	m.Stop("app", 1)
	assert.False(t, m.Leave("app", 1))
}

func TestVoiceManagerPlayerLimit(t *testing.T) {
	m := NewVoiceManager()
	for i := range maxVoicePlayers {
		m.players[voiceKey{appID: "app", guildID: discord.GuildID(i + 1)}] = &voicePlayer{}
	}

	_, err := m.acquire(voiceKey{appID: "app", guildID: 9999})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too many bots")
}
