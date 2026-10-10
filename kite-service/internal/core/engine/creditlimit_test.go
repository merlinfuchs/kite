package engine

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

type creditLimitTestStore struct {
	store.CreditLimitStore
	limits   []*model.CreditLimit
	settings *model.CreditLimitSettings
	err      error
	reads    int
}

func (s *creditLimitTestStore) CreditLimitSettings(ctx context.Context, appID string) (*model.CreditLimitSettings, error) {
	if s.settings == nil {
		return nil, store.ErrNotFound
	}
	return s.settings, nil
}

func (s *creditLimitTestStore) CreditLimitsByApp(ctx context.Context, appID string) ([]*model.CreditLimit, error) {
	s.reads++
	if s.err != nil {
		return nil, s.err
	}
	return s.limits, nil
}

type creditUsageTestStore struct {
	store.UsageStore
	used  map[string]int
	reads int
}

func (s *creditUsageTestStore) UsageCreditsUsedByTargetSince(ctx context.Context, appID string, scope model.CreditLimitScope, targetID string, since time.Time) (int, error) {
	s.reads++
	return s.used[string(scope)+":"+targetID], nil
}

func newTestCreditLimiter(limits []*model.CreditLimit, used map[string]int) (*CreditLimiter, *creditLimitTestStore, *creditUsageTestStore, *time.Time) {
	limitStore := &creditLimitTestStore{limits: limits}
	usageStore := &creditUsageTestStore{used: used}
	limiter := NewCreditLimiter(limitStore, usageStore)

	now := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
	limiter.now = func() time.Time { return now }
	return limiter, limitStore, usageStore, &now
}

func guildLimit(targetID string, period model.CreditLimitPeriod, credits int) *model.CreditLimit {
	limit := &model.CreditLimit{
		ID:      targetID + string(period),
		AppID:   "app",
		Scope:   model.CreditLimitScopeGuild,
		Period:  period,
		Credits: null.IntFrom(int64(credits)),
	}
	if targetID != "" {
		limit.TargetID = null.StringFrom(targetID)
	}
	return limit
}

func TestCreditLimiterNoLimits(t *testing.T) {
	limiter, _, usageStore, _ := newTestCreditLimiter(nil, nil)

	exceeded, err := limiter.Check(context.Background(), "app", creditLimitTargets("1", "2")...)
	require.NoError(t, err)
	assert.Nil(t, exceeded)
	assert.Equal(t, 0, usageStore.reads, "usage shouldn't be read for apps without limits")
}

func TestCreditLimiterDefaultLimit(t *testing.T) {
	limits := []*model.CreditLimit{guildLimit("", model.CreditLimitPeriodDay, 100)}

	limiter, _, _, _ := newTestCreditLimiter(limits, map[string]int{"guild:1": 99, "guild:2": 100})

	exceeded, err := limiter.Check(context.Background(), "app", creditLimitTargets("1", "")...)
	require.NoError(t, err)
	assert.Nil(t, exceeded)

	exceeded, err = limiter.Check(context.Background(), "app", creditLimitTargets("2", "")...)
	require.NoError(t, err)
	require.NotNil(t, exceeded)
	assert.Equal(t, "2", exceeded.TargetID)
	assert.Equal(t, 100, exceeded.Used)
	assert.True(t, exceeded.FirstReport)

	exceeded, err = limiter.Check(context.Background(), "app", creditLimitTargets("2", "")...)
	require.NoError(t, err)
	require.NotNil(t, exceeded)
	assert.False(t, exceeded.FirstReport, "a limit is only reported once")
}

func TestCreditLimiterSpecificLimitReplacesDefault(t *testing.T) {
	limits := []*model.CreditLimit{
		guildLimit("", model.CreditLimitPeriodDay, 100),
		guildLimit("1", model.CreditLimitPeriodDay, 500),
		{
			ID:       "exempt",
			AppID:    "app",
			Scope:    model.CreditLimitScopeGuild,
			TargetID: null.StringFrom("2"),
			Period:   model.CreditLimitPeriodDay,
		},
	}

	limiter, _, _, _ := newTestCreditLimiter(limits, map[string]int{"guild:1": 200, "guild:2": 100000, "guild:3": 200})

	exceeded, err := limiter.Check(context.Background(), "app", creditLimitTargets("1", "")...)
	require.NoError(t, err)
	assert.Nil(t, exceeded, "server 1 has a higher limit of its own")

	exceeded, err = limiter.Check(context.Background(), "app", creditLimitTargets("2", "")...)
	require.NoError(t, err)
	assert.Nil(t, exceeded, "server 2 is exempt")

	exceeded, err = limiter.Check(context.Background(), "app", creditLimitTargets("3", "")...)
	require.NoError(t, err)
	assert.NotNil(t, exceeded, "server 3 falls back to the default")
}

