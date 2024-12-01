package ifttt

import (
	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
)

type IFTTTHandlerConfig struct {
	ServiceKey   string
	ClientSecret string
}

type IFTTTHandler struct {
	cfg IFTTTHandlerConfig
}

func NewIFTTTHandler(cfg IFTTTHandlerConfig) *IFTTTHandler {
	return &IFTTTHandler{cfg: cfg}
}

func (h *IFTTTHandler) HandleTokenExchange(c *handler.Context, req wire.IFTTTTokenExchangeRequest) (*wire.IFTTTTokenExchangeResponse, error) {
	return nil, nil
}

func (h *IFTTTHandler) HandleTriggerActionNode(c *handler.Context, req wire.IFTTTTriggerActionNodeRequest) (*wire.IFTTTTriggerActionNodeResponse, error) {
	return nil, nil
}

func (h *IFTTTHandler) HandleActionEventTrigger(c *handler.Context, req wire.IFTTTActionEventTriggerRequest) (*wire.IFTTTActionEventTriggerResponse, error) {
	return nil, nil
}

func (h *IFTTTHandler) HandleFieldAppIDOptions(c *handler.Context) (*wire.IFTTTTokenExchangeResponse, error) {
	return nil, nil
}

func (h *IFTTTHandler) HandleFieldEventIDOptions(c *handler.Context) (*wire.IFTTTTokenExchangeResponse, error) {
	return nil, nil
}
