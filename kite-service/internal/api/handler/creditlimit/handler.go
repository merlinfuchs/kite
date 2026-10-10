package creditlimit

import (
	"errors"
	"fmt"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"gopkg.in/guregu/null.v4"
)

// MaxCreditLimitsPerApp is how many limits an app can have. Each one with a
// target costs a query when the list is loaded.
const MaxCreditLimitsPerApp = 50

const usageListLimit = 25

type CreditLimitHandler struct {
	creditLimitStore store.CreditLimitStore
	usageStore       store.UsageStore
}

func NewCreditLimitHandler(creditLimitStore store.CreditLimitStore, usageStore store.UsageStore) *CreditLimitHandler {
	return &CreditLimitHandler{
		creditLimitStore: creditLimitStore,
		usageStore:       usageStore,
	}
}

func (h *CreditLimitHandler) HandleCreditLimitList(c *handler.Context) (*wire.CreditLimitListResponse, error) {
	limits, err := h.creditLimitStore.CreditLimitsByApp(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get credit limits: %w", err)
	}

	now := time.Now().UTC()

	res := make([]*wire.CreditLimit, len(limits))
	for i, limit := range limits {
		used, err := h.creditsUsed(c, limit, now)
		if err != nil {
			return nil, err
		}
		res[i] = wire.CreditLimitToWire(limit, used)
	}
	return &res, nil
}

func (h *CreditLimitHandler) HandleCreditLimitCreate(c *handler.Context, req wire.CreditLimitCreateRequest) (*wire.CreditLimitCreateResponse, error) {
	count, err := h.creditLimitStore.CountCreditLimitsByApp(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to count credit limits: %w", err)
	}
	if count >= MaxCreditLimitsPerApp {
		return nil, handler.ErrBadRequest("resource_limit", fmt.Sprintf("maximum number of credit limits (%d) reached", MaxCreditLimitsPerApp))
	}

	now := time.Now().UTC()

	limit, err := h.creditLimitStore.CreateCreditLimit(c.Context(), &model.CreditLimit{
		ID:        util.UniqueID(),
		AppID:     c.App.ID,
		Scope:     model.CreditLimitScope(req.Scope),
		TargetID:  req.TargetID,
		Period:    model.CreditLimitPeriod(req.Period),
		Credits:   req.Credits,
		Message:   req.Message,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		if errors.Is(err, store.ErrAlreadyExists) {
			return nil, duplicateError()
		}
		return nil, fmt.Errorf("failed to create credit limit: %w", err)
	}

	used, err := h.creditsUsed(c, limit, now)
	if err != nil {
		return nil, err
	}
	return wire.CreditLimitToWire(limit, used), nil
}

func (h *CreditLimitHandler) HandleCreditLimitUpdate(c *handler.Context, req wire.CreditLimitUpdateRequest) (*wire.CreditLimitUpdateResponse, error) {
	limit, err := h.creditLimitStore.CreditLimit(c.Context(), c.App.ID, c.Param("limitID"))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_credit_limit", "Credit limit not found")
		}
		return nil, fmt.Errorf("failed to get credit limit: %w", err)
	}

	now := time.Now().UTC()

	limit.Scope = model.CreditLimitScope(req.Scope)
	limit.TargetID = req.TargetID
	limit.Period = model.CreditLimitPeriod(req.Period)
	limit.Credits = req.Credits
	limit.Message = req.Message
	limit.UpdatedAt = now

	limit, err = h.creditLimitStore.UpdateCreditLimit(c.Context(), limit)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_credit_limit", "Credit limit not found")
		}
		if errors.Is(err, store.ErrAlreadyExists) {
			return nil, duplicateError()
		}
		return nil, fmt.Errorf("failed to update credit limit: %w", err)
	}

	used, err := h.creditsUsed(c, limit, now)
	if err != nil {
		return nil, err
	}
	return wire.CreditLimitToWire(limit, used), nil
}

