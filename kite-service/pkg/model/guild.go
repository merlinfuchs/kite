package model

import (
	"time"

	"github.com/merlinfuchs/dismod/distype"
)

type Guild struct {
	ID          distype.Snowflake `json:"id"`
	Name        string            `json:"name"`
	Icon        string            `json:"icon"`
	Description string            `json:"description"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}
