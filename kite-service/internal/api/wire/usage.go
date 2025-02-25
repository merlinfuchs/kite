package wire

import "time"

type UsageCreditsGetResponse struct {
	CreditsUsed int `json:"credits_used"`
}

type UsageByDayListResponse []*UsageByDayEntry

type UsageByDayEntry struct {
	Date        time.Time `json:"date"`
	CreditsUsed int       `json:"credits_used"`
}

type UsageByTypeListResponse []*UsageByTypeEntry

type UsageByTypeEntry struct {
	Type        string `json:"type"`
	CreditsUsed int    `json:"credits_used"`
}
