package engine

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/placeholder"
	"github.com/sashabaranov/go-openai"
)

type EventListener struct {
	config               EngineConfig
	listener             *model.EventListener
	flow                 *flow.CompiledFlowNode
	appStore             store.AppStore
	logStore             store.LogStore
	messageStore         store.MessageStore
	messageInstanceStore store.MessageInstanceStore
	variableValueStore   store.VariableValueStore
	httpClient           *http.Client
	openaiClient         *openai.Client
}

func NewEventListener(
	config EngineConfig,
	listener *model.EventListener,
	appStore store.AppStore,
	logStore store.LogStore,
	messageStore store.MessageStore,
	messageInstanceStore store.MessageInstanceStore,
	variableValueStore store.VariableValueStore,
	httpClient *http.Client,
	openaiClient *openai.Client,
) (*EventListener, error) {
	flow, err := flow.CompileEventListener(listener.FlowSource)
	if err != nil {
		return nil, fmt.Errorf("failed to compile event listener flow: %w", err)
	}

	return &EventListener{
		config:               config,
		listener:             listener,
		flow:                 flow,
		appStore:             appStore,
		logStore:             logStore,
		messageStore:         messageStore,
		messageInstanceStore: messageInstanceStore,
		variableValueStore:   variableValueStore,
		httpClient:           httpClient,
		openaiClient:         openaiClient,
	}, nil
}

func (l *EventListener) HandleEvent(appID string, session *state.State, event gateway.Event) {
	defer l.recoverPanic()

	// TODO: check listener specific filters as well
	if !l.shouldHandleEvent(event) {
		return
	}

	var aiProvider flow.FlowAIProvider = &flow.MockAIProvider{}
	if l.openaiClient != nil {
		aiProvider = NewAIProvider(l.openaiClient)
	}

	providers := flow.FlowProviders{
		Discord:         NewDiscordProvider(appID, l.appStore, session),
		Log:             NewLogProvider(appID, l.logStore),
		HTTP:            NewHTTPProvider(l.httpClient),
		AI:              aiProvider,
		MessageTemplate: NewMessageTemplateProvider(l.messageStore, l.messageInstanceStore),
		Variable:        NewVariableProvider(l.variableValueStore),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fCtx := flow.NewContext(ctx,
		&EventData{
			event: event,
		},
		providers,
		flow.FlowContextLimits{
			MaxStackDepth: l.config.MaxStackDepth,
			MaxOperations: l.config.MaxOperations,
			MaxActions:    l.config.MaxActions,
		},
		placeholder.NewEngine(),
	)

	if err := l.flow.Execute(fCtx); err != nil {
		go l.createLogEntry(model.LogLevelError, fmt.Sprintf("Failed to execute event listener flow: %v", err))
		slog.With("error", err).Error("Failed to execute event listener flow")
	}
}

func (l *EventListener) createLogEntry(level model.LogLevel, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Create log entry which will be displayed in the dashboard
	err := l.logStore.CreateLogEntry(ctx, model.LogEntry{
		AppID:     l.listener.AppID,
		Level:     level,
		Message:   message,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		slog.With("error", err).With("app_id", l.listener.AppID).Error("Failed to create log entry from engine event listener")
	}
}

func (l *EventListener) recoverPanic() {
	if r := recover(); r != nil {
		go l.createLogEntry(model.LogLevelError, fmt.Sprintf("Recovered from panic: %v", r))
		slog.With("error", r).
			With("app_id", l.listener.AppID).
			With("event_listener_id", l.listener.ID).
			Error("Recovered from panic in event listener handler")
	}
}

func (l *EventListener) shouldHandleEvent(e ws.Event) bool {
	switch d := e.(type) {
	case *gateway.MessageCreateEvent:
		// TODO?: It would be better if we check if the author is specifically the current app
		return !d.Author.Bot
	case *gateway.MessageUpdateEvent:
		return !d.Author.Bot
	case *gateway.MessageDeleteEvent:
		return true
	case *gateway.GuildMemberAddEvent:
		return true
	case *gateway.GuildMemberRemoveEvent:
		return true
	}

	return false
}
