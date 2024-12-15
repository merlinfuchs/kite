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
	"github.com/sashabaranov/go-openai"
	"gopkg.in/guregu/null.v4"
)

type MessageInstance struct {
	config               EngineConfig
	appID                string
	msg                  *model.MessageInstance
	flows                map[string]*flow.CompiledFlowNode
	appStore             store.AppStore
	logStore             store.LogStore
	usageStore           store.UsageStore
	messageStore         store.MessageStore
	messageInstanceStore store.MessageInstanceStore
	variableValueStore   store.VariableValueStore
	httpClient           *http.Client
	openaiClient         *openai.Client
}

func NewMessageInstance(
	config EngineConfig,
	appID string,
	msg *model.MessageInstance,
	appStore store.AppStore,
	logStore store.LogStore,
	usageStore store.UsageStore,
	messageStore store.MessageStore,
	messageInstanceStore store.MessageInstanceStore,
	variableValueStore store.VariableValueStore,
	httpClient *http.Client,
	openaiClient *openai.Client,
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
		appID:                appID,
		msg:                  msg,
		flows:                flows,
		appStore:             appStore,
		logStore:             logStore,
		usageStore:           usageStore,
		messageStore:         messageStore,
		messageInstanceStore: messageInstanceStore,
		variableValueStore:   variableValueStore,
		httpClient:           httpClient,
		openaiClient:         openaiClient,
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

	var aiProvider flow.FlowAIProvider = &flow.MockAIProvider{}
	if c.openaiClient != nil {
		aiProvider = NewAIProvider(c.openaiClient)
	}

	providers := flow.FlowProviders{
		Discord:         NewDiscordProvider(appID, c.appStore, session),
		Log:             NewLogProvider(appID, c.logStore),
		HTTP:            NewHTTPProvider(c.httpClient),
		AI:              aiProvider,
		MessageTemplate: NewMessageTemplateProvider(c.messageStore, c.messageInstanceStore),
		Variable:        NewVariableProvider(c.variableValueStore),
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
			MaxCredits:    c.config.MaxCredits,
		},
		placeholder.NewEngine(),
	)

	if err := targetFlow.Execute(fCtx); err != nil {
		slog.With("error", err).Error("Failed to execute message instance flow")
		c.createLogEntry(model.LogLevelError, fmt.Sprintf("Failed to execute message instance flow: %v", err))
	}

	c.createUsageRecord(fCtx.CreditsUsed())
}

func (c *MessageInstance) createLogEntry(level model.LogLevel, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Create log entry which will be displayed in the dashboard
	err := c.logStore.CreateLogEntry(ctx, model.LogEntry{
		AppID:     c.appID,
		Level:     level,
		Message:   message,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		slog.With("error", err).With("app_id", c.appID).Error("Failed to create log entry from engine message instance")
	}
}

func (c *MessageInstance) createUsageRecord(creditsUsed int) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := c.usageStore.CreateUsageRecord(ctx, model.UsageRecord{
		AppID:       c.appID,
		Type:        model.UsageRecordTypeFlowExecution,
		MessageID:   null.NewString(c.msg.MessageID, true),
		CreditsUsed: uint32(creditsUsed),
		CreatedAt:   time.Now().UTC(),
	})
	if err != nil {
		slog.With("error", err).With("app_id", c.appID).Error("Failed to create usage record from engine message instance")
	}
}

func (c *MessageInstance) recoverPanic() {
	if r := recover(); r != nil {
		go c.createLogEntry(model.LogLevelError, fmt.Sprintf("Recovered from panic: %v", r))
		slog.With("error", r).
			With("app_id", c.appID).
			With("message_id", c.msg.MessageID).
			With("message_instance_id", c.msg.ID).
			Error("Recovered from panic in component handler")
	}
}
