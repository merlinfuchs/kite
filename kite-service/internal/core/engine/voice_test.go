package engine

import (
	"context"
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/internal/core/sound"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type voiceTestAssetStore struct {
	store.AssetStore
	asset *model.Asset
}

func (s *voiceTestAssetStore) AssetWithContent(ctx context.Context, id string) (*model.Asset, error) {
	if s.asset == nil || s.asset.ID != id {
		return nil, store.ErrNotFound
	}
	return s.asset, nil
}

func TestVoiceManagerPlayInvalidSound(t *testing.T) {
	tests := []struct {
		name  string
		asset *model.Asset
		err   string
	}{
		{
			name: "deleted",
			err:  "doesn't exist anymore",
		},
		{
			name:  "other app",
			asset: &model.Asset{ID: "sound", AppID: "other", ContentType: sound.ContentType},
			err:   "doesn't exist anymore",
		},
		{
			name:  "not a sound",
			asset: &model.Asset{ID: "sound", AppID: "app", ContentType: "audio/mpeg"},
			err:   "isn't a sound",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewVoiceManager(&voiceTestAssetStore{asset: tt.asset})
			err := m.Play(context.Background(), "app", nil, provider.PlayAudioOpts{GuildID: 1, ChannelID: 2, AssetID: "sound"})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.err)
			assert.Empty(t, m.players, "nothing joins")
		})
	}
}

func TestVoiceManagerStopWithoutPlayer(t *testing.T) {
	m := NewVoiceManager(nil)
	m.Stop("app", 1)
	m.Leave("app", 1)
	assert.Empty(t, m.players)
}

func TestVoiceManagerPlayerLimit(t *testing.T) {
	m := NewVoiceManager(nil)
	for i := range maxVoicePlayers {
		m.players[voiceKey{appID: "app", guildID: discord.GuildID(i + 1)}] = &voicePlayer{}
	}

	_, err := m.player(voiceKey{appID: "app", guildID: 9999}, true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too many bots")
}
