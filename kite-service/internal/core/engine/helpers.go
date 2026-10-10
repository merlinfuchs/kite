package engine

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/openai/openai-go/v2"
	"gopkg.in/guregu/null.v4"
)

type Env struct {
	Config               EngineConfig
	AppStore             store.AppStore
	FeatureProvider      FeatureProvider
	BlockRateLimiter     *BlockRateLimiter
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
	AppSecretStore       store.AppSecretStore
	AppIntegrationStore  store.AppIntegrationStore
	HttpClient           *http.Client
	OpenaiClient         *openai.Client
	TokenCrypt           *util.SymmetricCrypt
	CooldownProvider     provider.CooldownProvider
	// CreditLimiter enforces the credit limits app owners set for single
	// servers and users. Nil enforces none.
	CreditLimiter *CreditLimiter
}

type entityLinks struct {
	CommandID         null.String
	EventListenerID   null.String
	MessageID         null.String
	MessageInstanceID null.Int
	FlowSourceID      null.String // For message templates that have multiple flows
}

func (l entityLinks) usageRecordType() model.UsageRecordType {
	switch {
	case l.EventListenerID.Valid:
		return model.UsageRecordTypeEventListenerFlowExecution
	case l.MessageID.Valid:
		return model.UsageRecordTypeMessageFlowExecution
	default:
		return model.UsageRecordTypeCommandFlowExecution
	}
}

func (s Env) flowProviders(appID string, session *state.State, links entityLinks) flow.FlowProviders {
	var aiProvider provider.AIProvider = &provider.MockAIProvider{}
	if s.OpenaiClient != nil {
		aiProvider = NewAIProvider(s.OpenaiClient)
	}

	var cooldownProvider provider.CooldownProvider = &provider.MockCooldownProvider{}
	if s.CooldownProvider != nil {
		cooldownProvider = s.CooldownProvider
	}

	return flow.FlowProviders{
		Discord: NewDiscordProvider(appID, s.AppStore, s.FeatureProvider, s.BlockRateLimiter, session),
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
		Cooldown:        cooldownProvider,
		ResumePoint: NewResumePointProvider(
			s.ResumePointStore,
			s.TokenCrypt,
			appID,
			links,
		),
		Secret:      NewSecretProvider(appID, s.AppSecretStore, s.TokenCrypt),
		Integration: NewIntegrationProvider(appID, s.AppSecretStore, s.AppIntegrationStore, s.TokenCrypt),
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

	if state != nil {
		earlier := make([]eval.Context, len(state.Triggers))
		for i := range state.Triggers {
			earlier[i] = triggerEvalContext(&state.Triggers[i], session)
		}
		evalCtx.SetResumeContext(earlier)
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

	if s.creditLimitReached(ctx, appID, session, fCtx, links) {
		return
	}

	s.finishFlowRun(appID, links, fCtx, node.Execute(fCtx), "Failed to execute flow event")
}

// executeFlowAfterSleep continues a flow after the durable sleep in node. event
// is the interaction or event the flow ran with before it suspended.
func (s Env) executeFlowAfterSleep(
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

	// The sleep acknowledged the interaction before suspending, so responses
	// have to be follow-ups.
	if interaction := fCtx.Data.Interaction(); interaction != nil {
		fCtx.Discord.MarkInteractionResponded(interaction.ID)
	}

	s.finishFlowRun(appID, links, fCtx, node.ResumeAfterSleep(fCtx), "Failed to execute flow after sleep")
}

// finishFlowRun records the error and usage of a flow execution.
func (s Env) finishFlowRun(appID string, links entityLinks, fCtx *flow.FlowContext, err error, errMessage string) {
	if err != nil {
		s.createLogEntry(
			appID,
			model.LogLevelError,
			fmt.Sprintf("%s: %v", errMessage, err),
			links,
		)
	}

	guildID, userID := executionTargetIDs(fCtx)

	s.createUsageRecord(
		appID,
		fCtx.CreditsUsed(),
		links,
		guildID,
		userID,
	)

	if s.CreditLimiter != nil {
		s.CreditLimiter.Record(appID, fCtx.CreditsUsed(), creditLimitTargets(guildID.String, userID.String)...)
	}
}

// executionTargetIDs returns the server and user a flow runs for, if any.
func executionTargetIDs(fCtx *flow.FlowContext) (null.String, null.String) {
	var guildID, userID null.String
	if id := fCtx.Data.GuildID(); id.IsValid() {
		guildID = null.StringFrom(id.String())
	}
	if id := fCtx.Data.UserID(); id.IsValid() {
		userID = null.StringFrom(id.String())
	}
	return guildID, userID
}

// creditLimitReached checks the credit limits of the server and user the flow
// runs for. If one is reached the flow doesn't run, and the user is told why if
// they triggered it with an interaction. Failing to check lets the flow run, a
// database hiccup shouldn't take every bot down with it.
func (s Env) creditLimitReached(
	ctx context.Context,
	appID string,
	session *state.State,
	fCtx *flow.FlowContext,
	links entityLinks,
) bool {
	if s.CreditLimiter == nil {
		return false
	}

	guildID, userID := executionTargetIDs(fCtx)

	exceeded, err := s.CreditLimiter.Check(ctx, appID, creditLimitTargets(guildID.String, userID.String)...)
	if err != nil {
		logCreditLimitError(appID, err)
		return false
	}
	if exceeded == nil {
		return false
	}

	if exceeded.FirstReport {
		s.createLogEntry(
			appID,
			model.LogLevelWarn,
			fmt.Sprintf(
				"Credit limit reached for %s %s: %d of %d credits used per %s. Executions for it are skipped until the %s ends.",
				creditLimitScopeName(exceeded.Limit.Scope),
				exceeded.TargetID,
				exceeded.Used,
				exceeded.Limit.Credits.Int64,
				exceeded.Limit.Period,
				exceeded.Limit.Period,
			),
			links,
		)
	}

	interaction := fCtx.Data.Interaction()
	if interaction == nil || session == nil {
		return true
	}
	if _, ok := interaction.Data.(*discord.AutocompleteInteraction); ok {
		return true
	}

	err = session.RespondInteraction(interaction.ID, interaction.Token, api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Content: option.NewNullableString(creditLimitMessage(exceeded)),
			Flags:   discord.EphemeralMessage,
		},
	})
	if err != nil {
		slog.Error(
			"Failed to respond to interaction over credit limit",
			slog.String("app_id", appID),
			slog.String("error", err.Error()),
		)
	}

	return true
}

func creditLimitScopeName(scope model.CreditLimitScope) string {
	if scope == model.CreditLimitScopeGuild {
		return "server"
	}
	return "user"
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

func (s Env) createUsageRecord(appID string, creditsUsed int, links entityLinks, guildID null.String, userID null.String) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	start := time.Now()
	err := s.UsageStore.CreateUsageRecord(ctx, model.UsageRecord{
		AppID:           appID,
		Type:            links.usageRecordType(),
		CommandID:       links.CommandID,
		EventListenerID: links.EventListenerID,
		MessageID:       links.MessageID,
		GuildID:         guildID,
		UserID:          userID,
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
