package engine

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
	"github.com/sashabaranov/go-openai"
	"gopkg.in/guregu/null.v4"
)

type Env struct {
	Config               EngineConfig
	AppStateManager      store.AppStateManager
	AppStore             store.AppStore
	LogStore             store.LogStore
	UsageStore           store.UsageStore
	MessageStore         store.MessageStore
	MessageInstanceStore store.MessageInstanceStore
	CommandStore         store.CommandStore
	EventListenerStore   store.EventListenerStore
	VariableValueStore   store.VariableValueStore
	ResumePointStore     store.ResumePointStore
	PluginInstanceStore  store.PluginInstanceStore
	PluginValueStore     store.PluginValueStore
	PluginRegistry       *plugin.Registry
	HttpClient           *http.Client
	OpenaiClient         *openai.Client
}

type entityLinks struct {
	CommandID         null.String
	EventListenerID   null.String
	MessageID         null.String
	MessageInstanceID null.Int
	FlowSourceID      null.String // For message templates that have multiple flows
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
			Error("Recovered from panic in message instance handler")

		s.createLogEntry(
			appID,
			model.LogLevelError,
			fmt.Sprintf("Recovered from panic: %v", r),
			links,
		)
	}
}
