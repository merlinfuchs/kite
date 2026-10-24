package bundleinstall

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/core/bundle"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

// HandleBundleExport exports items of the app as a bundle, which the user
// shares as a share code. The parameters of their flows become the bundle's,
// so whoever imports it fills in their own.
func (h *BundleInstallHandler) HandleBundleExport(c *handler.Context, req wire.BundleExportRequest) (*wire.BundleExportResponse, error) {
	if len(req.CommandIDs)+len(req.EventListenerIDs)+len(req.MessageIDs) == 0 {
		return nil, handler.ErrBadRequest("nothing_selected", "select at least one item to export")
	}

	all, err := h.appContents(c)
	if err != nil {
		return nil, err
	}

	var items bundle.ExportItems
	if items.Commands, err = pick(all.Commands, req.CommandIDs, func(c *model.Command) string { return c.ID }); err != nil {
		return nil, err
	}
	if items.EventListeners, err = pick(all.EventListeners, req.EventListenerIDs, func(l *model.EventListener) string { return l.ID }); err != nil {
		return nil, err
	}
	if items.Messages, err = pick(all.Messages, req.MessageIDs, func(m *model.Message) string { return m.ID }); err != nil {
		return nil, err
	}
	addUsed(&items, all.Messages, all.Variables)

	b, err := bundle.Export(c.App.Name, items)
	if err != nil {
		return nil, fmt.Errorf("failed to export bundle: %w", err)
	}
	content, err := json.Marshal(b)
	if err != nil {
		return nil, fmt.Errorf("failed to encode bundle: %w", err)
	}

	return &wire.BundleExportResponse{Content: content}, nil
}

// HandleBundleImport adds items of an exported bundle to the app with the
// parameters filled in. They go through the same checks as an install, but
// aren't linked to the bundle.
func (h *BundleInstallHandler) HandleBundleImport(c *handler.Context, req wire.BundleImportRequest) (*wire.BundleImportResponse, error) {
	b, err := bundle.Parse(req.Content)
	if err != nil {
		return nil, handler.ErrBadRequest("invalid_bundle", err.Error())
	}

	params := req.Parameters
	if params == nil {
		params = map[string]string{}
	}

	all, err := h.appContents(c)
	if err != nil {
		return nil, err
	}

	// Exports from this app use its message templates and variables instead
	// of copies, so importing doesn't pile them up.
	ids := bundleIDs{messages: map[string]string{}, variables: map[string]string{}}
	keys := make([]string, 0, len(req.Items))
	for _, key := range req.Items {
		switch {
		case hasMessage(b, key) && slices.ContainsFunc(all.Messages, func(m *model.Message) bool { return m.ID == key }):
			ids.messages[key] = key
		case hasVariable(b, key) && slices.ContainsFunc(all.Variables, func(v *model.Variable) bool { return v.ID == key }):
			ids.variables[key] = key
		default:
			keys = append(keys, key)
		}
	}

	sel, err := selectItems(b, keys, params, "invalid_bundle_import")
	if err != nil {
		return nil, err
	}

	created, items, err := h.prepareContents(c, sel, ids, null.String{})
	if err != nil {
		return nil, err
	}
	if err := h.bundleInstallStore.ImportBundleContents(c.Context(), created); err != nil {
		return nil, fmt.Errorf("failed to import bundle: %w", err)
	}

	res := &wire.BundleImportResponse{Items: make([]wire.BundleInstallItem, len(items))}
	for i, item := range items {
		res.Items[i] = wire.BundleInstallItem{Key: item.Key, Type: string(item.Type), ID: item.ID}
	}
	return res, nil
}

func hasMessage(b *bundle.Bundle, key string) bool {
	_, ok := b.Message(key)
	return ok
}

func hasVariable(b *bundle.Bundle, key string) bool {
	_, ok := b.Variable(key)
	return ok
}

// pick returns the items with the IDs, in the order of the IDs.
func pick[T comparable](all []T, ids []string, id func(T) string) ([]T, error) {
	res := make([]T, 0, len(ids))
	for _, want := range ids {
		i := slices.IndexFunc(all, func(item T) bool { return id(item) == want })
		if i == -1 {
			return nil, handler.ErrNotFound("unknown_item", fmt.Sprintf("the app has no item %s", want))
		}
		if !slices.Contains(res, all[i]) {
			res = append(res, all[i])
		}
	}
	return res, nil
}

// addUsed adds the message templates and variables the items use, also
// through the buttons of message templates. Blocks that use ones the app
// doesn't have anymore are cleared, they couldn't be imported.
func addUsed(items *bundle.ExportItems, messages []*model.Message, variables []*model.Variable) {
	var flows []*flow.FlowData
	for _, cmd := range items.Commands {
		flows = append(flows, &cmd.FlowSource)
	}
	for _, listener := range items.EventListeners {
		flows = append(flows, &listener.FlowSource)
	}
	addMessageFlows := func(msg *model.Message) {
		for _, flowSource := range msg.FlowSources {
			// The copy shares its nodes with the map, which is all that's
			// changed.
			flows = append(flows, &flowSource)
		}
	}
	for _, msg := range items.Messages {
		addMessageFlows(msg)
	}

	// Message templates add the flows of their buttons, so this goes on until
	// no new ones are found.
	for i := 0; i < len(flows); i++ {
		for j := range flows[i].Nodes {
			node := &flows[i].Nodes[j]
			if id := node.Data.MessageTemplateID; id != "" && !slices.ContainsFunc(items.Messages, func(m *model.Message) bool { return m.ID == id }) {
				k := slices.IndexFunc(messages, func(m *model.Message) bool { return m.ID == id })
				if k == -1 {
					node.Data.MessageTemplateID = ""
				} else {
					items.Messages = append(items.Messages, messages[k])
					addMessageFlows(messages[k])
				}
			}
			if id := node.Data.VariableID; id != "" && !slices.ContainsFunc(items.Variables, func(v *model.Variable) bool { return v.ID == id }) {
				k := slices.IndexFunc(variables, func(v *model.Variable) bool { return v.ID == id })
				if k == -1 {
					node.Data.VariableID = ""
				} else {
					items.Variables = append(items.Variables, variables[k])
				}
			}
		}
	}
}
