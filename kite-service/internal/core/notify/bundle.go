package notify

import (
	"fmt"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

// BundleSubmitted tells the admins that the listing of the bundle was
// submitted for the marketplace. The bundle is as it was before.
func (n *Notifier) BundleSubmitted(b *model.Bundle) {
	if n == nil {
		return
	}
	what := fmt.Sprintf("**%s** was submitted for the marketplace", b.Name)
	if b.Listed {
		what = fmt.Sprintf("Changes to the listing of **%s** were submitted", b.Name)
	} else if b.ApprovedVersion > 0 {
		what = fmt.Sprintf("**%s** was submitted for the marketplace again", b.Name)
	}
	n.Review(fmt.Sprintf("%s. Review it: %s/admin/reviews", what, n.appBaseURL))
}

// BundleVersionInReview tells the admins that a new version of the bundle
// waits for review.
func (n *Notifier) BundleVersionInReview(b *model.Bundle) {
	if n == nil {
		return
	}
	n.Review(fmt.Sprintf("Version %d of **%s** waits for review: %s/admin/reviews", b.Version, b.Name, n.appBaseURL))
}

// BundleReviewDecided tells the owner of the bundle whether the review
// approved it.
func (n *Notifier) BundleReviewDecided(ownerDiscordID string, b *model.Bundle, review *model.BundleReview) {
	if n == nil {
		return
	}
	if review.Status == model.BundleReviewStatusApproved {
		n.DM(ownerDiscordID, fmt.Sprintf(
			"Kite approved version %d of your template **%s**. It's listed in the marketplace: %s/marketplace/%s",
			review.Version, b.Name, n.appBaseURL, b.ID,
		))
		return
	}
	n.DM(ownerDiscordID, fmt.Sprintf(
		"Kite didn't approve version %d of your template **%s** for the marketplace: %s\n\nChange it and submit it again, or publish an update with the fix.%s",
		review.Version, b.Name, review.Reason, n.bundlePageLink(b),
	))
}

// BundleRemoved tells the owner of the bundle that Kite took it down.
func (n *Notifier) BundleRemoved(ownerDiscordID string, b *model.Bundle, reason string) {
	if n == nil {
		return
	}
	n.DM(ownerDiscordID, fmt.Sprintf(
		"Kite took your template **%s** down for breaking the rules: %s\n\nOnly you can see it now. To appeal, ask on our Discord server.%s",
		b.Name, reason, n.bundlePageLink(b),
	))
}

// bundlePageLink links to the page of the bundle in the app it's made from,
// if that still exists.
func (n *Notifier) bundlePageLink(b *model.Bundle) string {
	if !b.SourceAppID.Valid {
		return ""
	}
	return fmt.Sprintf(" %s/apps/%s/bundles/%s", n.appBaseURL, b.SourceAppID.String, b.ID)
}
