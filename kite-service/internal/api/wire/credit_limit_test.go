package wire

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
)

func TestCreditLimitCreateRequestValidate(t *testing.T) {
	tests := []struct {
		name  string
		req   CreditLimitCreateRequest
		valid bool
	}{
		{"default", CreditLimitCreateRequest{Scope: "guild", Period: "day", Credits: null.IntFrom(100)}, true},
		{"specific", CreditLimitCreateRequest{Scope: "user", TargetID: null.StringFrom("123456789012345678"), Period: "month", Credits: null.IntFrom(0)}, true},
		{"exempt", CreditLimitCreateRequest{Scope: "user", TargetID: null.StringFrom("123456789012345678"), Period: "month"}, true},
		{"default without credits", CreditLimitCreateRequest{Scope: "guild", Period: "day"}, false},
		{"negative credits", CreditLimitCreateRequest{Scope: "guild", Period: "day", Credits: null.IntFrom(-1)}, false},
		{"too many credits", CreditLimitCreateRequest{Scope: "guild", Period: "day", Credits: null.IntFrom(MaxCreditLimitCredits + 1)}, false},
		{"bad target", CreditLimitCreateRequest{Scope: "guild", TargetID: null.StringFrom("abc"), Period: "day", Credits: null.IntFrom(1)}, false},
		{"empty target", CreditLimitCreateRequest{Scope: "guild", TargetID: null.StringFrom(""), Period: "day", Credits: null.IntFrom(1)}, false},
		{"bad scope", CreditLimitCreateRequest{Scope: "channel", Period: "day", Credits: null.IntFrom(1)}, false},
		{"bad period", CreditLimitCreateRequest{Scope: "guild", Period: "week", Credits: null.IntFrom(1)}, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.req.Validate()
			if test.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