func (h *CreditLimitHandler) HandleCreditLimitDelete(c *handler.Context) (*wire.CreditLimitDeleteResponse, error) {
	err := h.creditLimitStore.DeleteCreditLimit(c.Context(), c.App.ID, c.Param("limitID"))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_credit_limit", "Credit limit not found")
		}
		return nil, fmt.Errorf("failed to delete credit limit: %w", err)
	}

	return &wire.CreditLimitDeleteResponse{}, nil
}

// HandleCreditLimitUsageList lists the servers or users that used the most
// credits in the current day or month, with the limit that applies to each, so
// owners can see who to limit.
func (h *CreditLimitHandler) HandleCreditLimitUsageList(c *handler.Context) (*wire.CreditLimitUsageListResponse, error) {
	scope := model.CreditLimitScope(c.Query("scope"))
	if scope != model.CreditLimitScopeGuild && scope != model.CreditLimitScopeUser {
		return nil, handler.ErrBadRequest("invalid_scope", "scope must be guild or user")
	}

	period := model.CreditLimitPeriod(c.Query("period"))
	if period == "" {
		period = model.CreditLimitPeriodMonth
	}
	if period != model.CreditLimitPeriodDay && period != model.CreditLimitPeriodMonth {
		return nil, handler.ErrBadRequest("invalid_period", "period must be day or month")
	}

	limits, err := h.creditLimitStore.CreditLimitsByApp(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get credit limits: %w", err)
	}

	now := time.Now().UTC()

	entries, err := h.usageStore.TopUsageCreditsByTargetBetween(c.Context(), c.App.ID, scope, period.Start(now), now, usageListLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage by %s: %w", scope, err)
	}

	res := make([]*wire.CreditLimitUsageEntry, len(entries))
	for i, entry := range entries {
		var credits null.Int
		for _, limit := range model.EffectiveCreditLimits(limits, scope, entry.TargetID) {
			if limit.Period == period {
				credits = limit.Credits
			}
		}

		res[i] = &wire.CreditLimitUsageEntry{
			TargetID:    entry.TargetID,
			CreditsUsed: entry.CreditsUsed,
			Credits:     credits,
		}
	}
	return &res, nil
}

func (h *CreditLimitHandler) HandleCreditLimitSettingsGet(c *handler.Context) (*wire.CreditLimitSettingsGetResponse, error) {
	settings, err := h.creditLimitStore.CreditLimitSettings(c.Context(), c.App.ID)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return nil, fmt.Errorf("failed to get credit limit settings: %w", err)
	}

	return wire.CreditLimitSettingsToWire(settings), nil
}

func (h *CreditLimitHandler) HandleCreditLimitSettingsUpdate(c *handler.Context, req wire.CreditLimitSettingsUpdateRequest) (*wire.CreditLimitSettingsUpdateResponse, error) {
	now := time.Now().UTC()

	settings, err := h.creditLimitStore.UpsertCreditLimitSettings(c.Context(), &model.CreditLimitSettings{
		AppID:     c.App.ID,
		Message:   req.Message,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update credit limit settings: %w", err)
	}

	return wire.CreditLimitSettingsToWire(settings), nil
}

func (h *CreditLimitHandler) creditsUsed(c *handler.Context, limit *model.CreditLimit, now time.Time) (null.Int, error) {
	if !limit.TargetID.Valid {
		return null.Int{}, nil
	}

	used, err := h.usageStore.UsageCreditsUsedByTargetSince(c.Context(), c.App.ID, limit.Scope, limit.TargetID.String, limit.Period.Start(now))
	if err != nil {
		return null.Int{}, fmt.Errorf("failed to get credits used: %w", err)
	}
	return null.IntFrom(int64(used)), nil
}

func duplicateError() error {
	return handler.ErrBadRequest("duplicate_credit_limit", "There already is a limit for this target and period")
}