func TestCreditLimiterUserLimit(t *testing.T) {
	limits := []*model.CreditLimit{{
		ID:      "user",
		AppID:   "app",
		Scope:   model.CreditLimitScopeUser,
		Period:  model.CreditLimitPeriodMonth,
		Credits: null.IntFrom(0),
	}}

	limiter, _, _, _ := newTestCreditLimiter(limits, nil)

	exceeded, err := limiter.Check(context.Background(), "app", creditLimitTargets("1", "")...)
	require.NoError(t, err)
	assert.Nil(t, exceeded, "executions without a user aren't limited per user")

	exceeded, err = limiter.Check(context.Background(), "app", creditLimitTargets("1", "5")...)
	require.NoError(t, err)
	require.NotNil(t, exceeded, "a limit of 0 blocks every execution")
	assert.Equal(t, model.CreditLimitScopeUser, exceeded.Limit.Scope)
}

func TestCreditLimiterRecordCountsCachedUsage(t *testing.T) {
	limits := []*model.CreditLimit{guildLimit("", model.CreditLimitPeriodMonth, 10)}

	limiter, _, usageStore, _ := newTestCreditLimiter(limits, map[string]int{"guild:1": 5})

	exceeded, err := limiter.Check(context.Background(), "app", creditLimitTargets("1", "")...)
	require.NoError(t, err)
	assert.Nil(t, exceeded)

	limiter.Record("app", 5, creditLimitTargets("1", "")...)

	exceeded, err = limiter.Check(context.Background(), "app", creditLimitTargets("1", "")...)
	require.NoError(t, err)
	assert.NotNil(t, exceeded, "credits recorded in this process count before the next read")
	assert.Equal(t, 1, usageStore.reads)
}

func TestCreditLimiterNewPeriodResetsUsage(t *testing.T) {
	limits := []*model.CreditLimit{guildLimit("", model.CreditLimitPeriodDay, 10)}

	limiter, _, usageStore, now := newTestCreditLimiter(limits, map[string]int{"guild:1": 10})

	exceeded, err := limiter.Check(context.Background(), "app", creditLimitTargets("1", "")...)
	require.NoError(t, err)
	assert.NotNil(t, exceeded)

	usageStore.used["guild:1"] = 0
	*now = now.Add(13 * time.Hour)

	exceeded, err = limiter.Check(context.Background(), "app", creditLimitTargets("1", "")...)
	require.NoError(t, err)
	assert.Nil(t, exceeded)
}

func TestCreditLimiterCachesLimits(t *testing.T) {
	limiter, limitStore, _, now := newTestCreditLimiter(nil, nil)

	for range 3 {
		_, err := limiter.Check(context.Background(), "app", creditLimitTargets("1", "2")...)
		require.NoError(t, err)
	}
	assert.Equal(t, 1, limitStore.reads)

	*now = now.Add(creditLimitCacheTTL)
	_, err := limiter.Check(context.Background(), "app", creditLimitTargets("1", "2")...)
	require.NoError(t, err)
	assert.Equal(t, 2, limitStore.reads)
}

func TestCreditLimiterKeepsLimitsWhenReloadFails(t *testing.T) {
	limits := []*model.CreditLimit{guildLimit("", model.CreditLimitPeriodDay, 1)}

	limiter, limitStore, _, now := newTestCreditLimiter(limits, map[string]int{"guild:1": 1})

	exceeded, err := limiter.Check(context.Background(), "app", creditLimitTargets("1", "")...)
	require.NoError(t, err)
	assert.NotNil(t, exceeded)

	limitStore.err = errors.New("database down")
	*now = now.Add(creditLimitCacheTTL)

	exceeded, err = limiter.Check(context.Background(), "app", creditLimitTargets("1", "")...)
	require.NoError(t, err)
	assert.NotNil(t, exceeded)
}

