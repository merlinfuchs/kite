package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/message"
	"gopkg.in/guregu/null.v4"
)

// AppDuplicateStores are the stores of everything a copy of an app includes.
type AppDuplicateStores struct {
	Commands        store.CommandStore
	EventListeners  store.EventListenerStore
	Messages        store.MessageStore
	Variables       store.VariableStore
	PluginInstances store.PluginInstanceStore
	Secrets         store.AppSecretStore
	Assets          store.AssetStore
}

// appContent is what a copy of an app includes. Variable values, logs, usage,
// collaborators, integrations and premium aren't copied.
type appContent struct {
	commands        []*model.Command
	eventListeners  []*model.EventListener
	messages        []*model.Message
	variables       []*model.Variable
	pluginInstances []*model.PluginInstance
	secrets         []*model.AppSecret
}

func (h *AppHandler) HandleAppDuplicate(c *handler.Context, req wire.AppDuplicateRequest) (*wire.AppDuplicateResponse, error) {
	if !c.UserAppRole.CanDuplicateApp() {
		return nil, handler.ErrForbidden("missing_permissions", "Only the owner can duplicate this app")
	}

	appCount, err := h.appStore.CountAppsByUser(c.Context(), c.Session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to count apps: %w", err)
	}
	if appCount >= h.maxAppsPerUser {
		return nil, handler.ErrBadRequest("resource_limit", fmt.Sprintf("maximum number of apps (%d) reached", h.maxAppsPerUser))
	}

	appInfo, err := h.getDiscordAppInfo(c.Context(), req.DiscordToken)
	if err != nil {
		if util.IsDiscordRestStatusCode(err, http.StatusUnauthorized) {
			return nil, handler.ErrBadRequest("invalid_discord_token", "Discord didn't accept this token. Copy it again from Bot in the Discord Developer Portal.")
		}
		return nil, fmt.Errorf("failed to get discord app info: %w", err)
	}

	if appInfo.ID == c.App.DiscordID {
		return nil, handler.ErrBadRequest("same_discord_app", "This token belongs to the bot of this app. A copy needs its own bot, so create a new application in the Discord Developer Portal and use its token.")
	}

	content, err := h.loadAppContent(c.Context(), c.App.ID)
	if err != nil {
		return nil, err
	}

	// The copy starts on the free plan, because premium belongs to the app it
	// was bought for.
	if err := checkContentLimits(content, h.planManager.DefaultFeatures()); err != nil {
		return nil, err
	}

	encryptedToken, err := h.tokenCrypt.EncryptString(req.DiscordToken)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt discord token: %w", err)
	}

	app, err := h.appStore.CreateApp(c.Context(), &model.App{
		ID:            util.UniqueID(),
		Name:          appInfo.Name,
		Description:   appInfo.Description,
		Enabled:       true,
		OwnerUserID:   c.Session.UserID,
		CreatorUserID: c.Session.UserID,
		DiscordToken:  encryptedToken,
		DiscordID:     appInfo.ID,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	})
	if err != nil {
		if errors.Is(err, store.ErrAlreadyExists) {
			return nil, handler.ErrBadRequest("app_exists", "This bot is already in Kite. A copy needs a bot that isn't in Kite yet, so create a new application in the Discord Developer Portal and use its token.")
		}
		return nil, fmt.Errorf("failed to create app: %w", err)
	}

	appID := app.ID
	err = h.copyAppContent(c.Context(), content, appID, c.Session.UserID)
	if err == nil && c.App.DiscordStatus != nil {
		// The status isn't set when an app is created.
		app, err = h.appStore.UpdateApp(c.Context(), store.AppUpdateOpts{
			ID:            appID,
			Name:          app.Name,
			Description:   app.Description,
			DiscordToken:  app.DiscordToken,
			DiscordStatus: c.App.DiscordStatus,
			Enabled:       app.Enabled,
			UpdatedAt:     time.Now().UTC(),
		})
	}
	if err != nil {
		// Don't leave a half finished copy behind.
		if err := h.appStore.DeleteApp(context.WithoutCancel(c.Context()), appID); err != nil {
			slog.Error(
				"Failed to delete incomplete app copy",
				slog.String("app_id", appID),
				slog.String("error", err.Error()),
			)
		}

		slog.Error(
			"Failed to duplicate app",
			slog.String("app_id", c.App.ID),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to duplicate app: %w", err)
	}

	return wire.AppToWire(app), nil
}

