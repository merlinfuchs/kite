package model

// Stats are totals over all apps, shown on the website.
type Stats struct {
	Users          int
	Apps           int
	Commands       int
	EventListeners int
	// Flow executions since the time passed to the store.
	Executions int
}
