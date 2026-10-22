package auth

import (
	"strings"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
)

func TestParseSignupSource(t *testing.T) {
	assert.Equal(t, model.SignupSource{
		Referrer:    null.StringFrom("chatgpt.com"),
		UTMSource:   null.StringFrom("chatgpt.com"),
		LandingPage: null.StringFrom("/guides/how-to-make-a-discord-bot"),
	}, parseSignupSource("referrer=chatgpt.com&utm_source=chatgpt.com&utm_medium=&landing_page=%2Fguides%2Fhow-to-make-a-discord-bot"))

	assert.Equal(t, model.SignupSource{}, parseSignupSource(""))
	assert.Equal(t, model.SignupSource{}, parseSignupSource("%zz"))

	long := parseSignupSource("utm_campaign=" + strings.Repeat("a", 500))
	assert.Len(t, long.UTMCampaign.String, maxSignupSourceValueLength)
}
