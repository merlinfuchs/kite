package knowledge

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Plan struct {
	Title                      string  `json:"title"`
	Description                string  `json:"description"`
	Price                      float64 `json:"price"`
	Hidden                     bool    `json:"hidden"`
	MaxCollaborators           int     `json:"feature_max_collaborators"`
	UsageCreditsPerMonth       int     `json:"feature_usage_credits_per_month"`
	MaxGuilds                  int     `json:"feature_max_guilds"`
	MaxCommands                int     `json:"feature_max_commands"`
	MaxVariables               int     `json:"feature_max_variables"`
	MaxMessages                int     `json:"feature_max_messages"`
	MaxEventListeners          int     `json:"feature_max_event_listeners"`
	MaxScheduledEventListeners int     `json:"feature_max_scheduled_event_listeners"`
	MinScheduleIntervalSeconds int     `json:"feature_min_schedule_interval_seconds"`
	MaxAIPromptsPerMonth       int     `json:"feature_max_ai_prompts_per_month"`
	PrioritySupport            bool    `json:"feature_priority_support"`
	RotatingStatus             bool    `json:"feature_rotating_status"`
}

// FetchPlans gets the plans from the Kite API, the same ones the Premium page
// of the dashboard shows.
func FetchPlans(ctx context.Context, url string) ([]Plan, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch plans: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch plans: status %d", resp.StatusCode)
	}

	var body struct {
		Success bool   `json:"success"`
		Data    []Plan `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode plans: %w", err)
	}
	if !body.Success {
		return nil, fmt.Errorf("fetch plans: request failed")
	}

	plans := body.Data[:0]
	for _, p := range body.Data {
		if !p.Hidden {
			plans = append(plans, p)
		}
	}
	return plans, nil
}

func FormatPlans(plans []Plan) string {
	var b strings.Builder
	b.WriteString("=== Plans\nThe current plans and their limits per app, as shown on the Premium page of each app.\n")
	for _, p := range plans {
		price := "free"
		if p.Price > 0 {
			price = fmt.Sprintf("$%.2f per month", p.Price)
		}
		fmt.Fprintf(&b, "\n%s (%s): %s\n", p.Title, price, p.Description)
		fmt.Fprintf(&b, "- Credits per month: %d\n", p.UsageCreditsPerMonth)
		fmt.Fprintf(&b, "- Commands: %s, stored variables: %s, message templates: %s, event listeners: %s\n", limit(p.MaxCommands), limit(p.MaxVariables), limit(p.MaxMessages), limit(p.MaxEventListeners))
		fmt.Fprintf(&b, "- Scheduled event listeners: %s, running at most every %s\n", limit(p.MaxScheduledEventListeners), formatInterval(p.MinScheduleIntervalSeconds))
		fmt.Fprintf(&b, "- Servers: %d\n", p.MaxGuilds)
		fmt.Fprintf(&b, "- Collaborators, including the owner: %s\n", limit(p.MaxCollaborators))
		fmt.Fprintf(&b, "- Flow AI prompts per month: %d\n", p.MaxAIPromptsPerMonth)
		fmt.Fprintf(&b, "- Rotating status and Set status block: %s\n", yesNo(p.RotatingStatus))
		fmt.Fprintf(&b, "- Priority support: %s\n", yesNo(p.PrioritySupport))
	}
	return b.String()
}

// limit formats a limit that Kite doesn't enforce when it's 0.
func limit(n int) string {
	if n == 0 {
		return "unlimited"
	}
	return fmt.Sprint(n)
}

func formatInterval(seconds int) string {
	if seconds <= 0 {
		seconds = 300
	}
	if seconds%60 == 0 {
		if seconds == 60 {
			return "minute"
		}
		return fmt.Sprintf("%d minutes", seconds/60)
	}
	return fmt.Sprintf("%d seconds", seconds)
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
