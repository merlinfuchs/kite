package model

import "time"

type User struct {
	ID            string
	Username      string
	Discriminator string
	GlobalName    string
	Avatar        string
	PublicFlags   int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
