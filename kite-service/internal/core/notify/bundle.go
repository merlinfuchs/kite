package notify

import (
	"fmt"
	"strings"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

// publishingGuideURL explains how to get a template approved, which
// rejections link to.
const publishingGuideURL = "https://docs.kite.onl/guides/publishing-templates"

// BundleSubmitted tells the admins that the page of the bundle was
// submitted for the marketplace. The bundle is as it was before.
func (n *Notifier) BundleSubmitted(b *model.Bundle) {
	if n == nil {
		return
	}
	what := fmt.Sprintf("**%s** was submitted for the marketplace", escape(b.Name))
	if b.Listed {
		what = fmt.Sprintf("Changes to the page of **%s** were submitted", escape(b.Name))
	} else if b.ApprovedVersion > 0 {
		what = fmt.Sprintf("**%s** was submitted for the marketplace again", escape(b.Name))
	}
	n.Review(fmt.Sprintf("%s. Review it: %s/admin/reviews", what, n.appBaseURL))
}

// BundleVersionInReview tells the admins that a new version of the bundle
// waits for review.
func (n *Notifier) BundleVersionInReview(b *model.Bundle) {
	if n == nil {
		return
	}
	n.Review(fmt.Sprintf("Version %d of **%s** waits for review: %s/admin/reviews", b.Version, escape(b.Name), n.appBaseURL))
}

// BundleReviewDecided tells the owner of the bundle whether the review
// approved it. The bundle has its owner.
func (n *Notifier) BundleReviewDecided(b *model.Bundle, review *model.BundleReview) {
	if n == nil {
		return
	}
	if review.Status == model.BundleReviewStatusApproved {
		n.DM(b.Owner.DiscordID, fmt.Sprintf(
			"Kite approved version %d of your template **%s**. It's listed in the marketplace: %s/marketplace/%s",
			review.Version, escape(b.Name), n.appBaseURL, b.ID,
		))
		return
	}
	n.DM(b.Owner.DiscordID, fmt.Sprintf(
		"Kite didn't approve version %d of your template **%s** for the marketplace: %s\n\nChange it and submit it again, or publish an update with the fix. How to get a template approved: <%s>%s",
		review.Version, escape(b.Name), review.Reason, publishingGuideURL, n.bundlePageLink(b),
	))
}

// BundleAutoRejected tells the owner of the bundle that the pre-screen
// rejected it, and the admins why. The bundle has its owner.
func (n *Notifier) BundleAutoRejected(b *model.Bundle, review *model.BundleReview, reason string) {
	if n == nil {
		return
	}
	n.DM(b.Owner.DiscordID, fmt.Sprintf(
		"Kite's automatic check didn't approve version %d of your template **%s** for the marketplace: %s\n\nChange it and submit it again, or publish an update with the fix. How to get a template approved: <%s>\nIf you think the check got it wrong, ask on our Discord server.%s",
		review.Version, escape(b.Name), review.Reason, publishingGuideURL, n.bundlePageLink(b),
	))
	n.Review(fmt.Sprintf(
		"The pre-screen rejected version %d of **%s** (%s): %s\nReopen it if that's wrong: %s/admin/reviews",
		review.Version, escape(b.Name), reason, review.Reason, n.appBaseURL,
	))
}

// BundleRemoved tells the owner of the bundle that Kite took it down. The
// bundle has its owner.
func (n *Notifier) BundleRemoved(b *model.Bundle, reason string) {
	if n == nil {
		return
	}
	n.DM(b.Owner.DiscordID, fmt.Sprintf(
		"Kite took your template **%s** down for breaking the rules: %s\n\nOnly you can see it now. To appeal, ask on our Discord server.%s",
		escape(b.Name), reason, n.bundlePageLink(b),
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

// markdownEscaper escapes the Discord markdown in names users chose, so they
// can't format the message or hide links in it.
var markdownEscaper = strings.NewReplacer(
	`\`, `\\`, "*", `\*`, "_", `\_`, "~", `\~`, "`", "\\`", "|", `\|`,
	">", `\>`, "#", `\#`, "[", `\[`, "]", `\]`, "(", `\(`, ")", `\)`,
)

func escape(s string) string {
	return markdownEscaper.Replace(s)
}

// UserBanned tells the user that Kite banned them from the marketplace, the
// only feature they can be banned from for now.
func (n *Notifier) UserBanned(user *model.User, ban *model.UserBan) {
	if n == nil {
		return
	}
	n.DM(user.DiscordID, fmt.Sprintf(
		"Kite banned you from submitting templates to the marketplace: %s\n\nYour listed templates stay listed, but new versions of them aren't reviewed, so apps keep the version they have. To appeal, ask on our Discord server.",
		ban.Reason,
	))
}
