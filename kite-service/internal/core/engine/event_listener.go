package engine

import (
	"fmt"
	"net/http"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
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

func (c *EventListener) HandleEvent(appID string, session *state.State, event gateway.Event) {
	fmt.Println("TODO: HandleEvent")
}
