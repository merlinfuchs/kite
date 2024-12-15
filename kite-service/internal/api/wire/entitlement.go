package wire

type EntitlementsCreditsGetResponse struct {
	TotalCredits int `json:"total_credits"`
	CreditsUsed  int `json:"credits_used"`
}
