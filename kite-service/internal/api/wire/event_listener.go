package wire

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

type EventListener struct {
	ID            string               `json:"id"`
	Source        string               `json:"source"`
	Type          string               `json:"type"`
	Description   string               `json:"description"`
	Enabled       bool                 `json:"enabled"`
	AppID         string               `json:"app_id"`
	ModuleID      null.String          `json:"module_id"`
	CreatorUserID string               `json:"creator_user_id"`
	Filter        *EventListenerFilter `json:"filter"`
	FlowSource    flow.FlowData        `json:"flow_source"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

type EventListenerFilter struct{}

type EventListenerGetResponse = EventListener

type EventListenerListResponse = []*EventListener

type EventListenerCreateRequest struct {
	Source     string        `json:"source"`
	FlowSource flow.FlowData `json:"flow_source"`
	Enabled    bool          `json:"enabled"`
}

func (req EventListenerCreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Source, validation.Required, validation.In(string(model.EventSourceDiscord))),
		validation.Field(&req.FlowSource, validation.Required),
	)
}

type EventListenerCreateResponse = EventListener

type EventListenerUpdateRequest struct {
	FlowSource flow.FlowData `json:"flow_source"`
	Enabled    bool          `json:"enabled"`
}

func (req EventListenerUpdateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.FlowSource, validation.Required),
	)
}

type EventListenerUpdateResponse = EventListener

type EventListenerDeleteResponse = Empty

func EventListenerToWire(eventListener *model.EventListener) *EventListener {
	if eventListener == nil {
		return nil
	}

	return &EventListener{
		ID:            eventListener.ID,
		Source:        string(eventListener.Source),
		Type:          string(eventListener.Type),
		Description:   eventListener.Description,
		Enabled:       eventListener.Enabled,
		AppID:         eventListener.AppID,
		ModuleID:      eventListener.ModuleID,
		CreatorUserID: eventListener.CreatorUserID,
		Filter:        (*EventListenerFilter)(eventListener.Filter),
		FlowSource:    eventListener.FlowSource,
		CreatedAt:     eventListener.CreatedAt,
		UpdatedAt:     eventListener.UpdatedAt,
	}
}
