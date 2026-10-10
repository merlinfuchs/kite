package wire

import (
	"regexp"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

// Discord IDs are snowflakes, which have at most 20 digits.
var snowflakeRegex = regexp.MustCompile(`^[0-9]{15,20}$`)

// MaxCreditLimitCredits keeps limits within what the database column holds.
const MaxCreditLimitCredits = 1_000_000_000

type CreditLimit struct {
	ID string `json:"id"`
	// Scope is guild or user.
	Scope string `json:"scope"`
	// TargetID is the server or user ID, or null for the default of the scope.
	TargetID null.String `json:"target_id"`
	// Period is day or month.
	Period string `json:"period"`
	// Credits is null when the target is exempt from the default limit.
	Credits null.Int `json:"credits"`
	// CreditsUsed is how many credits the target used in the current period,
	// null for defaults.
	CreditsUsed null.Int  `json:"credits_used"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreditLimitListResponse = []*CreditLimit

type CreditLimitCreateRequest struct {
	Scope    string      `json:"scope"`
	TargetID null.String `json:"target_id"`
	Period   string      `json:"period"`
	Credits  null.Int    `json:"credits"`
}

func (req CreditLimitCreateRequest) Validate() error {
	return validateCreditLimit(req.Scope, req.TargetID, req.Period, req.Credits)
}

type CreditLimitCreateResponse = CreditLimit

type CreditLimitUpdateRequest = CreditLimitCreateRequest

type CreditLimitUpdateResponse = CreditLimit

type CreditLimitDeleteResponse = Empty

// CreditLimitUsageListResponse lists the servers or users that used the most
// credits in the current period.
type CreditLimitUsageListResponse = []*CreditLimitUsageEntry

type CreditLimitUsageEntry struct {
	TargetID    string `json:"target_id"`
	CreditsUsed int    `json:"credits_used"`
	// Credits is the limit that applies to the target for the period, null if
	// it has none.
	Credits null.Int `json:"credits"`
}

func validateCreditLimit(scope string, targetID null.String, period string, credits null.Int) error {
	return validation.Errors{
		"scope": validation.Validate(scope, validation.Required, validation.In(
			string(model.CreditLimitScopeGuild),
			string(model.CreditLimitScopeUser),
		)),
		"target_id": validation.Validate(targetID, validation.When(targetID.Valid,
			validation.Required,
			validation.Match(snowflakeRegex).Error("must be a valid Discord ID"),
		)),
		"period": validation.Validate(period, validation.Required, validation.In(
			string(model.CreditLimitPeriodDay),
			string(model.CreditLimitPeriodMonth),
		)),
		"credits": validation.Validate(credits,
			validation.When(!targetID.Valid, validation.NotNil.Error("is required for a default limit")),
			validation.When(credits.Valid, validation.Min(int64(0)), validation.Max(int64(MaxCreditLimitCredits))),
		),
	}.Filter()
}

func CreditLimitToWire(limit *model.CreditLimit, creditsUsed null.Int) *CreditLimit {
	if limit == nil {
		return nil
	}

	return &CreditLimit{
		ID:          limit.ID,
		Scope:       string(limit.Scope),
		TargetID:    limit.TargetID,
		Period:      string(limit.Period),
		Credits:     limit.Credits,
		CreditsUsed: creditsUsed,
		CreatedAt:   limit.CreatedAt,
		UpdatedAt:   limit.UpdatedAt,
	}
}
