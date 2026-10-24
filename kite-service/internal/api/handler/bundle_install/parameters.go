package bundleinstall

import (
	"errors"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
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
	}
	for key, value := range req.Parameters {
		def, ok := defs[key]
		if !ok {
			return nil, handler.ErrBadRequest("unknown_parameter", fmt.Sprintf("no flow of the install has the parameter %s", key))
		}
		if err := checkParameterValue(def, value); err != nil {
			return nil, handler.ErrBadRequest("invalid_parameter", err.Error())
		}
	}

	var changed store.BundleInstallContents
	for _, item := range install.Items {
		if _, ok := contents.hash(install.ID, item); !ok {
			continue
		}
		switch item.Type {
		case model.BundleInstallItemTypeCommand:
			current := contents.commands[item.ID]
			if f, ok := withParameterValues(current.FlowSource, req.Parameters); ok {
				updated := *current
				updated.FlowSource = f
				updated.UpdatedAt = time.Now().UTC()
				changed.Commands = append(changed.Commands, &updated)
			}
		case model.BundleInstallItemTypeEventListener:
			current := contents.eventListeners[item.ID]
			if f, ok := withParameterValues(current.FlowSource, req.Parameters); ok {
				updated := *current
				updated.FlowSource = f
				updated.UpdatedAt = time.Now().UTC()
				changed.EventListeners = append(changed.EventListeners, &updated)
			}
		case model.BundleInstallItemTypeMessage:
			current := contents.messages[item.ID]
			flowSources := make(map[string]flow.FlowData, len(current.FlowSources))
			anyChanged := false
			for key, flowSource := range current.FlowSources {
				f, ok := withParameterValues(flowSource, req.Parameters)
				flowSources[key] = f
				anyChanged = anyChanged || ok
			}
			if anyChanged {
				updated := *current
				updated.FlowSources = flowSources
				updated.UpdatedAt = time.Now().UTC()
				changed.Messages = append(changed.Messages, &updated)
			}
		}
	}

	params := make(map[string]string, len(install.Parameters)+len(req.Parameters))
	maps.Copy(params, install.Parameters)
	maps.Copy(params, req.Parameters)
	install.Parameters = params
	install.UpdatedAt = time.Now().UTC()

	install, err = h.bundleInstallStore.UpdateBundleInstall(c.Context(), install, install.BundleVersion, store.BundleInstallChanges{Updated: changed})
	if err != nil {
		if errors.Is(err, store.ErrOutdated) {
			return nil, handler.ErrBadRequest("bundle_install_changed", "the install changed, reload the page and try again")
		}
		return nil, fmt.Errorf("failed to update install: %w", err)
	}

	if err := contents.replace(changed); err != nil {
		return nil, err
	}
	return contents.toWire(install), nil
}

// checkParameterValue checks a value like adding a template does.
func checkParameterValue(p flow.FlowParameter, value string) error {
	label := p.Label
	if label == "" {
		label = p.Key
	}
	if p.Required && strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", label)
	}
	if value != "" && (p.Type == flow.FlowParameterTypeChannel || p.Type == flow.FlowParameterTypeRole) {
		if id, err := discord.ParseSnowflake(value); err != nil || !id.IsValid() {
			return fmt.Errorf("%s must be the ID of a %s", label, p.Type)
		}
	}
	return nil
}

// withParameterValues returns the flow with the values for the parameters it
// has, and whether one of them changed.
func withParameterValues(data flow.FlowData, values map[string]string) (flow.FlowData, bool) {
	params := make([]flow.FlowParameter, len(data.Parameters))
	changed := false
	for i, p := range data.Parameters {
		if value, ok := values[p.Key]; ok && value != p.Value {
			p.Value = value
			changed = true
		}
		params[i] = p
	}
	data.Parameters = params
	return data, changed
}
