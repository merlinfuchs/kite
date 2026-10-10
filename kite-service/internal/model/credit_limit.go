package model

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

type CreditLimitScope string

const (
	CreditLimitScopeGuild CreditLimitScope = "guild"
	CreditLimitScopeUser  CreditLimitScope = "user"
)

type CreditLimitPeriod string

const (
	CreditLimitPeriodDay   CreditLimitPeriod = "day"
	CreditLimitPeriodMonth CreditLimitPeriod = "month"
)

// Start returns the start of the period that t is in. Periods are in UTC like
// the monthly credits of the app itself.
func (p CreditLimitPeriod) Start(t time.Time) time.Time {
	t = t.UTC()
	switch p {
	case CreditLimitPeriodDay:
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	default:
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	}
}

// CreditLimit caps the credits a single server or user can use per period.
//
// Without a TargetID it's the default for every server or user of its scope.
// A limit with a TargetID replaces the default of the same scope and period
// for that server or user, and without Credits it exempts them from it.
type CreditLimit struct {
	ID       string
	AppID    string
	Scope    CreditLimitScope
	TargetID null.String
	Period   CreditLimitPeriod
	Credits  null.Int
	// Message is shown to users who run into the limit instead of the default
	// message of the app. It can contain placeholders.
	Message   null.String
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreditLimitSettings are the settings for all credit limits of an app.
type CreditLimitSettings struct {
	AppID string
	// Message is shown to users who run into a limit without a message of its
	// own, instead of Kite's message. It can contain placeholders.
	Message   null.String
	CreatedAt time.Time
	UpdatedAt time.Time
}

// EffectiveCreditLimits returns the limits that apply to one server or user:
// for each period its own limit if it has one, otherwise the default.
func EffectiveCreditLimits(limits []*CreditLimit, scope CreditLimitScope, targetID string) []*CreditLimit {
	var res []*CreditLimit

	for _, period := range []CreditLimitPeriod{CreditLimitPeriodDay, CreditLimitPeriodMonth} {
		var specific, fallback *CreditLimit
		for _, limit := range limits {
			if limit.Scope != scope || limit.Period != period {
				continue
			}
			if !limit.TargetID.Valid {
				fallback = limit
			} else if limit.TargetID.String == targetID {
				specific = limit
			}
		}

		if specific != nil {
			res = append(res, specific)
		} else if fallback != nil {
			res = append(res, fallback)
		}
	}

	return res
}

type UsageCreditsUsedByTarget struct {
	TargetID    string
	CreditsUsed int
}
