package bundleinstall

import (
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
)

// HandleBundleInstallParametersUpdate sets parameters of an install in every
// flow it created that uses them, including where the user changed them in
// the editor. Like a change in the editor it isn't an edit of the flows, so
// updates still replace them.
func (h *BundleInstallHandler) HandleBundleInstallParametersUpdate(c *handler.Context, req wire.BundleInstallParametersUpdateRequest) (*wire.BundleInstallParametersUpdateResponse, error) {
	install, err := h.bundleInstall(c)
	if err != nil {
		return nil, err
	}

	contents, err := h.installedContents(c)
	if err != nil {
		return nil, err
	}

	// The flows define the parameters, the first one for a key is used.
	defs := map[string]flow.FlowParameter{}
	var changed store.BundleInstallContents
	now := time.Now().UTC()
	for _, item := range install.Items {
		if _, ok := contents.hash(install.ID, item); !ok {
			continue
		}
		for _, f := range contents.flows(item) {
			for _, p := range f.Parameters {
				if _, ok := defs[p.Key]; !ok {
					defs[p.Key] = p
				}
			}
		}

		switch item.Type {
		case model.BundleInstallItemTypeCommand:
			updated := *contents.commands[item.ID]
			if f, ok := updated.FlowSource.WithParameterValues(req.Parameters); ok {
				updated.FlowSource = f
				updated.UpdatedAt = now
				changed.Commands = append(changed.Commands, &updated)
			}
		case model.BundleInstallItemTypeEventListener:
			updated := *contents.eventListeners[item.ID]
			if f, ok := updated.FlowSource.WithParameterValues(req.Parameters); ok {
				updated.FlowSource = f
				updated.UpdatedAt = now
				changed.EventListeners = append(changed.EventListeners, &updated)
			}
		case model.BundleInstallItemTypeMessage:
			updated := *contents.messages[item.ID]
			updated.FlowSources = maps.Clone(updated.FlowSources)
			anyChanged := false
			for key, flowSource := range updated.FlowSources {
				f, ok := flowSource.WithParameterValues(req.Parameters)
				updated.FlowSources[key] = f
				anyChanged = anyChanged || ok
			}
			if anyChanged {
				updated.UpdatedAt = now
				changed.Messages = append(changed.Messages, &updated)
			}
		}
	}

	for key, value := range req.Parameters {
		def, ok := defs[key]
		if !ok {
			return nil, handler.ErrBadRequest("unknown_parameter", fmt.Sprintf("no flow of the install has the parameter %s", key))
		}
		if def.Required && strings.TrimSpace(value) == "" {
			return nil, handler.ErrBadRequest("invalid_parameter", fmt.Sprintf("%s is required", def.DisplayLabel()))
		}
		if err := def.CheckValue(value); err != nil {
			return nil, handler.ErrBadRequest("invalid_parameter", err.Error())
		}
	}

	params := make(map[string]string, len(install.Parameters)+len(req.Parameters))
	maps.Copy(params, install.Parameters)
	maps.Copy(params, req.Parameters)
	install.Parameters = params

	install, err = h.updateInstall(c, install, store.BundleInstallChanges{Updated: changed})
	if err != nil {
		return nil, err
	}

	if err := contents.replace(changed); err != nil {
		return nil, err
	}
	return contents.toWire(install), nil
}
