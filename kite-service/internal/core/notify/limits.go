package notify

import (
	"fmt"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

// OverLimit tells the owner that the app has more enabled commands or event
// listeners than its plan allows. It's sent even with credit messages turned
// off, because the app changes if they don't act.
func (n *AppOwnerNotifier) OverLimit(app *model.App, details string) {
	n.dm(app, fmt.Sprintf(
		"Your bot **%s** %s See its plan: %s/apps/%s/premium%s",
		escape(app.Name), details, n.appBaseURL, app.ID, n.ownerFooter(),
	))
}

// LimitEnforced tells the owner that commands or event listeners of the app
// were disabled because it was over its limits.
func (n *AppOwnerNotifier) LimitEnforced(app *model.App, details string) {
	n.dm(app, fmt.Sprintf(
		"Your bot **%s** %s See it: %s/apps/%s%s",
		escape(app.Name), details, n.appBaseURL, app.ID, n.ownerFooter(),
	))
}

func (n *AppOwnerNotifier) ownerFooter() string {
	return "\n-# You get this because you own this app on Kite."
}
