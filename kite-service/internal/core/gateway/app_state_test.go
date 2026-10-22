package gateway

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/diamondburned/arikawa/v3/utils/httputil"
	"github.com/stretchr/testify/assert"
)

func TestIsNotInGuild(t *testing.T) {
	assert.True(t, isNotInGuild(&httputil.HTTPError{Status: http.StatusForbidden}))
	assert.True(t, isNotInGuild(fmt.Errorf("wrapped: %w", &httputil.HTTPError{Status: http.StatusNotFound})))

	assert.False(t, isNotInGuild(nil))
	assert.False(t, isNotInGuild(&httputil.HTTPError{Status: http.StatusInternalServerError}))
	assert.False(t, isNotInGuild(errors.New("timeout")))
}
