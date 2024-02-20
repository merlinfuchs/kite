package model

import (
	"time"

	"github.com/merlinfuchs/dismod/distype"
	"gopkg.in/guregu/null.v4"
)

type Guild struct {
	ID          distype.Snowflake `json:"id"`
	Name        string            `json:"name"`
	Icon        null.String       `json:"icon"`
	Description null.String       `json:"description"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}