func TestCreditLimitPeriodStart(t *testing.T) {
	at := time.Date(2026, 10, 10, 23, 59, 0, 0, time.UTC)
	assert.Equal(t, time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC), model.CreditLimitPeriodDay.Start(at))
	assert.Equal(t, time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), model.CreditLimitPeriodMonth.Start(at))
}

func TestCreditLimiterMessage(t *testing.T) {
	custom := guildLimit("1", model.CreditLimitPeriodDay, 1)
	custom.Message = null.StringFrom("Server {{limit.scope}} is out")

	limits := []*model.CreditLimit{guildLimit("", model.CreditLimitPeriodDay, 1), custom}

	limiter, limitStore, _, _ := newTestCreditLimiter(limits, map[string]int{"guild:1": 1, "guild:2": 1})

	exceeded, err := limiter.Check(context.Background(), "app", creditLimitTargets("2", "")...)
	require.NoError(t, err)
	require.NotNil(t, exceeded)
	assert.False(t, exceeded.Message.Valid, "without any custom message Kite's message is used")

	exceeded, err = limiter.Check(context.Background(), "app", creditLimitTargets("1", "")...)
	require.NoError(t, err)
	require.NotNil(t, exceeded)
	assert.Equal(t, "Server {{limit.scope}} is out", exceeded.Message.String, "the message of the limit wins")

	limitStore.settings = &model.CreditLimitSettings{AppID: "app", Message: null.StringFrom("App default")}
	limiter.limits = make(map[string]cachedCreditLimits)

	exceeded, err = limiter.Check(context.Background(), "app", creditLimitTargets("2", "")...)
	require.NoError(t, err)
	require.NotNil(t, exceeded)
	assert.Equal(t, "App default", exceeded.Message.String, "limits without a message use the app default")

	exceeded, err = limiter.Check(context.Background(), "app", creditLimitTargets("1", "")...)
	require.NoError(t, err)
	require.NotNil(t, exceeded)
	assert.Equal(t, "Server {{limit.scope}} is out", exceeded.Message.String)
}

func TestCreditLimitEnv(t *testing.T) {
	limit := guildLimit("", model.CreditLimitPeriodDay, 50)
	now := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)

	env := creditLimitEnv(&CreditLimitExceeded{Limit: limit, Used: 51}, now)
	assert.Equal(t, int64(50), env["credits"])
	assert.Equal(t, 51, env["used"])
	assert.Equal(t, "server", env["scope"])
	assert.Equal(t, "today", env["period"])
	assert.Equal(t, "<t:1791676800:R>", env["resets"])

	limit.Period = model.CreditLimitPeriodMonth
	env = creditLimitEnv(&CreditLimitExceeded{Limit: limit}, now)
	assert.Equal(t, "this month", env["period"])
	assert.Equal(t, fmt.Sprintf("<t:%d:R>", time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC).Unix()), env["resets"])
}

func TestRenderCreditLimitMessage(t *testing.T) {
	render := func(message null.String) (string, error) {
		fCtx := flow.NewContext(
			context.Background(),
			time.Second,
			&EventData{},
			flow.FlowProviders{},
			flow.FlowContextLimits{},
			eval.NewContext(eval.Env{"user": map[string]any{"mention": "<@1>"}}),
			nil,
		)
		defer fCtx.Cancel()

		return renderCreditLimitMessage(fCtx, &CreditLimitExceeded{
			Limit:   guildLimit("", model.CreditLimitPeriodDay, 50),
			Used:    50,
			Message: message,
		})
	}

	content, err := render(null.String{})
	require.NoError(t, err)
	assert.Equal(t, "This server has reached its usage limit for today. Try again later.", content)

	content, err = render(null.StringFrom("{{user.mention}} used {{limit.used}}/{{limit.credits}} credits {{limit.period}}"))
	require.NoError(t, err)
	assert.Equal(t, "<@1> used 50/50 credits today", content)

	content, err = render(null.StringFrom("   "))
	require.NoError(t, err)
	assert.Equal(t, "This server has reached its usage limit for today. Try again later.", content)

	content, err = render(null.StringFrom(strings.Repeat("{{limit.period}}", 500)))
	require.NoError(t, err)
	assert.Len(t, []rune(content), maxCreditLimitMessageLength)
}
