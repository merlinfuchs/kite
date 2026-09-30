package marketplace

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

const (
	defaultPageLimit = 24
	maxPageLimit     = 50
)

type MarketplaceHandlerConfig struct {
	// AdminDiscordIDs can moderate and add or remove other moderators.
	AdminDiscordIDs []string
	// RequireReview keeps new and changed listings hidden until a moderator
	// approves them.
	RequireReview bool
	// MaxListingsPerUser is how many listings one user can publish.
	MaxListingsPerUser int
	// AutoHideReports is how many open reports send an approved listing back
	// to the moderation queue, 0 disables it.
	AutoHideReports int
}

type MarketplaceHandler struct {
	config           MarketplaceHandlerConfig
	marketplaceStore store.MarketplaceStore
	userStore        store.UserStore
}

func NewMarketplaceHandler(
	config MarketplaceHandlerConfig,
	marketplaceStore store.MarketplaceStore,
	userStore store.UserStore,
) *MarketplaceHandler {
	return &MarketplaceHandler{
		config:           config,
		marketplaceStore: marketplaceStore,
		userStore:        userStore,
	}
}

func (h *MarketplaceHandler) HandleMarketplaceMeGet(c *handler.Context) (*wire.MarketplaceMeGetResponse, error) {
	isModerator, isAdmin, err := h.roles(c)
	if err != nil {
		return nil, err
	}

	return &wire.MarketplaceMe{
		IsModerator: isModerator,
		IsAdmin:     isAdmin,
		MaxListings: h.config.MaxListingsPerUser,
	}, nil
}

func (h *MarketplaceHandler) HandleMarketplaceListingList(c *handler.Context) (*wire.MarketplaceListingListResponse, error) {
	filter, err := listingsFilter(c)
	if err != nil {
		return nil, err
	}
	filter.Status = model.MarketplaceListingStatusApproved

	return h.listListings(c, filter)
}

func (h *MarketplaceHandler) HandleMarketplaceListingListMine(c *handler.Context) (*wire.MarketplaceListingListResponse, error) {
	listings, err := h.marketplaceStore.MarketplaceListingsByAuthor(c.Context(), c.Session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get marketplace listings: %w", err)
	}

	res := make([]*wire.MarketplaceListing, len(listings))
	for i, listing := range listings {
		res[i] = wire.MarketplaceListingToWire(&listing.MarketplaceListing, listing.Author, false)
	}

	return &res, nil
}

func (h *MarketplaceHandler) HandleMarketplaceListingGet(c *handler.Context) (*wire.MarketplaceListingGetResponse, error) {
	listing, err := h.visibleListing(c)
	if err != nil {
		return nil, err
	}

	return wire.MarketplaceListingToWire(&listing.MarketplaceListing, listing.Author, true), nil
}

