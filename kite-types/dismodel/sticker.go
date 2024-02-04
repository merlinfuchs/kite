package dismodel

type Sticker struct{}

type GuildStickersUpdateEvent struct {
	GuildID  string    `json:"guild_id"`
	Stickers []Sticker `json:"stickers"`
}

type StickerListCall struct{}

type StickerListResponse = []Sticker

type StickerGetCall struct {
	ID string `json:"id"`
}

type StickerGetResponse = Sticker

type StickerCreateCall struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Tags        string `json:"tags"`
	File        string `json:"file"`
}

type StickerCreateResponse = Sticker

type StickerUpdateCall struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Tags        string `json:"tags"`
}

type StickerUpdateResponse = Sticker

type StickerDeleteCall struct {
	ID string `json:"id"`
}

type StickerDeleteResponse struct{}
