package model

import (
	"encoding/json"
	"slices"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/message"
	"gopkg.in/guregu/null.v4"
)

type Message struct {
	ID              string
	Name            string
	Description     null.String
	AppID           string
	BundleInstallID null.String
	CreatorUserID   string
	Data            message.MessageData
	FlowSources     map[string]flow.FlowData
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// ContentHash identifies the message's content and the flows of its buttons
// and select menus, like flow.FlowData.ContentHash does for flows.
func (m *Message) ContentHash() (string, error) {
	keys := make([]string, 0, len(m.FlowSources))
	for key := range m.FlowSources {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	flowHashes := make([][2]string, len(keys))
	for i, key := range keys {
		hash, err := m.FlowSources[key].ContentHash()
		if err != nil {
			return "", err
		}
		flowHashes[i] = [2]string{key, hash}
	}

	raw, err := json.Marshal(struct {
		Data  message.MessageData `json:"data"`
		Flows [][2]string         `json:"flows"`
	}{m.Data, flowHashes})
	if err != nil {
		return "", err
	}

	return util.HashBytes(raw), nil
}

type MessageInstance struct {
	ID               uint64
	MessageID        string
	DiscordGuildID   string
	DiscordChannelID string
	DiscordMessageID string
	Ephemeral        bool
	Hidden           bool
	FlowSources      map[string]flow.FlowData
	// TemplateFlowSources are the flows its message template has now, only
	// loaded with a single instance.
	TemplateFlowSources map[string]flow.FlowData
	CreatedAt           time.Time
	UpdatedAt           time.Time
	LastUsedAt          time.Time
}