func (h *AppHandler) loadAppContent(ctx context.Context, appID string) (*appContent, error) {
	var (
		res appContent
		err error
	)

	if res.commands, err = h.duplicateStores.Commands.CommandsByApp(ctx, appID); err != nil {
		return nil, fmt.Errorf("failed to get commands: %w", err)
	}
	if res.eventListeners, err = h.duplicateStores.EventListeners.EventListenersByApp(ctx, appID); err != nil {
		return nil, fmt.Errorf("failed to get event listeners: %w", err)
	}
	if res.messages, err = h.duplicateStores.Messages.MessagesByApp(ctx, appID); err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	if res.variables, err = h.duplicateStores.Variables.VariablesByAppWithoutTotals(ctx, appID); err != nil {
		return nil, fmt.Errorf("failed to get variables: %w", err)
	}
	if res.pluginInstances, err = h.duplicateStores.PluginInstances.PluginInstancesByApp(ctx, appID); err != nil {
		return nil, fmt.Errorf("failed to get plugin instances: %w", err)
	}

	secrets, err := h.duplicateStores.Secrets.AppSecretsByApp(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to get secrets: %w", err)
	}
	for _, secret := range secrets {
		// Credentials of integrations only work with the integration, which
		// has to be connected again for the copy.
		if secret.AppIntegrationID == "" {
			res.secrets = append(res.secrets, secret)
		}
	}

	return &res, nil
}

func checkContentLimits(content *appContent, features model.Features) error {
	listeners := map[model.EventSource]int{}
	for _, listener := range content.eventListeners {
		listeners[listener.Source]++
	}
	scheduled := listeners[model.EventSourceSchedule]
	delete(listeners, model.EventSourceSchedule)
	maxListeners := 0
	for _, count := range listeners {
		maxListeners = max(maxListeners, count)
	}

	limits := []struct {
		name  string
		count int
		limit int
	}{
		{"commands", len(content.commands), features.MaxCommands},
		{"variables", len(content.variables), features.MaxVariables},
		{"message templates", len(content.messages), features.MaxMessages},
		{"secrets", len(content.secrets), features.MaxSecrets},
		{"event listeners", maxListeners, features.MaxEventListeners},
		{"scheduled event listeners", scheduled, features.MaxScheduledEventListeners},
	}

	for _, l := range limits {
		if l.limit != 0 && l.count > l.limit {
			return handler.ErrBadRequest("resource_limit", fmt.Sprintf(
				"This app has %d %s, but a new app can have up to %d without premium. Remove some before duplicating it.",
				l.count, l.name, l.limit,
			))
		}
	}

	return nil
}

func (h *AppHandler) copyAppContent(ctx context.Context, content *appContent, appID string, userID string) error {
	now := time.Now().UTC()

	ids := &duplicateIDs{
		variables: map[string]string{},
		messages:  map[string]string{},
		assets:    map[string]string{},
	}
	for _, variable := range content.variables {
		ids.variables[variable.ID] = util.UniqueID()
	}
	for _, msg := range content.messages {
		ids.messages[msg.ID] = util.UniqueID()
	}

	for _, variable := range content.variables {
		_, err := h.duplicateStores.Variables.CreateVariable(ctx, &model.Variable{
			ID:        ids.variables[variable.ID],
			Name:      variable.Name,
			Scoped:    variable.Scoped,
			AppID:     appID,
			CreatedAt: now,
			UpdatedAt: now,
		})
		if err != nil {
			return fmt.Errorf("failed to create variable: %w", err)
		}
	}

	for _, msg := range content.messages {
		data := msg.Data.Copy()
		if err := h.copyAttachments(ctx, ids, data.Attachments, appID, userID); err != nil {
			return err
		}

		flowSources := make(map[string]flow.FlowData, len(msg.FlowSources))
		for key, flowSource := range msg.FlowSources {
			remapped, err := h.remapFlow(ctx, ids, flowSource, appID, userID)
			if err != nil {
				return err
			}
			flowSources[key] = remapped
		}

		_, err := h.duplicateStores.Messages.CreateMessage(ctx, &model.Message{
			ID:            ids.messages[msg.ID],
			Name:          msg.Name,
			Description:   msg.Description,
			AppID:         appID,
			CreatorUserID: userID,
			Data:          data,
			FlowSources:   flowSources,
			CreatedAt:     now,
			UpdatedAt:     now,
		})
		if err != nil {
			return fmt.Errorf("failed to create message: %w", err)
		}
	}

	for _, command := range content.commands {
		flowSource, err := h.remapFlow(ctx, ids, command.FlowSource, appID, userID)
		if err != nil {
			return err
		}

		_, err = h.duplicateStores.Commands.CreateCommand(ctx, &model.Command{
			ID:            util.UniqueID(),
			Name:          command.Name,
			Description:   command.Description,
			Enabled:       command.Enabled,
			AppID:         appID,
			CreatorUserID: userID,
			FlowSource:    flowSource,
			CreatedAt:     now,
			UpdatedAt:     now,
		})
		if err != nil {
			return fmt.Errorf("failed to create command: %w", err)
		}
	}

	for _, listener := range content.eventListeners {
		flowSource, err := h.remapFlow(ctx, ids, listener.FlowSource, appID, userID)
		if err != nil {
			return err
		}

		// The copy gets its own webhook URL, so requests meant for the
		// original don't run it.
		webhookSecret := null.String{}
		if listener.Source == model.EventSourceWebhook {
			webhookSecret = null.NewString(util.SecureURLKey(), true)
		}

		_, err = h.duplicateStores.EventListeners.CreateEventListener(ctx, &model.EventListener{
			ID:            util.UniqueID(),
			Source:        listener.Source,
			Type:          listener.Type,
			Description:   listener.Description,
			Enabled:       listener.Enabled,
			AppID:         appID,
			CreatorUserID: userID,
			Filter:        listener.Filter,
			FlowSource:    flowSource,
			CreatedAt:     now,
			UpdatedAt:     now,
			WebhookSecret: webhookSecret,
		})
		if err != nil {
			return fmt.Errorf("failed to create event listener: %w", err)
		}
	}

	for _, instance := range content.pluginInstances {
		_, err := h.duplicateStores.PluginInstances.CreatePluginInstance(ctx, &model.PluginInstance{
			ID:                 util.UniqueID(),
			PluginID:           instance.PluginID,
			Enabled:            instance.Enabled,
			AppID:              appID,
			CreatorUserID:      userID,
			Config:             instance.Config,
			EnabledResourceIDs: instance.EnabledResourceIDs,
			CreatedAt:          now,
			UpdatedAt:          now,
		})
		if err != nil {
			return fmt.Errorf("failed to create plugin instance: %w", err)
		}
	}

	for _, secret := range content.secrets {
		// Both apps are encrypted with the same key, so the value is copied
		// as it is.
		_, err := h.duplicateStores.Secrets.CreateAppSecret(ctx, &model.AppSecret{
			ID:             util.UniqueID(),
			AppID:          appID,
			Name:           secret.Name,
			ValueEncrypted: secret.ValueEncrypted,
			CreatedAt:      now,
			UpdatedAt:      now,
		})
		if err != nil {
			return fmt.Errorf("failed to create secret: %w", err)
		}
	}

	return nil
}

