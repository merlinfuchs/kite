package wire

import (
	"time"

	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"gopkg.in/guregu/null.v4"
)

type Guild struct {
	ID          distype.Snowflake `json:"id"`
	Name        string            `json:"name"`
	Icon        null.String       `json:"icon"`
	Description null.String       `json:"description"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`

	UserIsOwner     bool   `json:"user_is_owner,omitempty"`
	UserPermissions string `json:"user_permissions,omitempty"`
	BotPermissions  string `json:"bot_permissions,omitempty"`
}

type GuildListResponse APIResponse[[]Guild]

type GuildGetResponse APIResponse[Guild]

func GuildToWire(g *model.Guild) Guild {
	return Guild{
		ID:          g.ID,
		Name:        g.Name,
		Description: g.Description,
		Icon:        g.Icon,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
	}
}