func (h *MarketplaceHandler) HandleMarketplaceListingCreate(c *handler.Context, req wire.MarketplaceListingCreateRequest) (*wire.MarketplaceListingCreateResponse, error) {
	if h.config.MaxListingsPerUser > 0 {
		count, err := h.marketplaceStore.CountMarketplaceListingsByAuthor(c.Context(), c.Session.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to count marketplace listings: %w", err)
		}
		if count >= h.config.MaxListingsPerUser {
			return nil, handler.ErrBadRequest("resource_limit", fmt.Sprintf("maximum number of marketplace listings (%d) reached", h.config.MaxListingsPerUser))
		}
	}

	content, err := compileListingItems(req.Items)
	if err != nil {
		return nil, err
	}

	isModerator, _, err := h.roles(c)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	listing, err := h.marketplaceStore.CreateMarketplaceListing(c.Context(), &model.MarketplaceListing{
		ID:                 util.UniqueID(),
		Name:               strings.TrimSpace(req.Name),
		Description:        strings.TrimSpace(req.Description),
		AuthorUserID:       c.Session.UserID,
		SourceAppID:        req.AppID,
		Status:             h.initialStatus(isModerator),
		Items:              content.items,
		CommandCount:       content.commandCount,
		EventListenerCount: content.eventListenerCount,
		BlockTypes:         content.blockTypes,
		CreatedAt:          now,
		UpdatedAt:          now,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create marketplace listing: %w", err)
	}

	author, err := h.userStore.User(c.Context(), c.Session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return wire.MarketplaceListingToWire(listing, author, true), nil
}

func (h *MarketplaceHandler) HandleMarketplaceListingUpdate(c *handler.Context, req wire.MarketplaceListingUpdateRequest) (*wire.MarketplaceListingUpdateResponse, error) {
	existing, err := h.listing(c)
	if err != nil {
		return nil, err
	}
	if existing.AuthorUserID != c.Session.UserID {
		return nil, handler.ErrForbidden("missing_access", "Only the author can edit a listing")
	}

	content, err := compileListingItems(req.Items)
	if err != nil {
		return nil, err
	}

	isModerator, _, err := h.roles(c)
	if err != nil {
		return nil, err
	}

	listing, err := h.marketplaceStore.UpdateMarketplaceListing(c.Context(), &model.MarketplaceListing{
		ID:                 existing.ID,
		Name:               strings.TrimSpace(req.Name),
		Description:        strings.TrimSpace(req.Description),
		SourceAppID:        req.AppID,
		Status:             h.initialStatus(isModerator),
		Items:              content.items,
		CommandCount:       content.commandCount,
		EventListenerCount: content.eventListenerCount,
		BlockTypes:         content.blockTypes,
		UpdatedAt:          time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update marketplace listing: %w", err)
	}

	return wire.MarketplaceListingToWire(listing, existing.Author, true), nil
}

func (h *MarketplaceHandler) HandleMarketplaceListingDelete(c *handler.Context) (*wire.MarketplaceListingDeleteResponse, error) {
	listing, err := h.listing(c)
	if err != nil {
		return nil, err
	}

	if listing.AuthorUserID != c.Session.UserID {
		isModerator, _, err := h.roles(c)
		if err != nil {
			return nil, err
		}
		if !isModerator {
			return nil, handler.ErrForbidden("missing_access", "Only the author or a moderator can delete a listing")
		}
	}

	if err := h.marketplaceStore.DeleteMarketplaceListing(c.Context(), listing.ID); err != nil {
		return nil, fmt.Errorf("failed to delete marketplace listing: %w", err)
	}

	return &wire.MarketplaceListingDeleteResponse{}, nil
}

// HandleMarketplaceListingImport returns the listing with its flows and counts
// the import. The web app then imports the items with the regular import
// routes, so plan limits and flow validation apply as usual.
func (h *MarketplaceHandler) HandleMarketplaceListingImport(c *handler.Context) (*wire.MarketplaceListingImportResponse, error) {
	listing, err := h.visibleListing(c)
	if err != nil {
		return nil, err
	}

	if listing.Status == model.MarketplaceListingStatusApproved && listing.AuthorUserID != c.Session.UserID {
		if err := h.marketplaceStore.IncrementMarketplaceListingImportCount(c.Context(), listing.ID); err != nil {
			return nil, fmt.Errorf("failed to count marketplace import: %w", err)
		}
		listing.ImportCount++
	}

	return wire.MarketplaceListingToWire(&listing.MarketplaceListing, listing.Author, true), nil
}

func (h *MarketplaceHandler) HandleMarketplaceReportCreate(c *handler.Context, req wire.MarketplaceReportCreateRequest) (*wire.MarketplaceReportCreateResponse, error) {
	listing, err := h.listing(c)
	if err != nil {
		return nil, err
	}
	if listing.Status != model.MarketplaceListingStatusApproved {
		return nil, handler.ErrNotFound("unknown_listing", "Listing not found")
	}
	if listing.AuthorUserID == c.Session.UserID {
		return nil, handler.ErrBadRequest("own_listing", "You can't report your own listing")
	}

	now := time.Now().UTC()
	_, err = h.marketplaceStore.CreateMarketplaceReport(c.Context(), &model.MarketplaceReport{
		ID:             util.UniqueID(),
		ListingID:      listing.ID,
		ReporterUserID: c.Session.UserID,
		Reason:         strings.TrimSpace(req.Reason),
		CreatedAt:      now,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create marketplace report: %w", err)
	}

	if h.config.AutoHideReports > 0 {
		count, err := h.marketplaceStore.CountOpenMarketplaceReportsByListing(c.Context(), listing.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to count marketplace reports: %w", err)
		}

		// Hidden until a moderator looks at it, so a malicious listing can't
		// keep spreading while nobody is around.
		if count >= h.config.AutoHideReports {
			_, err := h.marketplaceStore.ReviewMarketplaceListing(
				c.Context(),
				listing.ID,
				model.MarketplaceListingStatusPending,
				"Hidden automatically after multiple reports, waiting for a moderator.",
				"",
				now,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to hide marketplace listing: %w", err)
			}
		}
	}

	return &wire.MarketplaceReportCreateResponse{}, nil
}

func (h *MarketplaceHandler) HandleModerationListingList(c *handler.Context) (*wire.MarketplaceListingListResponse, error) {
	filter, err := listingsFilter(c)
	if err != nil {
		return nil, err
	}

	status := c.Query("status")
	if status == "" {
		status = string(model.MarketplaceListingStatusPending)
	}
	switch model.MarketplaceListingStatus(status) {
	case model.MarketplaceListingStatusPending,
		model.MarketplaceListingStatusApproved,
		model.MarketplaceListingStatusRejected,
		model.MarketplaceListingStatusRemoved:
	default:
		return nil, handler.ErrBadRequest("invalid_status", "Unknown listing status")
	}
	filter.Status = model.MarketplaceListingStatus(status)

	return h.listListings(c, filter)
}

func (h *MarketplaceHandler) HandleModerationListingReview(c *handler.Context, req wire.MarketplaceListingReviewRequest) (*wire.MarketplaceListingReviewResponse, error) {
	existing, err := h.listing(c)
	if err != nil {
		return nil, err
	}

	status := model.MarketplaceListingStatus(req.Status)
	// Taking down a listing that was never public is a rejection.
	if status == model.MarketplaceListingStatusRemoved && existing.Status != model.MarketplaceListingStatusApproved {
		status = model.MarketplaceListingStatusRejected
	}

	now := time.Now().UTC()
	listing, err := h.marketplaceStore.ReviewMarketplaceListing(
		c.Context(),
		existing.ID,
		status,
		strings.TrimSpace(req.Note),
		c.Session.UserID,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to review marketplace listing: %w", err)
	}

	// The review answers every open report on the listing.
	err = h.marketplaceStore.ResolveMarketplaceReportsByListing(c.Context(), existing.ID, c.Session.UserID, now)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve marketplace reports: %w", err)
	}

	return wire.MarketplaceListingToWire(listing, existing.Author, true), nil
}

func (h *MarketplaceHandler) HandleModerationReportList(c *handler.Context) (*wire.MarketplaceReportListResponse, error) {
	reports, err := h.marketplaceStore.OpenMarketplaceReports(c.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to get marketplace reports: %w", err)
	}

	listings := make(map[string]*wire.MarketplaceListing)
	res := make([]*wire.MarketplaceReport, 0, len(reports))
	for _, report := range reports {
		listing, ok := listings[report.ListingID]
		if !ok {
			l, err := h.marketplaceStore.MarketplaceListing(c.Context(), report.ListingID)
			if err != nil {
				if errors.Is(err, store.ErrNotFound) {
					continue
				}
				return nil, fmt.Errorf("failed to get marketplace listing: %w", err)
			}
			listing = wire.MarketplaceListingToWire(&l.MarketplaceListing, l.Author, false)
			listings[report.ListingID] = listing
		}

		res = append(res, &wire.MarketplaceReport{
			ID:        report.ID,
			ListingID: report.ListingID,
			Listing:   listing,
			Reporter:  wire.MarketplaceUserToWire(report.Reporter),
			Reason:    report.Reason,
			CreatedAt: report.CreatedAt,
		})
	}

	return &res, nil
}

func (h *MarketplaceHandler) HandleModerationReportResolve(c *handler.Context) (*wire.MarketplaceReportResolveResponse, error) {
	err := h.marketplaceStore.ResolveMarketplaceReport(c.Context(), c.Param("reportID"), c.Session.UserID, time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("failed to resolve marketplace report: %w", err)
	}

	return &wire.MarketplaceReportResolveResponse{}, nil
}

func (h *MarketplaceHandler) HandleModeratorList(c *handler.Context) (*wire.MarketplaceModeratorListResponse, error) {
	moderators, err := h.marketplaceStore.MarketplaceModerators(c.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to get marketplace moderators: %w", err)
	}

	res := make([]*wire.MarketplaceModerator, 0, len(h.config.AdminDiscordIDs)+len(moderators))
	for _, discordID := range h.config.AdminDiscordIDs {
		moderator, err := h.moderatorToWire(c, discordID, null.Time{})
		if err != nil {
			return nil, err
		}
		moderator.IsAdmin = true
		res = append(res, moderator)
	}

	for _, m := range moderators {
		if slices.Contains(h.config.AdminDiscordIDs, m.DiscordUserID) {
			continue
		}

		moderator, err := h.moderatorToWire(c, m.DiscordUserID, null.TimeFrom(m.CreatedAt))
		if err != nil {
			return nil, err
		}
		res = append(res, moderator)
	}

	return &res, nil
}

func (h *MarketplaceHandler) HandleModeratorCreate(c *handler.Context, req wire.MarketplaceModeratorCreateRequest) (*wire.MarketplaceModeratorCreateResponse, error) {
	moderator, err := h.marketplaceStore.CreateMarketplaceModerator(c.Context(), &model.MarketplaceModerator{
		DiscordUserID: req.DiscordUserID,
		AddedByUserID: null.StringFrom(c.Session.UserID),
		CreatedAt:     time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create marketplace moderator: %w", err)
	}

	res, err := h.moderatorToWire(c, moderator.DiscordUserID, null.TimeFrom(moderator.CreatedAt))
	if err != nil {
		return nil, err
	}
	res.IsAdmin = slices.Contains(h.config.AdminDiscordIDs, moderator.DiscordUserID)

	return res, nil
}

func (h *MarketplaceHandler) HandleModeratorDelete(c *handler.Context) (*wire.MarketplaceModeratorDeleteResponse, error) {
	discordUserID := c.Param("discordUserID")
	if slices.Contains(h.config.AdminDiscordIDs, discordUserID) {
		return nil, handler.ErrBadRequest("admin_moderator", "Admins are set in the service config and can't be removed here")
	}

	if err := h.marketplaceStore.DeleteMarketplaceModerator(c.Context(), discordUserID); err != nil {
		return nil, fmt.Errorf("failed to delete marketplace moderator: %w", err)
	}

	return &wire.MarketplaceModeratorDeleteResponse{}, nil
}

// RequireModerator only lets marketplace moderators and admins through.
func (h *MarketplaceHandler) RequireModerator(next handler.HandlerFunc) handler.HandlerFunc {
	return func(c *handler.Context) error {
		isModerator, _, err := h.roles(c)
		if err != nil {
			return err
		}
		if !isModerator {
			return handler.ErrForbidden("missing_access", "You are not a marketplace moderator")
		}
		return next(c)
	}
}

// RequireAdmin only lets marketplace admins from the service config through.
func (h *MarketplaceHandler) RequireAdmin(next handler.HandlerFunc) handler.HandlerFunc {
	return func(c *handler.Context) error {
		_, isAdmin, err := h.roles(c)
		if err != nil {
			return err
		}
		if !isAdmin {
			return handler.ErrForbidden("missing_access", "You are not a marketplace admin")
		}
		return next(c)
	}
}

// roles looks up the session user's Discord ID, moderators are stored by it.
func (h *MarketplaceHandler) roles(c *handler.Context) (isModerator bool, isAdmin bool, err error) {
	user, err := h.userStore.User(c.Context(), c.Session.UserID)
	if err != nil {
		return false, false, fmt.Errorf("failed to get user: %w", err)
	}

	if slices.Contains(h.config.AdminDiscordIDs, user.DiscordID) {
		return true, true, nil
	}

	_, err = h.marketplaceStore.MarketplaceModerator(c.Context(), user.DiscordID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return false, false, nil
		}
		return false, false, fmt.Errorf("failed to get marketplace moderator: %w", err)
	}

	return true, false, nil
}

func (h *MarketplaceHandler) initialStatus(isModerator bool) model.MarketplaceListingStatus {
	if !h.config.RequireReview || isModerator {
		return model.MarketplaceListingStatusApproved
	}
	return model.MarketplaceListingStatusPending
}

func (h *MarketplaceHandler) listing(c *handler.Context) (*model.MarketplaceListingWithAuthor, error) {
	listing, err := h.marketplaceStore.MarketplaceListing(c.Context(), c.Param("listingID"))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_listing", "Listing not found")
		}
		return nil, fmt.Errorf("failed to get marketplace listing: %w", err)
	}

	return listing, nil
}

// visibleListing hides listings that aren't approved from everyone but their
// author and moderators.
func (h *MarketplaceHandler) visibleListing(c *handler.Context) (*model.MarketplaceListingWithAuthor, error) {
	listing, err := h.listing(c)
	if err != nil {
		return nil, err
	}

	if listing.Status != model.MarketplaceListingStatusApproved && listing.AuthorUserID != c.Session.UserID {
		isModerator, _, err := h.roles(c)
		if err != nil {
			return nil, err
		}
		if !isModerator {
			return nil, handler.ErrNotFound("unknown_listing", "Listing not found")
		}
	}

	return listing, nil
}

func (h *MarketplaceHandler) listListings(c *handler.Context, filter store.MarketplaceListingsFilter) (*wire.MarketplaceListingListResponse, error) {
	listings, err := h.marketplaceStore.MarketplaceListings(c.Context(), filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get marketplace listings: %w", err)
	}

	res := make([]*wire.MarketplaceListing, len(listings))
	for i, listing := range listings {
		res[i] = wire.MarketplaceListingToWire(&listing.MarketplaceListing, listing.Author, false)
	}

	return &res, nil
}

func (h *MarketplaceHandler) moderatorToWire(c *handler.Context, discordID string, createdAt null.Time) (*wire.MarketplaceModerator, error) {
	user, err := h.userStore.UserByDiscordID(c.Context(), discordID)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &wire.MarketplaceModerator{
		DiscordUserID: discordID,
		User:          wire.MarketplaceUserToWire(user),
		CreatedAt:     createdAt,
	}, nil
}

func listingsFilter(c *handler.Context) (store.MarketplaceListingsFilter, error) {
	filter := store.MarketplaceListingsFilter{
		Search: strings.TrimSpace(c.Query("search")),
		Kind:   c.Query("kind"),
		Sort:   c.Query("sort"),
		Limit:  defaultPageLimit,
	}

	if len(filter.Search) > 100 {
		return filter, handler.ErrBadRequest("invalid_search", "Search is too long")
	}

	switch filter.Kind {
	case "", "command", "event_listener", "module":
	default:
		return filter, handler.ErrBadRequest("invalid_kind", "Unknown listing kind")
	}

	switch filter.Sort {
	case "":
		filter.Sort = "popular"
	case "popular", "recent":
	default:
		return filter, handler.ErrBadRequest("invalid_sort", "Unknown sort")
	}

	if raw := c.Query("limit"); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > maxPageLimit {
			return filter, handler.ErrBadRequest("invalid_limit", fmt.Sprintf("Limit must be between 1 and %d", maxPageLimit))
		}
		filter.Limit = limit
	}

	if raw := c.Query("offset"); raw != "" {
		offset, err := strconv.Atoi(raw)
		if err != nil || offset < 0 {
			return filter, handler.ErrBadRequest("invalid_offset", "Offset must be a positive number")
		}
		filter.Offset = offset
	}

	return filter, nil
}

