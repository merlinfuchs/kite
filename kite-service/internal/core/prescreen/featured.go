package prescreen

import (
	"context"
	"errors"
	"log/slog"
)

type featuredOutput struct {
	Featurable bool   `json:"featurable"`
	Reason     string `json:"reason"`
}

var featuredSchema = map[string]any{
	"type":                 "object",
	"additionalProperties": false,
	"required":             []string{"reason", "featurable"},
	"properties": map[string]any{
		"reason": map[string]any{
			"type":        "string",
			"description": "What made you decide, in one short sentence, for the Kite staff.",
		},
		"featurable": map[string]any{
			"type":        "boolean",
			"description": "Whether it may be shown in the Featured row.",
		},
	},
}

// checkFeatured checks whether the listed templates that weren't checked yet
// may be featured. One it can't read or is too big to check isn't featured.
// When the model fails, it tries again on the next round.
func (p *Prescreener) checkFeatured(ctx context.Context) {
	bundles, err := p.bundleStore.UncheckedFeaturableBundles(ctx, 10)
	if err != nil {
		slog.Error("Failed to get unchecked featurable bundles", slog.String("error", err.Error()))
		return
	}
	for _, b := range bundles {
		var out featuredOutput
		input, err := pageAndContent(*b.Listing, b.Content)
		switch {
		case errors.Is(err, errTooBig):
			out.Reason = "Too big to check."
		case err != nil:
			slog.Warn("Bundle can't be checked for Featured", slog.String("bundle_id", b.ID), slog.String("error", err.Error()))
			out.Reason = "Couldn't read it."
		default:
			if err := p.call(ctx, "featured", p.featured, input, featuredSchema, &out); err != nil {
				slog.Warn("Bundle Featured check failed", slog.String("bundle_id", b.ID), slog.String("error", err.Error()))
				return
			}
		}

		slog.Info(
			"Bundle Featured check",
			slog.String("bundle_id", b.ID),
			slog.Bool("featurable", out.Featurable),
			slog.String("reason", out.Reason),
		)
		if err := p.bundleStore.SetBundleFeaturable(ctx, b.ID, b.ApprovedVersion, out.Featurable); err != nil {
			slog.Error("Failed to store Featured check", slog.String("bundle_id", b.ID), slog.String("error", err.Error()))
		}
	}
}
