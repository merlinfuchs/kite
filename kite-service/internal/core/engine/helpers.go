package engine

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/openai/openai-go"
	"gopkg.in/guregu/null.v4"
)

type Env struct {
	Config               EngineConfig
	AppStore             store.AppStore
	LogStore             store.LogStore
	UsageStore           store.UsageStore
	MessageStore         store.MessageStore
	MessageInstanceStore store.MessageInstanceStore
	CommandStore         store.CommandStore
	EventListenerStore   store.EventListenerStore
	PluginInstanceStore  store.PluginInstanceStore
	PluginValueStore     store.PluginValueStore
	PluginRegistry       *plugin.Registry
	VariableValueStore   store.VariableValueStore
	ResumePointStore     store.ResumePointStore
	HttpClient           *http.Client
	OpenaiClient         *openai.Client
	TokenCrypt           *util.SymmetricCrypt
}

type entityLinks struct {
	CommandID         null.String
	EventListenerID   null.String
	MessageID         null.String
	MessageInstanceID null.Int
	FlowSourceID      null.String // For message templates that have multiple flows
}

func (s Env) flowProviders(appID string, session *state.State, links entityLinks) flow.FlowProviders {
	var aiProvider provider.AIProvider = &provider.MockAIProvider{}
	if s.OpenaiClient != nil {
		aiProvider = NewAIProvider(s.OpenaiClient)
	}

	return flow.FlowProviders{
		Discord: NewDiscordProvider(appID, s.AppStore, session),
		Roblox:  NewRobloxProvider(s.HttpClient),
		Log: NewLogProvider(
			appID,
			s.LogStore,
			links,
		),
		HTTP:            NewHTTPProvider(s.HttpClient),
		AI:              aiProvider,
		MessageTemplate: NewMessageTemplateProvider(s.MessageStore, s.MessageInstanceStore),
		Variable:        NewVariableProvider(s.VariableValueStore),
		ResumePoint: NewResumePointProvider(
			s.ResumePointStore,
			appID,
			links,
		),
	}
}

func (s Env) flowContext(
	ctx context.Context,
	appID string,
	entryNodeID string,
	session *state.State,
	event gateway.Event,
	links entityLinks,
	state *flow.FlowContextState,
) *flow.FlowContext {
	providers := s.flowProviders(appID, session, links)

	var fCtx *flow.FlowContext

	switch e := event.(type) {
	case *gateway.InteractionCreateEvent:
		fCtx = flow.NewContext(
			ctx,
			30*time.Second,
			entryNodeID,
			&InteractionData{
				interaction: &e.InteractionEvent,
			},
			providers,
			flow.FlowContextLimits{
				MaxStackDepth: s.Config.MaxStackDepth,
				MaxOperations: s.Config.MaxOperations,
				MaxCredits:    s.Config.MaxCredits,
			},
			eval.NewContextFromInteraction(&e.InteractionEvent, session),
			state,
		)
	default:
		fCtx = flow.NewContext(
			ctx,
			30*time.Second,
			entryNodeID,
			&EventData{
				event: event,
			},
			providers,
			flow.FlowContextLimits{
				MaxStackDepth: s.Config.MaxStackDepth,
				MaxOperations: s.Config.MaxOperations,
				MaxCredits:    s.Config.MaxCredits,
			},
			eval.NewContextFromEvent(event, session),
			state,
		)
	}

	return fCtx
}

func (s Env) executeFlowEvent(
	ctx context.Context,
	appID string,
	node *flow.CompiledFlowNode,
	session *state.State,
	event gateway.Event,
	links entityLinks,
	state *flow.FlowContextState,
) {
	defer s.recoverPanic(appID, links)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	fCtx := s.flowContext(ctx, appID, node.ID, session, event, links, state)
	defer fCtx.Cancel()

	err := node.Execute(fCtx)
	if err != nil {
		slog.Error(
			"Failed to execute flow event",
			slog.String("app_id", appID),
			slog.String("command_id", links.CommandID.String),
			slog.String("message_id", links.MessageID.String),
			slog.String("event_listener_id", links.EventListenerID.String),
			slog.String("error", err.Error()),
		)

		s.createLogEntry(
			appID,
			model.LogLevelError,
			fmt.Sprintf("Failed to execute flow event: %v", err),
			links,
		)
	}

	s.createUsageRecord(
		appID,
		fCtx.CreditsUsed(),
		links,
	)
}

func (s Env) createLogEntry(appID string, level model.LogLevel, message string, links entityLinks) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Create log entry which will be displayed in the dashboard
	err := s.LogStore.CreateLogEntry(ctx, model.LogEntry{
		AppID:           appID,
		Level:           level,
		Message:         message,
		CommandID:       links.CommandID,
		EventListenerID: links.EventListenerID,
		MessageID:       links.MessageID,
		CreatedAt:       time.Now().UTC(),
	})
	if err != nil {
		slog.With("error", err).With("app_id", appID).Error("Failed to create log entry from engine")
	}
}

func (s Env) createUsageRecord(appID string, creditsUsed int, links entityLinks) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := s.UsageStore.CreateUsageRecord(ctx, model.UsageRecord{
		AppID:           appID,
		Type:            model.UsageRecordTypeCommandFlowExecution,
		CommandID:       links.CommandID,
		EventListenerID: links.EventListenerID,
		MessageID:       links.MessageID,
		CreditsUsed:     creditsUsed,
		CreatedAt:       time.Now().UTC(),
	})
	if err != nil {
		slog.With("error", err).With("app_id", appID).Error("Failed to create usage record from engine")
	}
}

func (s Env) recoverPanic(appID string, links entityLinks) {
	if r := recover(); r != nil {
		slog.With("error", r).
			With("app_id", appID).
			With("command_id", links.CommandID.String).
			With("message_id", links.MessageID.String).
			With("event_listener_id", links.EventListenerID.String).
			Error("Recovered from panic in engine handler")
		fmt.Println(fmt.Sprintf("%s", r), "\n", string(debug.Stack()))

		s.createLogEntry(
			appID,
			model.LogLevelError,
			fmt.Sprintf("Recovered from panic: %v", r),
			links,
		)
	}
}
