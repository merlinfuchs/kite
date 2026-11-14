package wire

import "gopkg.in/guregu/null.v4"

// LookupUserGetResponse is the public profile of a Discord user, for the user
// lookup tool on the website.
type LookupUserGetResponse struct {
	ID            string      `json:"id"`
	Username      string      `json:"username"`
	GlobalName    null.String `json:"global_name"`
	Discriminator string      `json:"discriminator"`
	Avatar        null.String `json:"avatar"`
	Banner        null.String `json:"banner"`
	AccentColor   null.Int    `json:"accent_color"`
	Bot           bool        `json:"bot"`
	System        bool        `json:"system"`
	PublicFlags   int         `json:"public_flags"`
	// AvatarDecoration is the asset of the decoration around the avatar.
	AvatarDecoration null.String `json:"avatar_decoration"`
	// PrimaryGuild is the server tag the user shows next to their name.
	PrimaryGuild *LookupUserPrimaryGuild `json:"primary_guild"`
}

type LookupUserPrimaryGuild struct {
	GuildID string `json:"guild_id"`
	Tag     string `json:"tag"`
	Badge   string `json:"badge"`
}
