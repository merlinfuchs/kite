package wire

import (
	"time"

	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"gopkg.in/guregu/null.v4"
)

type Guild struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Icon        null.String `json:"icon"`
	Description null.String `json:"description"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type GuildListResponse APIResponse[[]Guild]

type GuildGetResponse APIResponse[Guild]

func GuildToWire(g *model.Guild) Guild {
	return Guild{
		ID:          g.ID,
		Name:        g.Name,
		Description: null.NewString(g.Description, g.Description != ""),
		Icon:        null.NewString(g.Icon, g.Icon != ""),
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
	}
}
