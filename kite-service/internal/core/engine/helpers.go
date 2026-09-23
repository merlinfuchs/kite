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
		MessageTemplate: NewMessageTemplateProvider(appID, s.MessageStore, s.MessageInstanceStore),
		Variable:        NewVariableProvider(appID, s.VariableValueStore),
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
	session *state.State,
	event gateway.Event,
	links entityLinks,
	state *flow.FlowContextState,
) *flow.FlowContext {
	providers := s.flowProviders(appID, session, links)

	var data flow.FlowContextData
	var evalCtx eval.Context

	switch e := event.(type) {
	case *gateway.InteractionCreateEvent:
		data = &InteractionData{
			interaction: &e.InteractionEvent,
		}
		evalCtx = eval.NewContextFromInteraction(&e.InteractionEvent, session)
	default:
		data = &EventData{
			event: event,
		}
		evalCtx = eval.NewContextFromEvent(event, session)
	}

	if state != nil && len(state.Triggers) > 0 {
		origin := triggerEvalContext(state.Origin(), session)
		previous := origin
		if len(state.Triggers) > 1 {
			previous = triggerEvalContext(state.Previous(), session)
		}
		evalCtx.SetResumeContext(origin, previous, state.ModalInputs())
	}

	return flow.NewContext(
		ctx,
		30*time.Second,
		data,
		providers,
		flow.FlowContextLimits{
			MaxStackDepth: s.Config.MaxStackDepth,
			MaxOperations: s.Config.MaxOperations,
			MaxCredits:    s.Config.MaxCredits,
		},
		evalCtx,
		state,
	)
}

func triggerEvalContext(trigger *flow.FlowTrigger, session *state.State) eval.Context {
	if trigger.Interaction != nil {
		return eval.NewContextFromInteraction(trigger.Interaction, session)
	}
	return eval.NewContextFromEvent(trigger.Event, session)
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

	fCtx := s.flowContext(ctx, appID, session, event, links, state)
	defer fCtx.Cancel()

	shouldExecute, err := node.FilterEvent(fCtx)
	if err != nil {
		s.createLogEntry(
			appID,
			model.LogLevelError,
			fmt.Sprintf("Failed to filter events: %v", err),
			links,
		)
		return
	}

	if !shouldExecute {
		return
	}

	err = node.Execute(fCtx)
	if err != nil {
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	start := time.Now()
	err := s.UsageStore.CreateUsageRecord(ctx, model.UsageRecord{
		AppID:           appID,
		Type:            model.UsageRecordTypeCommandFlowExecution,
		CommandID:       links.CommandID,
		EventListenerID: links.EventListenerID,
		MessageID:       links.MessageID,
		CreditsUsed:     creditsUsed,
		CreatedAt:       time.Now().UTC(),
	})
	duration := time.Since(start)

	if duration > time.Second*10 {
		slog.With("duration", duration).
			With("app_id", appID).
			Warn("Usage record creation took longer than 10 seconds")
	}

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