type listingContent struct {
	items              []model.MarketplaceListingItem
	commandCount       int
	eventListenerCount int
	blockTypes         []string
}

// compileListingItems checks that every item is a valid flow of its type and
// takes the names and descriptions from the flows, so a listing can't claim
// to contain something it doesn't.
func compileListingItems(reqItems []wire.MarketplaceListingItemRequest) (*listingContent, error) {
	res := &listingContent{
		items: make([]model.MarketplaceListingItem, len(reqItems)),
	}

	blockTypes := make(map[string]struct{})
	commandNames := make(map[string]struct{})

	for i, reqItem := range reqItems {
		item := model.MarketplaceListingItem{
			Type:       model.MarketplaceListingItemType(reqItem.Type),
			FlowSource: reqItem.FlowSource,
		}

		switch item.Type {
		case model.MarketplaceListingItemTypeCommand:
			cmdFlow, err := flow.CompileCommand(reqItem.FlowSource)
			if err != nil {
				return nil, handler.ErrBadRequest("invalid_flow", fmt.Sprintf("item %d: %s", i+1, err))
			}

			item.Name = cmdFlow.CommandName()
			item.Description = cmdFlow.CommandDescription()

			if _, ok := commandNames[item.Name]; ok {
				return nil, handler.ErrBadRequest("duplicate_command", fmt.Sprintf("command /%s is in the listing twice", item.Name))
			}
			commandNames[item.Name] = struct{}{}
			res.commandCount++
		case model.MarketplaceListingItemTypeEventListener:
			eventFlow, err := flow.CompileEventListener(reqItem.FlowSource)
			if err != nil {
				return nil, handler.ErrBadRequest("invalid_flow", fmt.Sprintf("item %d: %s", i+1, err))
			}

			source := model.EventSource(reqItem.Source)
			typeSource := model.EventSourceForType(model.EventListenerType(eventFlow.EventListenerType()))
			if typeSource != source {
				return nil, handler.ErrBadRequest(
					"invalid_source",
					fmt.Sprintf("item %d: event type %s belongs to source %s, not %s", i+1, eventFlow.EventListenerType(), typeSource, source),
				)
			}

			item.Name = eventFlow.EventListenerType()
			item.Description = eventFlow.EventDescription()
			item.Source = string(source)
			res.eventListenerCount++
		}

		for _, node := range reqItem.FlowSource.Nodes {
			if node.Type != "" {
				blockTypes[string(node.Type)] = struct{}{}
			}
		}

		res.items[i] = item
	}

	res.blockTypes = make([]string, 0, len(blockTypes))
	for t := range blockTypes {
		res.blockTypes = append(res.blockTypes, t)
	}
	sort.Strings(res.blockTypes)

	return res, nil
}
