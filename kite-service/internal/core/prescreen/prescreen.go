// Package prescreen has an AI look at templates submitted to the marketplace
// before an admin does. It rejects the ones that are obviously not meant for
// it, and gives the admins a summary, concerns and a rating for the others.
// It never approves. Once a template is listed, it checks whether it may be
// featured.
package prescreen

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/core/bundle"
	"github.com/kitecloud/kite/kite-service/internal/core/notify"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	"github.com/openai/openai-go/v2/responses"
	"github.com/openai/openai-go/v2/shared"
	"gopkg.in/guregu/null.v4"
)

//go:embed instructions.md
var instructionsTemplate string

//go:embed featured.md
var featuredTemplate string

// maxContentSize is the most content it screens, without positions, about 30k
// tokens or a third of a cent with gpt-6-luna. The biggest template in the
// marketplace has half of it. Bigger ones go to the admins unscreened instead
// of being cut, so nothing in them goes unread.
const maxContentSize = 100_000

// pollInterval is how often it looks for reviews to screen.
const pollInterval = 30 * time.Second

const callTimeout = 3 * time.Minute

// The reasons it may reject for. Anything else goes to the admins.
var rejectReasons = []string{"test", "empty", "server_specific", "harmful"}

type Prescreener struct {
	client       *openai.Client
	model        string
	bundleStore  store.BundleStore
	notifier     *notify.Notifier
	instructions string
	featured     string
}

func New(client *openai.Client, model string, bundleStore store.BundleStore, notifier *notify.Notifier) *Prescreener {
	names := []string{}
	for _, b := range bundle.Official() {
		names = append(names, b.Name)
	}
	official := strings.Join(names, ", ")
	return &Prescreener{
		client:       client,
		model:        model,
		bundleStore:  bundleStore,
		notifier:     notifier,
		instructions: strings.ReplaceAll(instructionsTemplate, "{{official}}", official),
		featured:     strings.ReplaceAll(featuredTemplate, "{{official}}", official),
	}
}

// Run screens the pending reviews that weren't screened yet, and checks
// whether the listed templates that weren't checked yet may be featured,
// until ctx is done. Reviews are screened again when their version or page
// changes, and templates are checked again when a review of them is approved.
func (p *Prescreener) Run(ctx context.Context) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		ids, err := p.bundleStore.UnscreenedBundleReviewIDs(ctx, 10)
		if err != nil {
			slog.Error("Failed to get unscreened bundle reviews", slog.String("error", err.Error()))
		}
		for _, id := range ids {
			if err := p.screen(ctx, id); err != nil {
				slog.Error("Failed to pre-screen bundle review", slog.String("review_id", id), slog.String("error", err.Error()))
			}
		}
		p.checkFeatured(ctx)

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (p *Prescreener) screen(ctx context.Context, reviewID string) error {
	r, err := p.bundleStore.PendingBundleReview(ctx, reviewID)
	if err != nil {
		// Decided or changed since.
		if errors.Is(err, store.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("failed to get bundle review: %w", err)
	}

	result := model.BundleReviewPrescreen{
		Model:      p.model,
		ScreenedAt: null.TimeFrom(time.Now().UTC()),
	}
	var out *output
	input, err := prompt(r)
	switch {
	case errors.Is(err, errTooBig):
		result.Verdict = model.BundleReviewPrescreenVerdictSkipped
		result.Summary = "Too big for the pre-screen, review it yourself."
	case err != nil:
		// Like content that isn't JSON. Trying again wouldn't help.
		slog.Warn("Bundle review can't be pre-screened", slog.String("review_id", reviewID), slog.String("error", err.Error()))
		result.Verdict = model.BundleReviewPrescreenVerdictFailed
		result.Summary = "The pre-screen couldn't read it, review it yourself."
	default:
		out, err = p.ask(ctx, input)
		if err != nil {
			slog.Warn("Bundle review pre-screen failed", slog.String("review_id", reviewID), slog.String("error", err.Error()))
			result.Verdict = model.BundleReviewPrescreenVerdictFailed
			result.Summary = "The pre-screen failed, run it again or review it yourself."
			break
		}
		result.Verdict = model.BundleReviewPrescreenVerdictReview
		result.Rating = null.IntFrom(int64(out.Rating))
		result.Summary = out.Summary
		result.Concerns = out.Concerns
		// An admin reopened it after a rejection, which running it again
		// doesn't undo.
		reopened := r.Review.Prescreen.Verdict == model.BundleReviewPrescreenVerdictReject
		if out.rejects() && !reopened {
			result.Verdict = model.BundleReviewPrescreenVerdictReject
		}
	}

	if result.Verdict == model.BundleReviewPrescreenVerdictReject {
		return p.reject(ctx, r, result, out)
	}
	if err := p.bundleStore.SetBundleReviewPrescreen(ctx, reviewID, result); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("failed to store pre-screen: %w", err)
	}
	return nil
}

