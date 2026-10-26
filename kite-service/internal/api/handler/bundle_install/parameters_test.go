package bundleinstall

import (
	"net/http"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBundleInstallParametersUpdate(t *testing.T) {
	h, installs := moderationInstall(t)
	install := installs.installs[0]
	install.BundleVersion = officialBundle("moderation").Version
	install.Parameters = map[string]string{"reason": "Install value"}

	// Ban still has the install's value, kick was changed in the editor.
	commands := h.commandStore.(*memoryCommandStore).commands
	param := func(value string) []flow.FlowParameter {
		return []flow.FlowParameter{{Key: "reason", Label: "Reason", Type: flow.FlowParameterTypeText, Value: value}}
	}
	commands[0].FlowSource.Parameters = param("Install value")
	commands[2].FlowSource.Parameters = param("Changed in editor")

	set := handler.TypedWithBody(h.HandleBundleInstallParametersUpdate)
	code, res := call(t, set, "install", wire.BundleInstallParametersUpdateRequest{Parameters: map[string]string{"reason": "New"}})
	require.Equal(t, http.StatusOK, code, res)

	require.Len(t, installs.changes.Updated.Commands, 2)
	for _, cmd := range installs.changes.Updated.Commands {
		assert.Equal(t, "New", cmd.FlowSource.ParameterValues()["reason"], cmd.ID)
	}
	assert.Equal(t, "New", res["data"].(map[string]any)["parameters"].(map[string]any)["reason"])
	// A value isn't an edit, so updates still replace the commands.
	assert.Equal(t, "edited=false deleted=false", itemStates(res)["ban"])

	code, res = call(t, set, "install", wire.BundleInstallParametersUpdateRequest{Parameters: map[string]string{"other": "x"}})
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "unknown_parameter", res["error"].(map[string]any)["code"])

	commands[0].FlowSource.Parameters = []flow.FlowParameter{{Key: "log", Label: "Log", Type: flow.FlowParameterTypeChannel, Required: true}}
	code, res = call(t, set, "install", wire.BundleInstallParametersUpdateRequest{Parameters: map[string]string{"log": "general"}})
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "Log must be the ID of a channel", res["error"].(map[string]any)["message"])

	code, _ = call(t, set, "install", wire.BundleInstallParametersUpdateRequest{Parameters: map[string]string{"log": " "}})
	assert.Equal(t, http.StatusBadRequest, code)
}
