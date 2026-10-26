package bundleinstall

import (
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	messagehandler "github.com/kitecloud/kite/kite-service/internal/api/handler/message"
	"github.com/kitecloud/kite/kite-service/internal/core/bundle"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

// A message the user changed while adding it is stored with their data, but
// its install item has the hash of the bundle's data, so it counts as changed
// and updates of the bundle keep it.
func TestPrepareContentsCustomizedMessage(t *testing.T) {
	def := message.MessageData{Content: "Default"}
	custom := message.MessageData{Content: "Mine"}
	sel := &bundle.Selection{Messages: []bundle.SelectedMessage{
		{ID: "sticky", Name: "Sticky", Data: custom, Customize: true, Default: &def},
		{ID: "panel", Name: "Panel", Data: def},
	}}

	h := NewBundleInstallHandler(nil, nil, nil, nil, nil, nil, nil, nil, messagehandler.NewMessageHandler(nil, nil, nil, nil), nil, 100)
	c := &handler.Context{App: &model.App{ID: "app"}, Session: &model.Session{UserID: "user"}}
	created, items, err := h.prepareContents(c, sel, bundleIDs{messages: map[string]string{}, variables: map[string]string{}}, null.StringFrom("install"))
	require.NoError(t, err)
	require.Len(t, created.Messages, 2)
	require.Len(t, items, 2)

	assert.Equal(t, custom, created.Messages[0].Data)
	defaultHash, err := (&model.Message{Data: def}).ContentHash()
	require.NoError(t, err)
	assert.Equal(t, defaultHash, items[0].Hash)
	currentHash, err := created.Messages[0].ContentHash()
	require.NoError(t, err)
	assert.NotEqual(t, currentHash, items[0].Hash)

	// Messages the user didn't change aren't edited.
	panelHash, err := created.Messages[1].ContentHash()
	require.NoError(t, err)
	assert.Equal(t, panelHash, items[1].Hash)
}
