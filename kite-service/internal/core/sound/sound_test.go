package sound

import (
	"bytes"
	"context"
	"encoding/binary"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testWAV is a mono 16-bit PCM WAV file with a sine tone of the duration.
func testWAV(duration time.Duration, sampleRate int) []byte {
	samples := int(duration.Seconds() * float64(sampleRate))

	data := make([]byte, samples*2)
	for i := range samples {
		v := int16(math.Sin(2*math.Pi*440*float64(i)/float64(sampleRate)) * 8000)
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

func requireFFmpeg(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg isn't installed")
	}
}

func TestConvert(t *testing.T) {
	requireFFmpeg(t)

	ogg, err := Convert(context.Background(), bytes.NewReader(testWAV(time.Second, 48000)))
	require.NoError(t, err)

	var rec frameRecorder
	require.NoError(t, Play(context.Background(), &rec, ogg))

	// One second of 20ms frames without the header packets, then the silence.
	assert.InDelta(t, 50+silenceFrames, len(rec.frames), 2)
	for _, frame := range rec.frames {
		assert.False(t, bytes.HasPrefix(frame, []byte("Opus")))
	}
	assert.Equal(t, opusSilence, rec.frames[len(rec.frames)-1])
}

type frameRecorder struct {
	frames [][]byte
	// cancel, if set, is called after 5 frames.
	cancel context.CancelFunc
}

func (r *frameRecorder) Write(b []byte) (int, error) {
	if len(r.frames) == 5 && r.cancel != nil {
		r.cancel()
	}
	r.frames = append(r.frames, bytes.Clone(b))
	return len(b), nil
}

func TestPlayCancel(t *testing.T) {
	requireFFmpeg(t)

	ogg, err := Convert(context.Background(), bytes.NewReader(testWAV(5*time.Second, 48000)))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rec := frameRecorder{cancel: cancel}
	assert.NoError(t, Play(ctx, &rec, ogg), "a stopped playback isn't an error")
	assert.Len(t, rec.frames, 6, "no frames after stopping, and no silence")
}

func TestConvertTooLong(t *testing.T) {
	requireFFmpeg(t)

	_, err := Convert(context.Background(), bytes.NewReader(testWAV(MaxDuration+5*time.Second, 8000)))
	assert.ErrorIs(t, err, ErrTooLong)
}

func TestConvertInvalid(t *testing.T) {
	requireFFmpeg(t)

	_, err := Convert(context.Background(), strings.NewReader("this is not audio"))
	assert.ErrorIs(t, err, ErrInvalid)
}

func TestPlayNotOpus(t *testing.T) {
	var rec frameRecorder
	assert.Error(t, Play(context.Background(), &rec, []byte("this is not ogg")))
	assert.Empty(t, rec.frames)
}

func TestConvertPlaylist(t *testing.T) {
	requireFFmpeg(t)

	playlist := "#EXTM3U\n#EXT-X-TARGETDURATION:1\n#EXTINF:1,\n/etc/hosts\n#EXT-X-ENDLIST\n"
	_, err := Convert(context.Background(), strings.NewReader(playlist))
	assert.ErrorIs(t, err, ErrInvalid, "playlists aren't read")
}

func TestConvertM4AIndexAtEnd(t *testing.T) {
	requireFFmpeg(t)

	// ffmpeg writes the index of MP4 files at the end unless told otherwise.
	m4a := filepath.Join(t.TempDir(), "sound.m4a")
	wav := filepath.Join(t.TempDir(), "sound.wav")
	require.NoError(t, os.WriteFile(wav, testWAV(30*time.Second, 48000), 0o600))
	out, err := exec.Command("ffmpeg", "-loglevel", "error", "-i", wav, "-c:a", "aac", m4a).CombinedOutput()
	require.NoError(t, err, string(out))

	f, err := os.Open(m4a)
	require.NoError(t, err)
	defer f.Close()

	ogg, err := Convert(context.Background(), f)
	require.NoError(t, err)

	var rec frameRecorder
	require.NoError(t, Play(context.Background(), &rec, ogg))
	assert.Greater(t, len(rec.frames), 1400, "all 30 seconds are converted")
}
