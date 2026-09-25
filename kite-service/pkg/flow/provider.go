package flow

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/provider"
)

type FlowProviders struct {
	Discord         provider.DiscordProvider
	Roblox          provider.RobloxProvider
	HTTP            provider.HTTPProvider
	AI              provider.AIProvider
	Log             provider.LogProvider
	Variable        provider.VariableProvider
	MessageTemplate provider.MessageTemplateProvider
	ResumePoint     ResumePointProvider
}

type ResumePointProvider interface {
	CreateResumePoint(ctx context.Context, p ResumePoint) (ResumePoint, error)
}

type MockResumePointProvider struct{}

func (m *MockResumePointProvider) CreateResumePoint(ctx context.Context, p ResumePoint) (ResumePoint, error) {
	return ResumePoint{}, nil
}

type ResumePointType string

const (
	ResumePointTypeModal             ResumePointType = "modal"
	ResumePointTypeMessageComponents ResumePointType = "message_components"
	ResumePointTypeTimer             ResumePointType = "timer"
)

type ResumePoint struct {
	ID     string
	Type   ResumePointType
	NodeID string
	State  FlowContextState

	// ResumeAt and InteractionToken are only set for timers.
	ResumeAt         time.Time
	InteractionToken string
}