// reject rejects the review with the pre-screen and without a reviewer,
// which marks it as the pre-screen's.
func (p *Prescreener) reject(ctx context.Context, r *model.PendingBundleReview, result model.BundleReviewPrescreen, out *output) error {
	decided, err := p.bundleStore.DecideBundleReview(ctx, &model.BundleReview{
		ID:        r.Review.ID,
		Status:    model.BundleReviewStatusRejected,
		Reason:    out.CreatorMessage,
		DecidedAt: null.TimeFrom(time.Now().UTC()),
		Prescreen: result,
	})
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("failed to reject bundle review: %w", err)
	}
	p.notifier.BundleAutoRejected(r.Bundle, decided, out.Reason)
	return nil
}

var errTooBig = errors.New("bundle content too big to pre-screen")

// prompt is the page and content of the version in review, without the
// positions of blocks.
func prompt(r *model.PendingBundleReview) (string, error) {
	pageAndContent, err := pageAndContent(r.Review.Listing, r.Bundle.Content)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	if r.Bundle.Listed {
		fmt.Fprintf(&b, "This is version %d of a template already in the marketplace.\n", r.Review.Version)
		for _, n := range r.Bundle.VersionNotes {
			fmt.Fprintf(&b, "What changed, by the creator: %s\n", n.Notes)
		}
	}
	b.WriteString(pageAndContent)
	return b.String(), nil
}

