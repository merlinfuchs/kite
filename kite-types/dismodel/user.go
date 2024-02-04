package dismodel

import (
	"fmt"
	"strconv"
)

type User struct {
	ID               string `json:"id"`
	Username         string `json:"username"`
	Discriminator    string `json:"discriminator"`
	GlobalName       string `json:"global_name,omitempty"`
	Avatar           string `json:"avatar,omitempty"`
	Bot              bool   `json:"bot,omitempty"`
	System           bool   `json:"system,omitempty"`
	Banner           string `json:"banner,omitempty"`
	AccentColor      int    `json:"accent_color,omitempty"`
	PublicFlags      int    `json:"public_flags,omitempty"`
	AvatarDecoration string `json:"avatar_decoration,omitempty"`
}

func (u User) AvatarURL() string {
	if u.Avatar != "" {
		return fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", u.ID, u.Avatar)
	}

	if u.Discriminator == "0" {
		dis, _ := strconv.ParseUint(u.Discriminator, 10, 8)
		return fmt.Sprintf("https://cdn.discordapp.com/embed/avatars/%d.png", dis%5)
	}

	id, _ := strconv.ParseUint(u.Discriminator, 10, 64)
	return fmt.Sprintf("https://cdn.discordapp.com/embed/avatars/%d.png", (id>>22)%6)
}

type UserGetCall struct {
	UserID string `json:"user_id"`
}

type UserGetResponse = User
