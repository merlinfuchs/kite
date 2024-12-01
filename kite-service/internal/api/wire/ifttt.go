package wire

type IFTTTTokenExchangeRequest struct {
	Code         string `json:"code"`
	ClientSecret string `json:"client_secret"`
}

type IFTTTTokenExchangeResponse struct {
	AccessToken string `json:"access_token"`
}

type IFTTTTriggerActionNodeRequest struct {
	TriggerIdentity string                 `json:"trigger_identity"`
	TriggerFields   map[string]interface{} `json:"trigger_fields"`
	Limit           int                    `json:"limit"`
	User            struct{}               `json:"user"`
	IFTTTSource     struct{}               `json:"ifttt_source"`
}

type IFTTTTriggerActionNodeResponse struct {
	Data []struct{} `json:"data"`
}
