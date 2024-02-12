package model

import (
	"time"

	"github.com/merlinfuchs/dismod/distype"
)

type User struct {
	ID            distype.Snowflake
	Username      string
	Discriminator string
	GlobalName    string
	Avatar        string
	PublicFlags   int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