// duplicateIDs maps the IDs of the original app to the IDs of their copies.
type duplicateIDs struct {
	variables map[string]string
	messages  map[string]string
	assets    map[string]string
}

// remapFlow points the blocks of a flow at the copies of the variables,
// message templates and attachments they use. References to things that don't
// exist in the original app are kept, like they were.
func (h *AppHandler) remapFlow(ctx context.Context, ids *duplicateIDs, source flow.FlowData, appID string, userID string) (flow.FlowData, error) {
	nodes := make([]flow.FlowNode, len(source.Nodes))
	for i, node := range source.Nodes {
		if id, ok := ids.variables[node.Data.VariableID]; ok {
			node.Data.VariableID = id
		}
		if id, ok := ids.messages[node.Data.MessageTemplateID]; ok {
			node.Data.MessageTemplateID = id
		}
		if node.Data.MessageData != nil {
			data := node.Data.MessageData.Copy()
			if err := h.copyAttachments(ctx, ids, data.Attachments, appID, userID); err != nil {
				return flow.FlowData{}, err
			}
			node.Data.MessageData = &data
		}
		nodes[i] = node
	}

	return flow.FlowData{Nodes: nodes, Edges: source.Edges}, nil
}

// copyAttachments copies the assets of the attachments to the new app, so
// they keep working when the original is deleted. Attachments are changed in
// place.
func (h *AppHandler) copyAttachments(ctx context.Context, ids *duplicateIDs, attachments []message.MessageAttachment, appID string, userID string) error {
	for i, attachment := range attachments {
		if attachment.AssetID == "" {
			continue
		}

		if id, ok := ids.assets[attachment.AssetID]; ok {
			attachments[i].AssetID = id
			continue
		}

		asset, err := h.duplicateStores.Assets.AssetWithContent(ctx, attachment.AssetID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				continue
			}
			return fmt.Errorf("failed to get asset: %w", err)
		}

		copied, err := h.duplicateStores.Assets.CreateAsset(ctx, &model.Asset{
			ID:            util.UniqueID(),
			AppID:         appID,
			CreatorUserID: userID,
			Name:          asset.Name,
			ContentType:   asset.ContentType,
			ContentHash:   asset.ContentHash,
			ContentSize:   asset.ContentSize,
			Content:       asset.Content,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			ExpiresAt:     asset.ExpiresAt,
		})
		if err != nil {
			return fmt.Errorf("failed to copy asset: %w", err)
		}

		ids.assets[attachment.AssetID] = copied.ID
		attachments[i].AssetID = copied.ID
	}

	return nil
}
