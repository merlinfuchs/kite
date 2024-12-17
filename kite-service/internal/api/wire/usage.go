package wire

type UsageCreditsGetResponse struct {
	TotalCredits int `json:"total_credits"`
	CreditsUsed  int `json:"credits_used"`
}
