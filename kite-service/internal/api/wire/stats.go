package wire

type StatsGetResponse struct {
	Users          int `json:"users"`
	Apps           int `json:"apps"`
	Commands       int `json:"commands"`
	EventListeners int `json:"event_listeners"`
	// Flow executions in the last 30 days.
	ExecutionsLastMonth int `json:"executions_last_month"`
	// Apps that added each template, by its ID.
	TemplateInstalls map[string]int `json:"template_installs"`
}
