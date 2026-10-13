package flow

import (
	"context"
	"fmt"
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
	Cooldown        provider.CooldownProvider
	ResumePoint     ResumePointProvider
	Secret          provider.SecretProvider
	Integration     provider.IntegrationProvider
	Function        FunctionProvider
}

// FunctionProvider loads the flows of the app's functions for the Call
// function block.
type FunctionProvider interface {
	FunctionFlow(ctx context.Context, id string) (FlowData, error)
}

type MockFunctionProvider struct {
	Functions map[string]FlowData
}

func (m *MockFunctionProvider) FunctionFlow(ctx context.Context, id string) (FlowData, error) {
	data, ok := m.Functions[id]
	if !ok {
		return FlowData{}, fmt.Errorf("function not found")
	}
	return data, nil
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
