package engine

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/placeholder"
)

type MessageInstance struct {
	config               EngineConfig
	msg                  *model.MessageInstance
	flows                map[string]*flow.CompiledFlowNode
	appStore             store.AppStore
	logStore             store.LogStore
	messageStore         store.MessageStore
	messageInstanceStore store.MessageInstanceStore
	httpClient           *http.Client
}

func NewMessageInstance(
	config EngineConfig,
	msg *model.MessageInstance,
	appStore store.AppStore,
	logStore store.LogStore,
	messageStore store.MessageStore,
	messageInstanceStore store.MessageInstanceStore,
	httpClient *http.Client,
) (*MessageInstance, error) {
	flows := make(map[string]*flow.CompiledFlowNode, len(msg.FlowSources))

	for id, flowSource := range msg.FlowSources {
		flow, err := flow.CompileComponentButton(flowSource)
		if err != nil {
			slog.With("error", err).Error("Failed to compile component button flow")
			continue
		}

		flows[id] = flow
	}

	return &MessageInstance{
		config:               config,
		msg:                  msg,
		flows:                flows,
		appStore:             appStore,
		logStore:             logStore,
		messageStore:         messageStore,
		messageInstanceStore: messageInstanceStore,
		httpClient:           httpClient,
	}, nil
}

func (c *MessageInstance) HandleEvent(appID string, session *state.State, event gateway.Event) {
	defer c.recoverPanic()

	i, ok := event.(*gateway.InteractionCreateEvent)
	if !ok {
		return
	}

	d, ok := i.InteractionEvent.Data.(*discord.ButtonInteraction)
	if !ok {
		return
	}

	targetFlow, ok := c.flows[string(d.CustomID)]
	if !ok {
		return
	}

	providers := flow.FlowProviders{
		Discord:         NewDiscordProvider(appID, c.appStore, session),
		Log:             NewLogProvider(appID, c.logStore),
		HTTP:            NewHTTPProvider(c.httpClient),
		MessageTemplate: NewMessageTemplateProvider(c.messageStore, c.messageInstanceStore),
		// TODO: Variable provider
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fCtx := flow.NewContext(ctx,
		&InteractionData{
			interaction: &i.InteractionEvent,
		},
		providers,
		flow.FlowContextLimits{
			MaxStackDepth: c.config.MaxStackDepth,
			MaxOperations: c.config.MaxOperations,
			MaxActions:    c.config.MaxActions,
		},
		placeholder.NewEngine(),
	)

	if err := targetFlow.Execute(fCtx); err != nil {
		go c.createLogEntry(model.LogLevelError, fmt.Sprintf("Failed to execute command flow: %v", err))
		slog.With("error", err).Error("Failed to execute command flow")
	}
}

func (c *MessageInstance) createLogEntry(level model.LogLevel, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Create log entry which will be displayed in the dashboard
	err := c.logStore.CreateLogEntry(ctx, model.LogEntry{
		AppID:     "TODO", // TODO: appID
		Level:     level,
		Message:   message,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		slog.With("error", err).With("app_id", "TODO").Error("Failed to create log entry from engine command")
	}
}

func (c *MessageInstance) recoverPanic() {
	if r := recover(); r != nil {
		go c.createLogEntry(model.LogLevelError, fmt.Sprintf("Recovered from panic: %v", r))
		slog.With("error", r).
			With("app_id", "TODO").
			With("message_id", c.msg.MessageID).
			With("message_instance_id", c.msg.ID).
			Error("Recovered from panic in component handler")
	}
}
