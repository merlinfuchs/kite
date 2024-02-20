package model

import (
	"time"

	"github.com/merlinfuchs/dismod/distype"
	"gopkg.in/guregu/null.v4"
)

type User struct {
	ID            distype.Snowflake
	Username      string
	Discriminator null.String
	GlobalName    null.String
	Avatar        null.String
	PublicFlags   int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