func pageAndContent(l model.BundleListing, rawContent json.RawMessage) (string, error) {
	var content any
	if err := json.Unmarshal(rawContent, &content); err != nil {
		return "", fmt.Errorf("failed to decode bundle content: %w", err)
	}
	contentJSON, err := json.Marshal(withoutLayout(content, false))
	if err != nil {
		return "", err
	}
	if len(contentJSON) > maxContentSize {
		return "", errTooBig
	}

	page, err := json.Marshal(map[string]any{
		"name":         l.Name,
		"description":  l.Description,
		"how_it_works": l.About,
		"setup":        l.Setup,
		"category":     l.Category,
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Page:\n%s\n\nContent:\n%s", page, contentJSON), nil
}

// withoutLayout drops where blocks are and how big, which says nothing about
// what a template does and is a good part of its size. Only blocks, the items
// of nodes, lose them: settings like the position of a channel stay.
func withoutLayout(v any, block bool) any {
	switch v := v.(type) {
	case map[string]any:
		res := make(map[string]any, len(v))
		for k, child := range v {
			if block {
				switch k {
				case "position", "positionAbsolute", "measured", "width", "height", "selected", "dragging":
					continue
				}
			}
			res[k] = withoutLayout(child, false)
			if items, ok := child.([]any); ok && k == "nodes" {
				blocks := make([]any, len(items))
				for i, item := range items {
					blocks[i] = withoutLayout(item, true)
				}
				res[k] = blocks
			}
		}
		return res
	case []any:
		res := make([]any, len(v))
		for i, child := range v {
			res[i] = withoutLayout(child, false)
		}
		return res
	}
	return v
}

type output struct {
	Rating         int                         `json:"rating"`
	Summary        string                      `json:"summary"`
	Concerns       []model.BundleReviewConcern `json:"concerns"`
	Reason         string                      `json:"reject_reason"`
	CreatorMessage string                      `json:"creator_message"`
}

// rejects is whether the model rejected it for one of the reasons it may,
// with a message for the creator.
func (o *output) rejects() bool {
	return slices.Contains(rejectReasons, o.Reason) && strings.TrimSpace(o.CreatorMessage) != ""
}

func (p *Prescreener) ask(ctx context.Context, input string) (*output, error) {
	var out output
	if err := p.call(ctx, "prescreen", p.instructions, input, outputSchema, &out); err != nil {
		return nil, err
	}
	out.Rating = min(max(out.Rating, 0), 100)
	return &out, nil
}

// call has the model answer the input with JSON of the schema, decoded into
// out.
func (p *Prescreener) call(ctx context.Context, name string, instructions string, input string, schema map[string]any, out any) error {
	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()

	resp, err := p.client.Responses.New(ctx, responses.ResponseNewParams{
		Model:           p.model,
		Instructions:    openai.String(instructions),
		Input:           responses.ResponseNewParamsInputUnion{OfString: openai.String(input)},
		MaxOutputTokens: openai.Int(16000),
		Reasoning: shared.ReasoningParam{
			Effort: shared.ReasoningEffortLow,
		},
		Text: responses.ResponseTextConfigParam{
			Format: responses.ResponseFormatTextConfigUnionParam{
				OfJSONSchema: &responses.ResponseFormatTextJSONSchemaConfigParam{
					Name:   name,
					Schema: schema,
					Strict: openai.Bool(true),
				},
			},
		},
		PromptCacheKey: openai.String("kite-" + name),
		Store:          openai.Bool(false),
	}, option.WithMaxRetries(1))
	if err != nil {
		return fmt.Errorf("failed to create response: %w", err)
	}
	slog.Info(
		"Bundle review pre-screen",
		slog.String("check", name),
		slog.String("status", string(resp.Status)),
		slog.Int("input_tokens", int(resp.Usage.InputTokens)),
		slog.Int("cached_input_tokens", int(resp.Usage.InputTokensDetails.CachedTokens)),
		slog.Int("output_tokens", int(resp.Usage.OutputTokens)),
	)
	if resp.Status != responses.ResponseStatusCompleted {
		return fmt.Errorf("response %s: %s", resp.Status, resp.IncompleteDetails.Reason)
	}

	if err := json.Unmarshal([]byte(resp.OutputText()), out); err != nil {
		return fmt.Errorf("failed to decode output: %w", err)
	}
	return nil
}

var outputSchema = map[string]any{
	"type":                 "object",
	"additionalProperties": false,
	"required":             []string{"summary", "concerns", "rating", "reject_reason", "creator_message"},
	"properties": map[string]any{
		"summary": map[string]any{
			"type":        "string",
			"description": "What the template does, in one to three sentences, for the reviewer.",
		},
		"concerns": map[string]any{
			"type":        "array",
			"description": "What the reviewer should look at, most important first. Empty if nothing.",
			"items": map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"required":             []string{"severity", "text"},
				"properties": map[string]any{
					"severity": map[string]any{"type": "string", "enum": []string{"high", "medium", "low"}},
					"text":     map[string]any{"type": "string"},
				},
			},
		},
		"rating": map[string]any{
			"type":        "integer",
			"description": "How likely the reviewer approves it, from 0 to 100.",
		},
		"reject_reason": map[string]any{
			"type":        "string",
			"description": "The reason it's obviously not meant for the marketplace, or none to leave it to the reviewer.",
			"enum":        append([]string{"none"}, rejectReasons...),
		},
		"creator_message": map[string]any{
			"type":        "string",
			"description": "When rejecting: what's wrong and what to change, for the creator. Empty otherwise.",
		},
	},
}
